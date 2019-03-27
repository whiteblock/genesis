package deploy

import(
    "errors"
    "fmt"
    "log"
    util "../util"
    db "../db"
    state "../state"
)

/**Quick naive interface to Docker calls over ssh*/

/*
    Kill a single node by index on a server
 */
func DockerKill(client *util.SshClient,node int) error {
    _,err := client.Run(fmt.Sprintf("docker rm -f %s%d",conf.NodePrefix,node))
    return err
}

/*
    Kill all nodes on a server
 */
func DockerKillAll(client *util.SshClient) error {
    _,err := client.Run(fmt.Sprintf("docker rm -f $(docker ps -aq -f name=\"%s\")",conf.NodePrefix));
    return err
}

/*
    Create the command to a docker network for a node
 */
func dockerNetworkCreateCmd(subnet string,gateway string,network int,name string) string {
    return fmt.Sprintf("docker network create --subnet %s --gateway %s -o \"com.docker.network.bridge.name=%s%d\" %s",
                            subnet,
                            gateway,
                            conf.BridgePrefix,
                            network,
                            name)
}

/*
    Create a docker network for a node
 */
func DockerNetworkCreate(server db.Server,client *util.SshClient,node int) error {
    command := dockerNetworkCreateCmd(
                    util.GetNetworkAddress(server.SubnetID,node),
                    util.GetGateway(server.SubnetID,node),
                    node,
                    fmt.Sprintf("%s%d",conf.NodeNetworkPrefix,node))
    
    res,err := client.Run(command)
    if err != nil{
        res,err = client.Run(command)
        if err != nil{
            log.Println(err)
            return fmt.Errorf(res)
        }
    }
    return nil
}

/*
    Create all of the node docker networks on a server
 */
func DockerNetworkCreateAll(server db.Server,client *util.SshClient,nodes int,buildState *state.BuildState) error {
    for i := 0; i < nodes; i++{
        buildState.IncrementDeployProgress()
        err := DockerNetworkCreate(server,client,i)
        if err != nil {
            log.Println(err)
            return err
        }
    }
    return nil
}

func DockerNetworkCreateAppendAll(server db.Server,client *util.SshClient,start int,
                                  nodes int,buildState *state.BuildState) error {
    for i := start; i < start+nodes; i++ {
        buildState.IncrementDeployProgress()
        err := DockerNetworkCreate(server,client,i)
        if err != nil {
            log.Println(err)
            return err
        }
    }
    return nil
}

func DockerNetworkDestroy(client *util.SshClient, node int ) error {
    res,err := client.Run(fmt.Sprintf("docker network rm %s%d",conf.NodeNetworkPrefix,node))
    if err != nil {
        log.Println(err)
        log.Println(res)
        return err
    }
    return nil
}

/*
    Remove all whiteblock networks on a node
 */
func DockerNetworkDestroyAll(client *util.SshClient,buildState *state.BuildState) error {
    _,err := client.Run(fmt.Sprintf("for net in $(docker network ls | grep %s | awk '{print $1}'); do docker network rm $net; done",conf.NodeNetworkPrefix))
    buildState.IncrementDeployProgress()
    return err
}

/*
    Pull an image on all the given servers
 */
func DockerPull(clients []*util.SshClient,image string) error {
    for _,client := range clients {
        res,err := client.Run("docker pull " + image)
        if err != nil {
            log.Println(err)
            log.Println(res)
            return err
        }
    }
    return nil
}

/*
    Makes a docker run command to start a node
 */
func dockerRunCmd(server db.Server,resources util.Resources,node int,image string,env map[string]string) (string,error) {
    command := "docker run -itd --entrypoint /bin/sh "
    command += fmt.Sprintf("--network %s%d",conf.NodeNetworkPrefix,node)

    if !resources.NoCpuLimits() {
        command += fmt.Sprintf(" --cpus %s",resources.Cpus)
    }

    if !resources.NoMemoryLimits() {
        mem,err := resources.GetMemory()
        if err != nil {
            return "",errors.New("Invalid value for memory")
        }
        command += fmt.Sprintf(" --memory %d",mem)
    }
    for key,value := range env {
        command += fmt.Sprintf(" -e \"%s=%s\"",key,value)
    }
    command += fmt.Sprintf(" --ip %s",util.GetNodeIP(server.SubnetID,node))
    command += fmt.Sprintf(" --hostname %s%d",conf.NodePrefix,node)
    command += fmt.Sprintf(" --name %s%d",conf.NodePrefix,node)
    command += " " + image
    return command,nil
}

