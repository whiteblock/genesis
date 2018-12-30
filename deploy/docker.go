package deploy

import(
    "errors"
    "fmt"
    util "../util"
    db "../db"
)

/**Quick naive interface to Docker calls over ssh*/



func DockerKill(client *util.SshClient,node int) error {
    _,err := client.Run(fmt.Sprintf("docker rm -f whiteblock-node%d",node))
    return err
}


func DockerKillAll(client *util.SshClient) error {
    _,err := client.Run("docker rm -f $(docker ps -aq -f name=whiteblock)");
    return err
}

func DockerNetworkCreate(server db.Server,client *util.SshClient,node int) error {
    command := "docker network create -d macvlan"
    command += fmt.Sprintf(" --subnet %d",util.GetSubnet())
    command += fmt.Sprintf(" --gateway %s",util.GetGateway(server.ServerID,node))
    command += fmt.Sprintf(" -o parent=%s.%d wb_vlan_%d",server.Iface,node+100,node)
    _,err := client.Run(command)
    return err
}

func DockerNetworkCreateAll(server db.Server,client *util.SshClient,nodes int) error {
    for i := 0; i < nodes; i++{
        err := DockerNetworkCreate(server,client,i)
        if err != nil {
            return err
        }
    }
    return nil
}

func DockerNetworkDestroyAll(client *util.SshClient) error {
    _,err := client.Run("for net in $(docker network ls | grep wb_v | awk '{print $1}'); do docker network rm $net; done")
    return err
}

func DockerPull(clients []*util.SshClient,image string) error {
    for _,client := range clients {
        _,err := client.Run("docker pull " + image)
        if err != nil {
            return err
        }
    }
    return nil
}

func DockerRun(server db.Server,client *util.SshClient,resources Resources,node int,image string) error {
    command := "docker run -itd "
    command += fmt.Sprintf("--network wb_vlan_%d",node)

    if !resources.NoCpuLimits() {
        command += fmt.Sprintf(" --cpus %s",resources.Cpus)
    }

    if !resources.NoMemoryLimits() {
        mem,err := resources.GetMemory()
        if err != nil {
            return errors.New("Invalid value for memory")
        }
        command += fmt.Sprintf(" --memory %d",mem)
    }

    command += fmt.Sprintf(" --ip %s",util.GetNodeIP(server.ServerID,node))
    command += fmt.Sprintf(" --hostname whiteblock-node%d",node)
    command += fmt.Sprintf(" --name whiteblock-node%d",node)
    command += " " + image
    _,err := client.Run(command)
    return err
}


func DockerRunAll(server db.Server,client *util.SshClient,resources Resources,nodes int,image string) error {
    for i := 0; i < nodes; i++{
        err := DockerRun(server,client,resources,i,image)
        if err != nil {
            return err
        }
    }
    return nil
}


