package parity

import (
    "context"
    "errors"
    "fmt"
    "log"
    "regexp"
    "strings"

    db "../../db"
    state "../../state"
    util "../../util"
    "golang.org/x/sync/semaphore"
)

var conf *util.Config

func init() {
    conf = util.GetConfig()
}

/**
 * Build the Ethereum Test Network
 * @param  map[string]interface{}   data    Configuration Data for the network
 * @param  int      nodes       The number of nodes in the network
 * @param  []Server servers     The list of servers passed from build
 */
func Build(data map[string]interface{}, nodes int, servers []db.Server, clients []*util.SshClient,
           buildState *state.BuildState) ([]string, error) {
    //var mutex = &sync.Mutex{}
    var sem = semaphore.NewWeighted(conf.ThreadLimit)
    ctx := context.TODO()
    pconf, err := NewConf(data)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    err = util.Rm("tmp")
    if err != nil {
        log.Println(err)
    }

    buildState.SetBuildSteps(8 + (5 * nodes))
    defer func() {
        fmt.Printf("Cleaning up...")
        util.Rm("tmp")
        fmt.Printf("done\n")
    }()

    for i := 1; i <= nodes; i++ {
        err = util.Mkdir(fmt.Sprintf("./tmp/node%d", i))
        if err != nil {
            log.Println(err)
            return nil, err
        }
        //fmt.Printf("---------------------  CREATING pre-allocated accounts for NODE-%d  ---------------------\n",i)
    }

    //Make the data directories
    for i, server := range servers {
        for j, _ := range server.Ips {
            res,err := clients[i].DockerExec(j,"mkdir -p /parity")
            if err != nil {
                log.Println(res)
                log.Println(err)
                return nil, err
            }
        }
    }
    /**Create the Password file**/
    {
        var data string
        for i := 1; i <= nodes; i++ {
            data += "second\n"
        }
        err = util.Write("./passwd", data)
        if err != nil {
            log.Println(err)
            return nil, err
        }
    }
    defer util.Rm("./passwd")
    /**Copy over the password file**/
    for i, server := range servers {
        err = clients[i].Scp("./passwd", "/home/appo/passwd")
        if err != nil {
            log.Println(err)
            return nil, err
        }
        defer clients[i].Run("rm /home/appo/passwd")

        for j, _ := range server.Ips {
            err = clients[i].DockerCp(j,"/home/appo/passwd","/parity/")
            if err != nil {
                log.Println(err)
                return nil, err
            }
        }
    }
    

    /**Create the wallets**/
    wallets := []string{}
    rawWallets := []string{}
    for i, server := range servers {
        for j, _ := range server.Ips {
            res,err := clients[i].DockerExec(j,
                    fmt.Sprintf("parity --base-path=/parity/ --password=/parity/passwd account new"))
            if err != nil{
                log.Println(res)
                log.Println(err)
                return nil, err
            }
            if len(res) == 0{
                return nil,errors.New("account new returned an empty response")
            }

            address := res[:len(res)-1]
            wallets = append(wallets,address)

            res,err = clients[i].DockerExec(j,"bash -c 'cat /parity/keys/ethereum/*'")
            if err != nil{
                log.Println(res)
                log.Println(err)
                return nil, err
            }

            rawWallets = append(rawWallets,strings.Replace(res,"\"","\\\"",-1)) 
        }
    }

    //Create the chain spec files
    spec,err := BuildSpec(pconf,wallets);
    if err != nil {
        log.Println(err)
        return nil,err
    }
    err = util.Write("./spec.json",spec)
    if err != nil {
        log.Println(err)
        return nil,err
    }
    defer util.Rm("./spec.json")

    //create config file
    configToml,err := BuildConfig(pconf,wallets,"/parity/passwd")
    if err != nil {
        log.Println(err)
        return nil, err
    }

    err = util.Write("./config.toml",configToml)
    if err != nil {
        log.Println(err)
        return nil,err
    }
    defer util.Rm("./config.toml")

    //Copy over the config file, spec file, and the accounts
    node := 0
    for i, server := range servers {
        err = clients[i].Scp("./config.toml", "/home/appo/config.toml")
        if err != nil {
            log.Println(err)
            return nil, err
        }
        defer clients[i].Run("rm /home/appo/config.toml")

        err = clients[i].Scp("./spec.json", "/home/appo/spec.json")
        if err != nil {
            log.Println(err)
            return nil, err
        }
        defer clients[i].Run("rm /home/appo/spec.json")

        for j, _ := range server.Ips {
            err = clients[i].DockerCp(j,"/home/appo/spec.json","/parity/")
            if err != nil {
                log.Println(err)
                return nil, err
            }

            err = clients[i].DockerCp(j,"/home/appo/config.toml","/parity/")
            if err != nil {
                log.Println(err)
                return nil, err
            }
            
            for k,rawWallet := range rawWallets {
                if k == node {
                    continue
                }

                _,err = clients[i].DockerExec(node,fmt.Sprintf("bash -c 'echo \"%s\">>/parity/account%d'",rawWallet,k))
                if err != nil {
                    log.Println(err)
                    return nil, err
                }
                defer clients[i].DockerExec(node,fmt.Sprintf("rm /parity/account%d",k))

                res,err := clients[i].DockerExec(j,
                    fmt.Sprintf("parity --base-path=/parity/ --password=/parity/passwd account import /parity/account%d",k))
                if err != nil{
                    log.Println(res)
                    log.Println(err)
                    return nil, err
                }
            }
            node++
        }
    }
    return nil,nil
    buildState.SetBuildStage("Bootstrapping network")
    err = initNodeDirectories(nodes, pconf.NetworkId, servers)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    buildState.IncrementBuildProgress()
    err = util.Mkdir("tmp/keystore")
    if err != nil {
        log.Println(err)
        return nil, err
    }
    buildState.SetBuildStage("Distributing keys")
    err = distributeUTCKeystore(nodes)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    buildState.IncrementBuildProgress()
    buildState.SetBuildStage("Starting Parity")

    
    util.Write("tmp/config.toml",configToml)
    node = 0
    for i, server := range servers {
        clients[i].Scp("tmp/config.toml", "/home/appo/config.toml")
        clients[i].Scp("tmp/spec.json", "/home/appo/spec.json")
        defer clients[i].Run("rm -f /home/appo/config.toml")
        defer clients[i].Run("rm -f /home/appo/spec.json")
        for j, ip := range server.Ips {
            sem.Acquire(ctx, 1)
            fmt.Printf("-----------------------------  Starting NODE-%d  -----------------------------\n", node)

            go func(networkId int64, node int, server string, num int, nodeIP string, i int) {
                defer sem.Release(1)

                clients[i].Run(fmt.Sprintf("docker cp /home/appo/config.toml whiteblock-node%d:/", node))
                clients[i].Run(fmt.Sprintf("docker cp /home/appo/spec.json whiteblock-node%d:/", node))

                name := fmt.Sprintf("whiteblock-node%d", num)

                err = clients[i].Scpr(fmt.Sprintf("tmp/node%d", node))
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }
                defer clients[i].Run(fmt.Sprintf("rm -rf tmp/node%d", node))

                buildState.IncrementBuildProgress()

                parityCmd := fmt.Sprintf(
                    `parity --base-path=/whiteblock/node%d --network-id=%d`+
                        `--jsonrpc-apis="all" --jsonrpc-cors="all" `+
                        ` --password /whiteblock/node%d/passwd.file --author=%s
                        --reserved-peers -c /config.toml --chain=/spec.json`,
                    node,
                    networkId,
                    node,
                    wallets[node-1])

                

                clients[i].DockerExec(num,fmt.Sprintf("mkdir -p /whiteblock/node%d/", node))
                clients[i].Run(fmt.Sprintf("docker cp ~/tmp/node%d %s:/whiteblock", node, name))
                clients[i].DockerExecd(num,fmt.Sprintf("tmux new -s whiteblock -d", name))
                clients[i].DockerExecd(num,fmt.Sprintf("%s tmux send-keys -t whiteblock '%s' C-m", name, parityCmd))

                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }

                buildState.IncrementBuildProgress()
            }(pconf.NetworkId, node+1, server.Addr, j, ip, i)
            node++
        }
    }
    err = sem.Acquire(ctx, conf.ThreadLimit)
    if err != nil {
        log.Println(err)
        return nil, err
    }
    buildState.IncrementBuildProgress()
    sem.Release(conf.ThreadLimit)
    if !buildState.ErrorFree() {
        return nil, buildState.GetError()
    }
    return nil, nil
}

