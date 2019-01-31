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
    netem "../net"
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
func AddNodes(buildConf *Config, servers []db.Server,resources util.Resources,clients []*util.SshClient) (map[int][]string,error){
    state.SetDeploySteps(3*buildConf.Nodes + 2 )
    defer state.FinishDeploy()

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
        
        state.SetBuildStage("Provisioning Nodes")
       
        err := DockerNetworkCreateAppendAll(servers[i],clients[i],servers[i].Nodes,nodes)
        if err != nil {
            log.Println(err)
            return nil,err
        }
        
        fmt.Printf("Creating the docker containers on server %d\n",i)
        state.SetBuildStage("Configuring Network")
        err = DockerRunAppendAll(servers[i],clients[i],resources,servers[i].Nodes,nodes,buildConf.Image)
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
        return out, nil
    }
    sem.Release(conf.ThreadLimit)
    if n != 0 {
        return out, errors.New(fmt.Sprintf("ERROR: Only able to build %d/%d nodes\n",(buildConf.Nodes - n),buildConf.Nodes))
    }
    return out, err
}