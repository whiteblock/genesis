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
    _,err := client.Run(fmt.Sprintf("docker rm -f whiteblock-node%d",node))
    return err
}


func DockerKillAll(client *util.SshClient) error {
    _,err := client.Run("docker rm -f $(docker ps -aq -f name=whiteblock)");
    state.IncrementDeployProgress()
    return err
}

func DockerNetworkCreate(server db.Server,client *util.SshClient,node int) error {
    command := "docker network create -d macvlan"
    command += fmt.Sprintf(" --subnet %s",util.GetNetworkAddress(server.ServerID,node))
    command += fmt.Sprintf(" --gateway %s",util.GetGateway(server.ServerID,node))
    command += fmt.Sprintf(" -o parent=%s.%d wb_vlan_%d",server.Iface,node+101,node)
    res,err := client.Run(command)
    if err != nil{
        return errors.New(res)
    }
    return nil
}

func DockerNetworkCreateAll(server db.Server,client *util.SshClient,nodes int) error {
    for i := 0; i < nodes; i++{
        state.IncrementDeployProgress()
        err := DockerNetworkCreate(server,client,i)
        if err != nil {
            return err
        }
    }
    return nil
}

func DockerNetworkDestroyAll(client *util.SshClient) error {
    _,err := client.Run("for net in $(docker network ls | grep wb_v | awk '{print $1}'); do docker network rm $net; done")
    state.IncrementDeployProgress()
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

func dockerRunCmd(server db.Server,resources Resources,node int,image string) (string,error) {
    command := "docker run -itd "
    command += fmt.Sprintf("--network wb_vlan_%d",node)

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
    command += fmt.Sprintf(" --hostname whiteblock-node%d",node)
    command += fmt.Sprintf(" --name whiteblock-node%d",node)
    command += " " + image
    return command,nil
}

func DockerRun(server db.Server,client *util.SshClient,resources Resources,node int,image string) error {
    command,err := dockerRunCmd(server,resources,node,image)
    if err != nil{
        return err
    }
    _,err = client.Run(command)
    return err
}


func DockerRunAll(server db.Server,client *util.SshClient,resources Resources,nodes int,image string) error {
    var command string
    for i := 0; i < nodes; i++ {
        state.IncrementDeployProgress()
        tmp,err := dockerRunCmd(server,resources,i,image)
        if err != nil{
            return err
        }

        if len(command) == 0 {
            command += tmp
        }else{
            command += "&&" + tmp
        }

        if i % 2 == 0 || i == nodes - 1 {
            _,err = client.Run(command)
            command = ""
            if err != nil {
                log.Println(err)
                return err
            }
        }
        
    }
    return nil
}