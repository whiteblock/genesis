package eth

import (
    "encoding/json"
    "context"
    "fmt"
    "golang.org/x/sync/semaphore"
    "regexp"
    "errors"
    "log"
    "os"
    util "../../util"
    db "../../db"
    state "../../state"
)

var conf *util.Config

func init() {
    conf = util.GetConfig()
}

const ETH_NET_STATS_PORT = 3338

/**
 * Build the Ethereum Test Network
 * @param  map[string]interface{}   data    Configuration Data for the network
 * @param  int      nodes       The number of nodes in the network
 * @param  []Server servers     The list of servers passed from build
 */
func Build(data map[string]interface{},nodes int,servers []db.Server,clients []*util.SshClient,
           buildState *state.BuildState) ([]string,error) {
    //var mutex = &sync.Mutex{}
    var sem = semaphore.NewWeighted(conf.ThreadLimit)
    ctx := context.TODO()
    ethconf,err := NewConf(data)
    if err != nil {
        log.Println(err)
        return nil,err
    }

    err = util.Rm("tmp")
    if err != nil {
        log.Println(err)
    }

    buildState.SetBuildSteps(8+(5*nodes))
    defer func(){
        fmt.Printf("Cleaning up...")
        util.Rm("tmp")
        fmt.Printf("done\n")
    }()
    
    for i := 1; i <= nodes; i++ {
        err = util.Mkdir(fmt.Sprintf("./tmp/node%d",i))
        if err != nil{
            log.Println(err)
            return nil,err
        }
        //fmt.Printf("---------------------  CREATING pre-allocated accounts for NODE-%d  ---------------------\n",i)

    }
    buildState.IncrementBuildProgress() 

    /**Create the Password files**/
    {
        var data string
        for i := 1; i <= nodes; i++{
            data += "second\n"
        }

        for i := 1; i <= nodes; i++{
            err = util.Write(fmt.Sprintf("tmp/node%d/passwd.file",i),data)
            if err != nil{
                log.Println(err)
                return nil,err
            }
        }
    }
    buildState.IncrementBuildProgress()


    /**Create the wallets**/
    wallets := []string{}
    buildState.SetBuildStage("Creating the wallets")
    for i := 1; i <= nodes; i++{

        node := i
        //sem.Acquire(ctx,1)
        gethResults,err := util.BashExec(
            fmt.Sprintf("geth --datadir tmp/node%d/ --password tmp/node%d/passwd.file account new",
                node,node))
        if err != nil {
            log.Println(err)
            return nil,err
        }
        //fmt.Printf("RAW:%s\n",gethResults)
        addressPattern := regexp.MustCompile(`\{[A-z|0-9]+\}`)
        addresses := addressPattern.FindAllString(gethResults,-1)
        if len(addresses) < 1 {
            return nil,errors.New("Unable to get addresses")
        }
        address := addresses[0]
        address = address[1:len(address)-1]
        //sem.Release(1)
        //fmt.Printf("Created wallet with address: %s\n",address)
        //mutex.Lock()
        wallets = append(wallets,address)
        //mutex.Unlock()
        buildState.IncrementBuildProgress() 
        
    }
    buildState.IncrementBuildProgress()
    unlock := ""

    for i,wallet := range wallets {
        if i != 0 {
            unlock += ","
        }
        unlock += wallet
    }
    fmt.Printf("unlock = %s\n%+v\n\n",wallets,unlock)

    buildState.IncrementBuildProgress()
    buildState.SetBuildStage("Creating the genesis block")
    err = createGenesisfile(ethconf,wallets)
    if err != nil{
        log.Println(err)
        return nil,err
    }

    buildState.IncrementBuildProgress()
    buildState.SetBuildStage("Bootstrapping network")
    err = initNodeDirectories(nodes,ethconf.NetworkId,servers)
    if err != nil {
        log.Println(err)
        return nil,err
    }

    buildState.IncrementBuildProgress()
    err = util.Mkdir("tmp/keystore")
    if err != nil {
        log.Println(err)
        return nil,err
    }
    buildState.SetBuildStage("Distributing keys")
    err = distributeUTCKeystore(nodes)
    if err != nil {
        log.Println(err)
        return nil,err
    }

    buildState.IncrementBuildProgress()
    buildState.SetBuildStage("Starting geth")
    node := 0
    for i, server := range servers {
        clients[i].Scp("tmp/CustomGenesis.json","/home/appo/CustomGenesis.json")
        defer clients[i].Run("rm -f /home/appo/CustomGenesis.json")
        for j, ip := range server.Ips{
            sem.Acquire(ctx,1)
            fmt.Printf("-----------------------------  Starting NODE-%d  -----------------------------\n",node)

            go func(networkId int64,node int,server string,num int,unlock string,nodeIP string, i int){
                defer sem.Release(1)
                name := fmt.Sprintf("whiteblock-node%d",num)
                _,err := clients[i].Run(fmt.Sprintf("rm -rf tmp/node%d",node))
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }
                err = clients[i].Scpr(fmt.Sprintf("tmp/node%d",node))
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }
                buildState.IncrementBuildProgress() 
                gethCmd := fmt.Sprintf(
                    `geth --datadir /whiteblock/node%d --maxpeers %d --networkid %d --rpc --rpcaddr %s`+
                        ` --rpcapi "web3,db,eth,net,personal,miner,txpool" --rpccorsdomain "0.0.0.0" --mine --unlock="%s"`+
                        ` --password /whiteblock/node%d/passwd.file --etherbase %s console`,
                            node,
                            ethconf.MaxPeers,
                            networkId,
                            nodeIP,
                            unlock,
                            node,
                            wallets[node-1])
                
                clients[i].Run(fmt.Sprintf("docker cp /home/appo/CustomGenesis.json whiteblock-node%d:/",node))
               
                clients[i].Run(fmt.Sprintf("docker exec %s mkdir -p /whiteblock/node%d/",name,node))
                clients[i].Run(fmt.Sprintf("docker cp ~/tmp/node%d %s:/whiteblock",node,name))
                clients[i].Run(fmt.Sprintf("docker exec -d %s tmux new -s whiteblock -d",name))
                clients[i].Run(fmt.Sprintf("docker exec -d %s tmux send-keys -t whiteblock '%s' C-m",name,gethCmd))
                
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }
                
                buildState.IncrementBuildProgress() 
            }(ethconf.NetworkId,node+1,server.Addr,j,unlock,ip,i)
            node ++
        }
    }
    err = sem.Acquire(ctx,conf.ThreadLimit)
    if err != nil{
        log.Println(err)
        return nil,err
    }
    buildState.IncrementBuildProgress()
    sem.Release(conf.ThreadLimit)
    if !buildState.ErrorFree(){
        return nil,buildState.GetError()
    }

    err = setupEthNetStats(clients[0])
    node = 0
    for i,server := range servers {
        for j,ip := range server.Ips{
            sem.Acquire(ctx,1)
            go func(i int,nodeIP string,ethnetIP string,absNum int,relNum int){
                relName := fmt.Sprintf("whiteblock-node%d",relNum)
                absName := fmt.Sprintf("whiteblock-node%d",absNum)
                sedCmd := fmt.Sprintf(`docker exec %s sed -i -r 's/"INSTANCE_NAME"(\s)*:(\s)*"(\S)*"/"INSTANCE_NAME"\t: "%s"/g' /eth-net-intelligence-api/app.json`,relName,absName)
                sedCmd2 := fmt.Sprintf(`docker exec %s sed -i -r 's/"WS_SERVER"(\s)*:(\s)*"(\S)*"/"WS_SERVER"\t: "http:\/\/%s:%d"/g' /eth-net-intelligence-api/app.json`,relName,ethnetIP,ETH_NET_STATS_PORT)
                sedCmd3 := fmt.Sprintf(`docker exec %s sed -i -r 's/"RPC_HOST"(\s)*:(\s)*"(\S)*"/"RPC_HOST"\t: "%s"/g' /eth-net-intelligence-api/app.json`,relName,nodeIP)

                //sedCmd3 := fmt.Sprintf("docker exec -it %s sed -i 's/\"WS_SECRET\"(\\s)*:(\\s)*\"[A-Z|a-z|0-9| ]*\"/\"WS_SECRET\"\\t: \"second\"/g' /eth-net-intelligence-api/app.json",container)
                res,err := clients[i].DockerExecd(relNum,"tmux new -s ethnet -d")
                if err != nil {
                    log.Println(err)
                    log.Println(res)
                    buildState.ReportError(err)
                    return
                }
                res,err = clients[i].Run(sedCmd)
                if err != nil {
                    log.Println(err)
                    log.Println(res)
                    buildState.ReportError(err)
                    return
                }
                res,err = clients[i].Run(sedCmd2)
                if err != nil {
                    log.Println(err)
                    log.Println(res)
                    buildState.ReportError(err)
                    return
                }
                _,err = clients[i].Run(sedCmd3)
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }
                _,err = clients[i].DockerExecd(relNum,"tmux send-keys -t ethnet 'cd /eth-net-intelligence-api && pm2 start app.json' C-m")
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }  
                
    
                sem.Release(1)
                buildState.IncrementBuildProgress()
            }(i,ip,servers[0].Iaddr.Ip,node,j)
            node++
        }
    }

    err = sem.Acquire(ctx,conf.ThreadLimit)
    if err != nil {
        log.Println(err)
        return nil,err
    }

    sem.Release(conf.ThreadLimit)
    return nil,nil
    
}
/***************************************************************************************************************************/