/*
    Starts a node 
 */
func DockerRun(server db.Server,client *util.SshClient,resources util.Resources,node int,image string,env map[string]string) error {
    command,err := dockerRunCmd(server,resources,node,image,env)
    if err != nil{
        log.Println(err)
        return err
    }
    res,err := client.Run(command)
    if err != nil{
        log.Println(err)
        log.Println(res)
        return errors.New(res)
    }
    return nil
}

/*
    Start a batch of nodes
 */
func DockerRunAll(server db.Server,client *util.SshClient,resources []util.Resources,nodes int,
                  image string,buildState *state.BuildState,envs []map[string]string) error {
    return DockerRunAppendAll(server,client,resources,0,nodes,image,buildState,envs)
}

/*
    Similar to docker run all, but start creating the nodes at a given starting point, 
    rather than 0
 */
func DockerRunAppendAll(server db.Server,client *util.SshClient,resources []util.Resources,start int,
                        nodes int,image string,buildState *state.BuildState,envs []map[string]string) error {
    var command string
    for i := start; i < start+nodes; i++ {
        //state.IncrementDeployProgress()
        resource := resources[0]
        var env map[string]string = nil

        if len(resources) > i {
            resource = resources[i]
        }
        if envs != nil && len(envs) > i && envs[i] != nil {
            env = envs[i]
        }
        tmp,err := dockerRunCmd(server,resource,i,image,env)
        if err != nil{
            return err
        }

        if len(command) == 0 {
            command += tmp
        }else{
            command += "&&" + tmp
        }

        if i % 2 == 0 || i == (start+nodes) - 1 {
            res,err := client.Run(command)
            command = ""
            if err != nil {
                log.Println(err)
                log.Println(res)
                return err
            }
        }
        
    }
    return nil
}

/*
    Creates the command to start a service container
 */
func serviceDockerRunCmd(network string,ip string,name string,env map[string]string,image string) string {
    envFlags := ""
    for k,v := range env{
        envFlags += fmt.Sprintf("-e \"%s=%s\" ",k,v)
    }
    envFlags += fmt.Sprintf("-e \"BIND_ADDR=%s\"",ip)
    ipFlag := ""
    if len(ip) > 0 {
        ipFlag = fmt.Sprintf("--ip %s",ip)
    }
    return fmt.Sprintf("docker run -itd --network %s %s --hostname %s --name %s %s %s",
                        network,
                        ipFlag,
                        name,
                        name,
                        envFlags,
                        image)
}

/*
    Stop all services and remove the service network from a server
 */
func DockerStopServices(client *util.SshClient) error {
    res,err := client.Run(fmt.Sprintf("docker rm -f $(docker ps -aq -f name=%s)",conf.ServicePrefix));
    client.Run("docker network rm "+conf.ServiceNetworkName)
    if err != nil {
        log.Println(res);
    }
    return err
}

/*
    Creates the service network and starts all the services on a server
 */
func DockerStartServices(server db.Server,client *util.SshClient,services []util.Service,buildState *state.BuildState) error {
    gateway,subnet,err := util.GetServiceNetwork()
    if err != nil {
        log.Println(err)
        return err
    }

    res,err := client.KeepTryRun(dockerNetworkCreateCmd(subnet,gateway,-1,conf.ServiceNetworkName))
    if err != nil{
        log.Println(err)
        log.Println(res)
        return err
    }
    ips,err := util.GetServiceIps(services)
    if err != nil{
        log.Println(err)
        return err
    }

    for i,service := range services {
        net := conf.ServiceNetworkName
        ip := ips[service.Name]
        if len(service.Network) != 0 {
            net = service.Network
            ip = ""
        }
        res,err := client.KeepTryRun(serviceDockerRunCmd(net,ip,
                                               fmt.Sprintf("%s%d",conf.ServicePrefix,i),
                                               service.Env,
                                               service.Image))
        if err != nil {
            log.Println(err)
            log.Println(res)
            return err
        }
        buildState.IncrementDeployProgress()
    }
    return nil
}