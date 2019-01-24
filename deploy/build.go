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
/**
 * Builds out the Docker Network on pre-setup servers
 * Returns a string of all of the IP addresses 
 */
func Build(buildConf *Config,servers []db.Server,resources Resources,clients []*util.SshClient,services []util.Service) ([]db.Server,error) {
    state.SetDeploySteps(3*buildConf.Nodes + 2 + len(services) )
    defer state.FinishDeploy()

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
        DockerKillAll(clients[i])
        state.SetBuildStage("Provisioning Nodes")
        //Destroy all the networks on the server
        DockerNetworkDestroyAll(clients[i])
        err := DockerNetworkCreateAll(servers[i],clients[i],nodes)
        if err != nil {
            log.Println(err)
            return nil,err
        }
        
        fmt.Printf("Creating the docker containers on server %d\n",i)
        state.SetBuildStage("Configuring Network")
        err = DockerRunAll(servers[i],clients[i],resources,nodes,buildConf.Image)
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
        return servers, nil
    }
    sem.Release(conf.ThreadLimit)
    if n != 0 {
        return servers, errors.New(fmt.Sprintf("ERROR: Only able to build %d/%d nodes\n",(buildConf.Nodes - n),buildConf.Nodes))
    }
    state.SetBuildStage("Setting up services")
    DockerStopServices(clients[0])
    if services != nil {
        err = DockerStartServices(servers[0],clients[0],services)
        if err != nil{
            log.Println(err)
            return nil,err
        }
    }
    err = finalize(servers,clients)
    return servers, err
}