func Add(data map[string]interface{},nodes int,servers []db.Server,clients []*util.SshClient,
         newNodes map[int][]string,buildState *state.BuildState) ([]string,error) {
    return nil,nil
}

func MakeFakeAccounts(accs int) []string {  
    out := make([]string,accs)
    for i := 1; i <= accs; i++ {
        acc := fmt.Sprintf("%X",i)
        for j := len(acc); j < 40; j++ {
                acc = "0"+acc
            }
        acc = "0x"+acc
        out[i-1] = acc
    }
    return out
}


/**
 * Create the custom genesis file for Ethereum
 * @param  *EthConf ethconf     The chain configuration
 * @param  []string wallets     The wallets to be allocated a balance
 */

func createGenesisfile(ethconf *EthConf,wallets []string) error {
   



    file,err := os.Create("tmp/CustomGenesis.json")
    if err != nil {
        log.Println(err)
        return err
    }
    defer file.Close()

    genesis := fmt.Sprintf(
`{
    "config": {
        "chainId": %d,
        "homesteadBlock": %d,
        "eip155Block": %d,
        "eip158Block": %d
    },
    "difficulty": "0x0%X",
    "gasLimit": "0x0%X",
    "alloc": {`,
    ethconf.ChainId,
    ethconf.HomesteadBlock,
    ethconf.Eip155Block,
    ethconf.Eip158Block,
    ethconf.Difficulty,
    ethconf.GasLimit)

    _,err = file.Write([]byte(genesis))
    if err != nil {
        log.Println(err)
        return err
    }

    //Fund the accounts
    _,err = file.Write([]byte("\n"))
    if err != nil {
        log.Println(err)
        return err
    }
    for i,wallet := range wallets {
        _,err = file.Write([]byte(fmt.Sprintf("\t\t\"%s\":{\"balance\":\"%s\"}",wallet,ethconf.InitBalance)))
        if err != nil {
            log.Println(err)
            return err
        }
        if len(wallets) - 1 != i {
            _,err = file.Write([]byte(","))
            if err != nil {
                log.Println(err)
                return err
            }
        }
        _,err = file.Write([]byte("\n"))
        if err != nil {
            log.Println(err)
            return err
        }
    }
    if ethconf.ExtraAccounts > 0 {
        _,err = file.Write([]byte(","))
        if err != nil {
            log.Println(err)
            return err
        }
    }
    accs := MakeFakeAccounts(int(ethconf.ExtraAccounts))
    log.Println("Finished making fake accounts")
    lenAccs := len(accs)
    for i,wallet := range accs {
        _,err = file.Write([]byte(fmt.Sprintf("\t\t\"%s\":{\"balance\":\"%s\"}",wallet,ethconf.InitBalance)))
        if err != nil {
            log.Println(err)
            return err
        }

        if lenAccs - 1 != i {
            _,err = file.Write([]byte(",\n"))
            if err != nil {
                log.Println(err)
                return err
            }
        }else{
            _,err = file.Write([]byte("\n"))
            if err != nil {
                log.Println(err)
                return err
            }
        }
        
    }


    _,err = file.Write([]byte("\n\t}\n}"))
    if err != nil {
        log.Println(err)
        return err
    }
    return nil
}