/***************************************************************************************************************************/

func Add(data map[string]interface{},nodes int,servers []db.Server,clients []*util.SshClient,
         newNodes map[int][]string,buildState *state.BuildState) ([]string,error) {
    return nil,nil
}

func MakeFakeAccounts(accs int) []string {
    out := make([]string, accs)
    for i := 1; i <= accs; i++ {
        acc := fmt.Sprintf("%X", i)
        for j := len(acc); j < 40; j++ {
            acc = "0" + acc
        }
        acc = "0x" + acc
        out[i-1] = acc
    }
    return out
}

/**
 * Creates the datadir for a node and returns the enode address
 * @param  int      node        The nodes number
 * @param  int64    networkId   The test net network id
 * @param  string   ip          The node's IP address
 * @return string               The node's enode address
 */
func initNode(node int, networkId int64, ip string) (string, error) {
    fmt.Printf("---------------------  CREATING block directory for NODE-%d ---------------------\n", node)
    parityResults, err := util.BashExec(fmt.Sprintf("echo -e \"admin.nodeInfo.enode\\nexit\\n\" |  parity --rpc --datadir tmp/node%d/ --networkid %d console", node, networkId))
    if err != nil {
        log.Println(err)
        return "", nil
    }
    enodePattern := regexp.MustCompile(`enode:\/\/[A-z|0-9]+@(\[\:\:\]|([0-9]|\.)+)\:[0-9]+`)
    enode := enodePattern.FindAllString(parityResults, 1)[0]
    fmt.Printf("ENODE fetched is: %s\n", enode)
    enodeAddressPattern := regexp.MustCompile(`\[\:\:\]|([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})`)
    enode = enodeAddressPattern.ReplaceAllString(enode, ip)

    err = util.Write(fmt.Sprintf("./tmp/node%d/enode", node), fmt.Sprintf("%s\n", enode))
    return enode, err
}

