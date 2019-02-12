package cosmos

import(
    "fmt"
    "log"
    "strings"
    util "../../util"
    db "../../db"
    state "../../state"
)

var conf *util.Config

func init(){
    conf = util.GetConfig()
}

func Build(data map[string]interface{},nodes int,servers []db.Server,clients []*util.SshClient) ([]string,error){

    peers := []string{}
    state.SetBuildSteps(4+(nodes*3))

    state.SetBuildStage("Setting up the first node")
    /**
     * Set up first node
     */
    res,err := clients[0].DockerExec(0,"gaiad init --chain-id=whiteblock whiteblock");
    if err != nil{
        log.Println(res)
        log.Println(err)
        return nil,err
    }
    state.IncrementBuildProgress()
    res,err = clients[0].DockerExec(0,"bash -c 'echo \"password\\n\" | gaiacli keys add validator -ojson'")
    if err != nil{
        log.Println(res)
        log.Println(err)
        return nil,err
    }


    res,err = clients[0].DockerExec(0,"gaiacli keys show validator -a")
    if err != nil{
        log.Println(res)
        log.Println(err)
        return nil,err
    }
    state.IncrementBuildProgress()
    res,err = clients[0].DockerExec(0,fmt.Sprintf("gaiad add-genesis-account %s 100000000stake,100000000validatortoken",res[:len(res) -1]))
    if err != nil{
        log.Println(res)
        log.Println(err)
        return nil,err
    }

    res,err = clients[0].DockerExec(0,"bash -c 'echo \"password\\n\" | gaiad gentx --name validator'")
    if err != nil{
        log.Println(res)
        log.Println(err)
        return nil,err
    }
    state.IncrementBuildProgress()
    res,err = clients[0].DockerExec(0,"gaiad collect-gentxs")
    if err != nil{
        log.Println(res)
        log.Println(err)
        return nil,err
    }
    genesisFile,err := clients[0].DockerExec(0,"cat /root/.gaiad/config/genesis.json")
    if err != nil{
        log.Println(res)
        log.Println(err)
        return nil,err
    }
    state.IncrementBuildProgress()
    state.SetBuildStage("Initializing the rest of the nodes")
    node := 0
    for i, server := range servers {
        for j, ip := range server.Ips{
            if node != 0 {
                //init everything
                res,err = clients[i].DockerExec(j,"gaiad init --chain-id=whiteblock whiteblock");
                if err != nil{
                    log.Println(res)
                    log.Println(err)
                    return nil,err
                }
            }
           

            //Get the node id
            res,err = clients[i].DockerExec(j,"gaiad tendermint show-node-id")
            if err != nil{
                log.Println(res)
                log.Println(err)
                return nil,err
            }
            nodeId := res[:len(res)-1]
            peers = append(peers,fmt.Sprintf("%s@%s:26656",nodeId,ip))

            state.IncrementBuildProgress()
            node++
        }
    }


    state.SetBuildStage("Copying the genesis file to each node")
    err = util.Write("./genesis.json",genesisFile)
    if err != nil {
        log.Println(err)
        return nil,err
    }
    defer util.Rm("./genesis.json")
    //distribute the created genensis file among the nodes
    for i, server := range servers {
        err := clients[i].Scp("./genesis.json", "/home/appo/genesis.json")
        if err != nil {
            log.Println(err)
            return nil,err
        }
        defer clients[i].Run("rm /home/appo/genesis.json")

        for j, _ := range server.Ips{
            if i == 0 && j == 0 {
                state.IncrementBuildProgress()
                continue
            }
            res,err := clients[i].DockerExec(j,"rm /root/.gaiad/config/genesis.json")
            if err != nil{
                log.Println(res)
                log.Println(err)
                return nil,err
            }
            res,err = clients[i].Run(fmt.Sprintf("docker cp /home/appo/genesis.json whiteblock-node%d:/root/.gaiad/config/", j))
            if err != nil{
                log.Println(res)
                log.Println(err)
                return nil,err
            }
            state.IncrementBuildProgress()
        }
    }
    log.Printf("%v",peers)
    state.SetBuildStage("Starting cosmos")
    node = 0
    for i, server := range servers {
        for j, _ := range server.Ips{
            cmd := fmt.Sprintf("gaiad start --p2p.persistent_peers=%s",
                                strings.Join(append(peers[:node],peers[node+1:]...),","))
            res,err := clients[i].DockerExecd(j,cmd)
            if err != nil{
                log.Println(res)
                log.Println(err)
                return nil,err
            }
            node++
            state.IncrementBuildProgress()
        }
    }
    return nil,nil
}