/**
 * Creates the datadir for a node and returns the enode address
 * @param  int      node        The nodes number
 * @param  int64    networkId   The test net network id
 * @param  string   ip          The node's IP address
 * @return string               The node's enode address
 */
func initNode(node int, networkId int64,ip string) (string,error) {
    fmt.Printf("---------------------  CREATING block directory for NODE-%d ---------------------\n",node)
    gethResults,err := util.BashExec(fmt.Sprintf("echo -e \"admin.nodeInfo.enode\\nexit\\n\" |  geth --rpc --datadir tmp/node%d/ --networkid %d console",node,networkId))
    if err != nil{
        log.Println(err)
        return "",nil
    }
    //fmt.Printf("RAWWWWWWWWWWWW%s\n\n\n",gethResults)
    enodePattern := regexp.MustCompile(`enode:\/\/[A-z|0-9]+@(\[\:\:\]|([0-9]|\.)+)\:[0-9]+`)
    enode := enodePattern.FindAllString(gethResults,1)[0]
    fmt.Printf("ENODE fetched is: %s\n",enode)
    enodeAddressPattern := regexp.MustCompile(`\[\:\:\]|([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})`)
    enode = enodeAddressPattern.ReplaceAllString(enode,ip);

    err = util.Write(fmt.Sprintf("./tmp/node%d/enode",node),fmt.Sprintf("%s\n",enode))
    return enode,err
}

