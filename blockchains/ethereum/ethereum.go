package eth

import (
    "encoding/json"
    "context"
    "fmt"
    "golang.org/x/sync/semaphore"
    "regexp"
    "errors"
    "log"
    util "../../util"
    db "../../db"
    state "../../state"
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
func Ethereum(data map[string]interface{},nodes int,servers []db.Server,clients []*util.SshClient) error {
    //var mutex = &sync.Mutex{}
    var sem = semaphore.NewWeighted(conf.ThreadLimit)
    ctx := context.TODO()
    ethconf,err := NewConf(data)
    if err != nil {
        log.Println(err)
        return err
    }

    err = util.Rm("tmp")
    if err != nil {
        log.Println(err)
    }

    state.SetBuildSteps(8+(4*nodes))
    defer func(){
        fmt.Printf("Cleaning up...")
        util.Rm("tmp")
        fmt.Printf("done\n")
    }()
    
    for i := 1; i <= nodes; i++ {
        err = util.Mkdir(fmt.Sprintf("./tmp/node%d",i))
        if err != nil{
            log.Println(err)
            return err
        }
        //fmt.Printf("---------------------  CREATING pre-allocated accounts for NODE-%d  ---------------------\n",i)

    }
    state.IncrementBuildProgress() 

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
                return err
            }
        }
    }
    state.IncrementBuildProgress()


    /**Create the wallets**/
    wallets := []string{}

    for i := 1; i <= nodes; i++{

        node := i
        //sem.Acquire(ctx,1)
        gethResults,err := util.BashExec(
            fmt.Sprintf("geth --datadir tmp/node%d/ --password tmp/node%d/passwd.file account new",
                node,node))
        if err != nil {
            log.Println(err)
            return err
        }
        //fmt.Printf("RAW:%s\n",gethResults)
        addressPattern := regexp.MustCompile(`\{[A-z|0-9]+\}`)
        addresses := addressPattern.FindAllString(gethResults,-1)
        if len(addresses) < 1 {
            return errors.New("Unable to get addresses")
        }
        address := addresses[0]
        address = address[1:len(address)-1]
        //sem.Release(1)
        //fmt.Printf("Created wallet with address: %s\n",address)
        //mutex.Lock()
        wallets = append(wallets,address)
        //mutex.Unlock()
        state.IncrementBuildProgress() 
        
    }
    state.IncrementBuildProgress()
    unlock := ""

    for i,wallet := range wallets {
        if i != 0 {
            unlock += ","
        }
        unlock += wallet
    }
    fmt.Printf("unlock = %s\n%+v\n\n",wallets,unlock)

    state.IncrementBuildProgress()

    err = createGenesisfile(ethconf,wallets)
    if err != nil{
        log.Println(err)
        return err
    }

    state.IncrementBuildProgress()

    err = initNodeDirectories(nodes,ethconf.NetworkId,servers)
    if err != nil {
        log.Println(err)
        return err
    }

    state.IncrementBuildProgress()
    err = util.Mkdir("tmp/keystore")
    if err != nil {
        log.Println(err)
        return err
    }
    err = distributeUTCKeystore(nodes)
    if err != nil {
        log.Println(err)
        return err
    }

    state.IncrementBuildProgress()

    node := 0
    for i, server := range servers {
        for j, ip := range server.Ips{
            sem.Acquire(ctx,1)
            fmt.Printf("-----------------------------  Starting NODE-%d  -----------------------------\n",node)

            go func(networkId int64,node int,server string,num int,unlock string,nodeIP string, i int){
                defer sem.Release(1)
                name := fmt.Sprintf("whiteblock-node%d",num)
                _,err := clients[i].Run(fmt.Sprintf("rm -rf tmp/node%d",node))
                if err != nil {
                    log.Println(err)
                    state.ReportError(err)
                    return
                }
                err = clients[i].Scpr(fmt.Sprintf("tmp/node%d",node))
                if err != nil {
                    log.Println(err)
                    state.ReportError(err)
                    return
                }
                state.IncrementBuildProgress() 
                gethCmd := fmt.Sprintf(
                    `geth --datadir /whiteblock/node%d --nodiscover --maxpeers %d --networkid %d --rpc --rpcaddr %s`+
                        ` --rpcapi "web3,db,eth,net,personal,miner" --rpccorsdomain "0.0.0.0" --mine --unlock="%s"`+
                        ` --password /whiteblock/node%d/passwd.file --etherbase %s console`,
                            node,
                            ethconf.MaxPeers,
                            networkId,
                            nodeIP,
                            unlock,
                            node,
                            wallets[node-1])
                clients[i].Run(fmt.Sprintf("docker exec %s mkdir -p /whiteblock/node%d/",name,node))
                clients[i].Run(fmt.Sprintf("docker cp ~/tmp/node%d %s:/whiteblock",node,name))
                clients[i].Run(fmt.Sprintf("docker exec -d %s tmux new -s whiteblock -d",name))
                clients[i].Run(fmt.Sprintf("docker exec -d %s tmux send-keys -t whiteblock '%s' C-m",name,gethCmd))
                
                if err != nil {
                    log.Println(err)
                    state.ReportError(err)
                    return
                }
                
                state.IncrementBuildProgress() 
            }(ethconf.NetworkId,node+1,server.Addr,j,unlock,ip,i)
            node ++
        }
    }
    err = sem.Acquire(ctx,conf.ThreadLimit)
    if err != nil{
        log.Println(err)
        return err
    }
    state.IncrementBuildProgress()
    sem.Release(conf.ThreadLimit)
    if !state.ErrorFree(){
        return state.GetError()
    }
    return nil
    
}
/***************************************************************************************************************************/




/**
 * Create the custom genesis file for Ethereum
 * @param  *EthConf ethconf     The chain configuration
 * @param  []string wallets     The wallets to be allocated a balance
 */
func createGenesisfile(ethconf *EthConf,wallets []string) error {
    alloc := "\n"
    for i,wallet := range wallets {
        alloc += fmt.Sprintf("\t\t\"%s\":{\"balance\":\"%s\"}",wallet,ethconf.InitBalance)
        if len(wallets) - 1 != i {
            alloc += ","
        }
        alloc += "\n"
    }

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
    "alloc": {%s    }
}`,
    ethconf.ChainId,
    ethconf.HomesteadBlock,
    ethconf.Eip155Block,
    ethconf.Eip158Block,
    ethconf.Difficulty,
    ethconf.GasLimit,
    alloc)

    return util.Write("tmp/CustomGenesis.json",genesis)
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
            _,err := util.BashExec(
                            fmt.Sprintf("geth --datadir tmp/node%d --networkid %d init tmp/CustomGenesis.json",node,networkId))
            if err != nil {
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
