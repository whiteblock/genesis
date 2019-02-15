package deploy

import(
    "context"
    "fmt"
    "golang.org/x/sync/semaphore"
    "errors"
    "log"
    db "../db"
    util "../util"
    state "../state"
)
/**
 * @brief Add nodes to the network instead of building independently
 * @details [long description]
 * 
 * @param Config The configuration to build
 * @param string [description]
 * @param r [description]
 * @return [description]
 */


func AddNodes(buildConf *Config,servers []db.Server,resources []util.Resources,clients []*util.SshClient,
              buildState *state.BuildState) (map[int][]string,error) {
    
    buildState.SetDeploySteps(2*buildConf.Nodes )
    defer buildState.FinishDeploy()

    var sem = semaphore.NewWeighted(conf.ThreadLimit)
    ctx := context.TODO()
    
    fmt.Println("-------------Building The Docker Containers-------------")

    buildState.SetBuildStage("Provisioning the nodes")

    availibleServers := make([]int,len(servers))
    for i,_ := range availibleServers {
        availibleServers[i] = i
    }
    out := map[int][]string{}
    index := 0
    for i := 0; i < buildConf.Nodes; i++ {
        serverIndex := availibleServers[index]
        nodeNum := servers[serverIndex].Nodes + i
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
        out[servers[serverIndex].Id] = append(out[servers[serverIndex].Id],util.GetNodeIP(servers[serverIndex].ServerID,nodeNum))

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
        }(serverIndex,nodeNum)
        

        index++
        index = index % len(availibleServers)
    }

    err := sem.Acquire(ctx,conf.ThreadLimit)
    if err != nil {
        log.Println(err)
        return nil, err
    }
    sem.Release(conf.ThreadLimit)
  

    buildState.SetBuildStage("Setting up services")

    sem.Acquire(ctx,1)
    go func(){
        defer sem.Release(1)
        err = finalizeNewNodes(servers,clients,out,buildState)
        if err != nil {
            log.Println(err)
            buildState.ReportError(err)
            return
        }
    }()

    for i,_ := range servers {
        clients[i].Run("sudo iptables --flush DOCKER-ISOLATION-STAGE-1")
    }
    
    //Acquire all of the resources here, then release and destroy
    err = sem.Acquire(ctx,conf.ThreadLimit)
    if err != nil {
        log.Println(err)
        return nil, nil
    }
    sem.Release(conf.ThreadLimit)
    
    log.Println("Finished adding nodes into the network")
    return out, buildState.GetError()
}


func AddNodesLegacy(buildConf *Config, servers []db.Server,resources util.Resources,
              clients []*util.SshClient,buildState *state.BuildState) (map[int][]string,error){
    
    buildState.SetDeploySteps(3*buildConf.Nodes + 2 )
    defer buildState.FinishDeploy()

    var sem = semaphore.NewWeighted(conf.ThreadLimit)
    
    ctx := context.TODO()

    fmt.Println("-------------Building The Docker Containers-------------")
    n := buildConf.Nodes
    i := 0
    out := map[int][]string{}
    for n > 0 && i < len(servers){
        out[servers[i].Id] = []string{}
        fmt.Printf("-------------Building on Server %d-------------\n",i)
        
        max_nodes := int(servers[i].Max - servers[i].Nodes)
        var nodes int
        if max_nodes > n {
            nodes = n
        }else{
            nodes = max_nodes
        }
        for j := 0; j < nodes; j++ {
            out[servers[i].Id] = append(out[servers[i].Id],util.GetNodeIP(servers[i].ServerID,j))
        }
        
        buildState.SetBuildStage("Provisioning Nodes")
       
        err := DockerNetworkCreateAppendAll(servers[i],clients[i],servers[i].Nodes,nodes,buildState)
        if err != nil {
            log.Println(err)
            return nil,err
        }
        
        fmt.Printf("Creating the docker containers on server %d\n",i)
        buildState.SetBuildStage("Configuring Network")
        err = DockerRunAppendAll(servers[i],clients[i],resources,servers[i].Nodes,nodes,buildConf.Image,buildState)
        if err != nil{
            log.Println(err)
            return nil,err
        }
        n -= nodes
        i++

    }
    //Acquire all of the resources here, then release and destroy
    err := sem.Acquire(ctx,conf.ThreadLimit)
    if err != nil {
        log.Println(err)
        return out, nil
    }
    sem.Release(conf.ThreadLimit)
    if n != 0 {
        return out, errors.New(fmt.Sprintf("ERROR: Only able to build %d/%d nodes\n",(buildConf.Nodes - n),buildConf.Nodes))
    }
    return out, err
}