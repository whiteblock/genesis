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


func Build(buildConf *db.DeploymentDetails,servers []db.Server,clients []*util.SshClient,
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

            resource := buildConf.Resources[0]
            var env map[string]string = nil

            if len(buildConf.Resources) > i {
                resource = buildConf.Resources[i]
            }
            if buildConf.Environments != nil && len(buildConf.Environments) > i && buildConf.Environments[i] != nil {
                env = buildConf.Environments[i]
            }

            err = DockerRun(servers[serverIndex],clients[serverIndex],resource,i,buildConf.Image,env)
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