/**
 * Initialize the chain from the configuration file
 * @param  int      nodes       The number of nodes
 * @param  int64    networkId   The test net network id
 * @param  []Server servers     The list of servers
 */
func initNodeDirectories(nodes int, networkId int64, servers []db.Server) error {
    static_nodes := []string{}
    node := 1
    for _, server := range servers {
        for _, ip := range server.Ips {
            res, err := util.BashExec(
                fmt.Sprintf("parity --config tmp/node%d", node))
            if err != nil {
                log.Println(res)
                log.Println(err)
                return err
            }
            static_node, err := initNode(node, networkId, ip)
            if err != nil {
                log.Println(err)
                return err
            }
            static_nodes = append(static_nodes, static_node)
            node++
        }
    }

    snodes := strings.Join(static_nodes, ",")

    for i := 1; i <= nodes; i++ {
        err := util.Write(fmt.Sprintf("tmp/node%d/peers.txt", i), snodes)
        if err != nil {
            log.Println(err)
            return err
        }
    }

    return nil
}

/**
 * Distribute the UTC keystore files amongst the nodes
 * @param  int  nodes   The number of nodes
 */
func distributeUTCKeystore(nodes int) error {
    //Copy all UTC keystore files to every Node directory
    for i := 1; i <= nodes; i++ {
        err := util.Cpr(fmt.Sprintf("tmp/node%d/keystore/", i), "tmp/")
        if err != nil {
            log.Println(err)
            return err
        }
    }
    for i := 1; i <= nodes; i++ {
        err := util.Cpr("tmp/keystore/", fmt.Sprintf("tmp/node%d/", i))
        if err != nil {
            log.Println(err)
            return err
        }
    }
    return nil
}
