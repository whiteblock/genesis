package deploy

import (
    "context"
    "fmt"
    "golang.org/x/sync/semaphore"
    "errors"
    "log"
    db "../db"
    util "../util"
    state "../state"
    netem "../net"
)

var conf *util.Config = util.GetConfig()





func Build(buildConf *Config,servers []db.Server,resources []util.Resources,clients []*util.SshClient,
           services []util.Service,buildState *state.BuildState) ([]db.Server,error) {
    
    buildState.SetDeploySteps(3*buildConf.Nodes + 2 + len(services) )
    defer buildState.FinishDeploy()

    var sem = semaphore.NewWeighted(conf.ThreadLimit)
    ctx := context.TODO()
    
    fmt.Println("-------------Building The Docker Containers-------------")

    buildState.SetBuildStage("Tearing down the previous testnet")
    for i,_ := range servers {
        sem.Acquire(ctx,1)
        go func(i int){
            defer sem.Release(1)
            DockerKillAll(clients[i],buildState)
            DockerNetworkDestroyAll(clients[i],buildState)
        }(i)
    }

    err := sem.Acquire(ctx,conf.ThreadLimit)
    if err != nil {
        log.Println(err)
        return nil, err
    }
    sem.Release(conf.ThreadLimit)
    
    buildState.SetBuildStage("Provisioning the nodes")

    availibleServers := make([]int,len(servers))
    for i,_ := range availibleServers {
        availibleServers[i] = i
    }
    
    index := 0
    for i := 0; i < buildConf.Nodes; i++ {
        serverIndex := availibleServers[index]
        if servers[serverIndex].Max <= servers[serverIndex].Nodes {
            if len(availibleServers) == 1 {
                return nil,errors.New("Cannot build that many nodes with the availible resources")
            }
            availibleServers = append(availibleServers[:serverIndex],availibleServers[serverIndex+1:]...) 
            i--
            index++
            index = index % len(availibleServers)
            continue
        }

        servers[serverIndex].Ips = append(servers[serverIndex].Ips,util.GetNodeIP(servers[serverIndex].ServerID,i))
        servers[serverIndex].Nodes++;

        sem.Acquire(ctx,1)
        go func(serverIndex int,i int){
            defer sem.Release(1)
            err := DockerNetworkCreate(servers[serverIndex],clients[serverIndex],i)
            if err != nil {
                log.Println(err)
                buildState.ReportError(err)
                return
            }
            buildState.IncrementDeployProgress()

            resource := resources[0]
            if len(resources) > i {
                resource = resources[i]
            }

            err = DockerRun(servers[serverIndex],clients[serverIndex],resource,i,buildConf.Image)
            if err != nil {
                log.Println(err)
                buildState.ReportError(err)
                return
            }
            buildState.IncrementDeployProgress()
        }(serverIndex,i)
        

        index++
        index = index % len(availibleServers)
    }

    err = sem.Acquire(ctx,conf.ThreadLimit)
    if err != nil {
        log.Println(err)
        return nil, err
    }
    sem.Release(conf.ThreadLimit)
  

    buildState.SetBuildStage("Setting up services")

    sem.Acquire(ctx,1)
    go func(){
        defer sem.Release(1)
        err = finalize(servers,clients,buildState)
        if err != nil {
            log.Println(err)
            buildState.ReportError(err)
            return
        }
    }()
    
    for i,_ := range servers {
        sem.Acquire(ctx,1)
        go func(i int){
            defer sem.Release(1)
            netem.RemoveAll(clients[i],servers[i].Nodes)
        }(i)
    }


    for i,_ := range servers {
        DockerStopServices(clients[i])
    }
    
    if services != nil {//Maybe distribute the services over multiple servers
        err := DockerStartServices(servers[0],clients[0],services,buildState)
        if err != nil{
            log.Println(err)
            return nil,err
        }
    }

    for i,_ := range servers {
        clients[i].Run("sudo iptables --flush DOCKER-ISOLATION-STAGE-1")
    }
    
    //Acquire all of the resources here, then release and destroy
    err = sem.Acquire(ctx,conf.ThreadLimit)
    if err != nil {
        log.Println(err)
        return nil, err
    }
    sem.Release(conf.ThreadLimit)
  
    return servers, buildState.GetError()
}


/**
 * Builds out the Docker Network on pre-setup servers
 * Returns a string of all of the IP addresses 
 */
func Build_legacy(buildConf *Config,servers []db.Server,resources util.Resources,clients []*util.SshClient,
           services []util.Service,buildState *state.BuildState) ([]db.Server,error) {
    buildState.SetDeploySteps(3*buildConf.Nodes + 2 + len(services) )
    defer buildState.FinishDeploy()

    var sem = semaphore.NewWeighted(conf.ThreadLimit)
    
    ctx := context.TODO()
    if !conf.NeoBuild {
        Prepare(buildConf.Nodes,servers)
    }
    
    fmt.Println("-------------Building The Docker Containers-------------")
    n := buildConf.Nodes
    i := 0

    for n > 0 && i < len(servers){
        fmt.Printf("-------------Building on Server %d-------------\n",i)
        
        max_nodes := int(servers[i].Max - servers[i].Nodes)
        var nodes int
        if max_nodes > n {
            nodes = n
        }else{
            nodes = max_nodes
        }
        for j := 0; j < nodes; j++ {
            servers[i].Ips = append(servers[i].Ips,util.GetNodeIP(servers[i].ServerID,j))
        }
        //Kill all the nodes on the server
        DockerKillAll(clients[i],buildState)
        buildState.SetBuildStage("Provisioning Nodes")
        //Destroy all the networks on the server
        DockerNetworkDestroyAll(clients[i],buildState)
        err := DockerNetworkCreateAll(servers[i],clients[i],nodes,buildState)
        if err != nil {
            log.Println(err)
            return nil,err
        }
        
        fmt.Printf("Creating the docker containers on server %d\n",i)
        buildState.SetBuildStage("Configuring Network")
        err = DockerRunAll(servers[i],clients[i],resources,nodes,buildConf.Image,buildState)
        if err != nil{
            log.Println(err)
            return nil,err
        }
        if conf.NeoBuild {
            clients[i].Run("sudo iptables --flush DOCKER-ISOLATION-STAGE-1")
            netem.RemoveAll(clients[i],nodes)
        }        
        n -= nodes
        i++

    }
    //Acquire all of the resources here, then release and destroy
    err := sem.Acquire(ctx,conf.ThreadLimit)
    if err != nil {
        log.Println(err)
        return nil, err
    }
    sem.Release(conf.ThreadLimit)
    if n != 0 {
        return servers, errors.New(fmt.Sprintf("ERROR: Only able to build %d/%d nodes\n",(buildConf.Nodes - n),buildConf.Nodes))
    }
    buildState.SetBuildStage("Setting up services")
    DockerStopServices(clients[0])
    if services != nil {
        err = DockerStartServices(servers[0],clients[0],services,buildState)
        if err != nil{
            log.Println(err)
            return nil,err
        }
    }
    err = finalize(servers,clients,buildState)
    return servers, err
}