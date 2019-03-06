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
func AddNodes(buildConf *db.DeploymentDetails,servers []db.Server,clients []*util.SshClient,
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

            resource := buildConf.Resources[0]
            if len(buildConf.Resources) > i {
                resource = buildConf.Resources[i]
            }

            var env map[string]string = nil
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