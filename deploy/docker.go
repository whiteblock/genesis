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


func DockerKill(client *util.SshClient,node int) error {
    _,err := client.Run(fmt.Sprintf("docker rm -f %s%d",conf.NodePrefix,node))
    return err
}

func DockerKillAll(client *util.SshClient) error {
    _,err := client.Run(fmt.Sprintf("docker rm -f $(docker ps -aq -f name=\"%s\")",conf.NodePrefix));
    state.IncrementDeployProgress()
    return err
}

func dockerNetworkCreateCmd(subnet string,gateway string,iface string,vlan int,network int,name string) string {
    if !conf.NeoBuild {
        return fmt.Sprintf("docker network create -d macvlan --subnet %s --gateway %s -o parent=%s.%d %s",
                            subnet,
                            gateway,
                            iface,
                            vlan,
                            name)
    }
    return fmt.Sprintf("docker network create --subnet %s --gateway %s -o \"com.docker.network.bridge.name=%s%d\" %s",
                            subnet,
                            gateway,
                            conf.BridgePrefix,
                            network,
                            name)

    
}

func DockerNetworkCreate(server db.Server,client *util.SshClient,node int) error {
    command := dockerNetworkCreateCmd(
                    util.GetNetworkAddress(server.ServerID,node),
                    util.GetGateway(server.ServerID,node),
                    server.Iface,
                    node+conf.NetworkVlanStart,
                    node,
                    fmt.Sprintf("%s%d",conf.NodeNetworkPrefix,node))
    
    res,err := client.Run(command)
    if err != nil{
        res,err = client.Run(command)
        if err != nil{
            log.Println(err)
            return errors.New(res)
        }
        
    }
    return nil
}

func DockerNetworkCreateAll(server db.Server,client *util.SshClient,nodes int) error {
    for i := 0; i < nodes; i++{
        state.IncrementDeployProgress()
        err := DockerNetworkCreate(server,client,i)
        if err != nil {
            log.Println(err)
            return err
        }
    }
    return nil
}

func DockerNetworkCreateAppendAll(server db.Server,client *util.SshClient,start int,nodes int) error {
    for i := start; i < start+nodes; i++ {
        state.IncrementDeployProgress()
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

func DockerNetworkDestroyAll(client *util.SshClient) error {
    _,err := client.Run(fmt.Sprintf("for net in $(docker network ls | grep %s | awk '{print $1}'); do docker network rm $net; done",conf.NodeNetworkPrefix))
    state.IncrementDeployProgress()
    return err
}

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

func dockerRunCmd(server db.Server,resources util.Resources,node int,image string) (string,error) {
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

    command += fmt.Sprintf(" --ip %s",util.GetNodeIP(server.ServerID,node))
    command += fmt.Sprintf(" --hostname %s%d",conf.NodePrefix,node)
    command += fmt.Sprintf(" --name %s%d",conf.NodePrefix,node)
    command += " " + image
    return command,nil
}

func DockerRun(server db.Server,client *util.SshClient,resources util.Resources,node int,image string) error {
    command,err := dockerRunCmd(server,resources,node,image)
    if err != nil{
        return err
    }
    res,err := client.Run(command)
    if err != nil{
        log.Println(err)
        log.Println(res)
    }
    return err
}

func DockerRunAll(server db.Server,client *util.SshClient,resources util.Resources,nodes int,image string) error {
    return DockerRunAppendAll(server,client,resources,0,nodes,image)
}

func DockerRunAppendAll(server db.Server,client *util.SshClient,resources util.Resources,start int,nodes int,image string) error {
    var command string
    for i := start; i < start+nodes; i++ {
        //state.IncrementDeployProgress()
        tmp,err := dockerRunCmd(server,resources,i,image)
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

/**
 * @brief Start a service container
 * @details [long description]
 * 
 * @param string [description]
 * @param string [description]
 * @param string [description]
 * @param mapstring [description]
 * @param string [description]
 * @return [description]
 */
func serviceDockerRunCmd(network string,ip string,name string,env map[string]string,image string) string {
    envFlags := ""
    for k,v := range env{
        envFlags += fmt.Sprintf("-e \"%s=%s\" ",k,v)
    }
    envFlags += fmt.Sprintf("-e \"BIND_ADDR=%s\"",ip)
    return fmt.Sprintf("docker run -itd --network %s --ip %s --hostname %s --name %s %s %s",
                        network,
                        ip,
                        name,
                        name,
                        envFlags,
                        image)
}

func DockerStopServices(client *util.SshClient) error {
    res,err := client.Run(fmt.Sprintf("docker rm -f $(docker ps -aq -f name=%s)",conf.ServicePrefix));
    client.Run("docker network rm "+conf.ServiceNetworkName)
    if err != nil {
        log.Println(res);
    }
    return err
}

func DockerStartServices(server db.Server,client *util.SshClient,services []util.Service) error {
    gateway,subnet,err := util.GetServiceNetwork()
    if err != nil {
        log.Println(err)
        return err
    }

    res,err := client.KeepTryRun(dockerNetworkCreateCmd(subnet,gateway,server.Iface,conf.ServiceVlan,-1,conf.ServiceNetworkName))
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
        res,err := client.KeepTryRun(serviceDockerRunCmd(conf.ServiceNetworkName,
                                               ips[service.Name],
                                               fmt.Sprintf("%s%d",conf.ServicePrefix,i),
                                               service.Env,
                                               service.Image))
        if err != nil {
            log.Println(err)
            log.Println(res)
            return err
        }
        state.IncrementDeployProgress()
    }
    return nil
}