/**
 * Initialize the chain from the custom genesis file
 * @param  int      nodes       The number of nodes
 * @param  int64    networkId   The test net network id
 * @param  []Server servers     The list of servers
 */
func initNodeDirectories(nodes int,networkId int64,servers []db.Server) error {
    static_nodes := []string{};
    node := 1
    for _,server := range servers{
        for _,ip := range server.Ips{
            //fmt.Printf("---------------------  CREATING block directory for NODE-%d ---------------------\n",i)
            //Load the CustomGenesis file
            res,err := util.BashExec(
                            fmt.Sprintf("geth --datadir tmp/node%d --networkid %d init tmp/CustomGenesis.json",node,networkId))
            if err != nil {
                log.Println(res)
                log.Println(err)
                return err
            }
            static_node,err := initNode(node,networkId,ip)
            if err != nil {
                log.Println(err)
                return err
            }
            static_nodes = append(static_nodes,static_node)
            node++;
        }
    }
    out, err := json.Marshal(static_nodes)
    //fmt.Printf("-----Static Nodes.json------\n%+v\n\n",static_nodes)
    if err != nil {
        log.Println(err)
        return err
    }

    for i := 1; i <= nodes; i++ {
        err = util.Write(fmt.Sprintf("tmp/node%d/static-nodes.json",i),string(out))
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
        err := util.Cpr(fmt.Sprintf("tmp/node%d/keystore/",i),"tmp/")
        if err != nil {
            log.Println(err)
            return err
        }
    }
    for i := 1; i <= nodes; i++ {
        err := util.Cpr("tmp/keystore/",fmt.Sprintf("tmp/node%d/",i))
        if err != nil{
            log.Println(err)
            return err
        }
    }
    return nil
}



/**
 * Setup Eth Net Stats on a server
 * @param  string    ip     The servers config
 */
func setupEthNetStats(client *util.SshClient) error {
    res,err := client.Run("[ -d ~/eth-netstats ] && echo \"success\"")
    if res != "success" || err != nil {
        log.Println("eth-net stats not found in server!")
        client.Run("rm -rf ~/eth-netstats")//ign
        _,err = client.Run("wget http://whiteblock.io/eth-netstats.tar.gz && tar xf eth-netstats.tar.gz && rm eth-netstats.tar.gz")
        if err != nil {
            log.Println(err)
            return err
        }
    }

    client.Run("tmux kill-session -t netstats")//ign
    
    _,err = client.Run("tmux new -s netstats -d")
    if err != nil {
        log.Println(err)
        return err
    }
    _,err = client.Run(fmt.Sprintf(
        "tmux send-keys -t netstats 'cd /home/appo/eth-netstats && npm install && grunt && WS_SECRET=second PORT=%d npm start' C-m",ETH_NET_STATS_PORT))
    if err != nil {
        log.Println(err)
        return err
    }
    return nil
}