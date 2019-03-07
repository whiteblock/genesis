package geth

import (
    "encoding/json"
    "context"
    "fmt"
    "golang.org/x/sync/semaphore"
    "regexp"
    "errors"
    "strings"
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

    buildState.SetBuildSteps(8+(5*nodes))
    
    buildState.IncrementBuildProgress() 

    /**Create the Password files**/
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
    buildState.SetBuildStage("Distributing secrets")
    /**Copy over the password file**/
    for i, server := range servers {
        err = clients[i].Scp("./passwd", "/home/appo/passwd")
        if err != nil {
            log.Println(err)
            return nil, err
        }
        defer clients[i].Run("rm /home/appo/passwd")

        for j, _ := range server.Ips {
            res,err := clients[i].DockerExec(j,"mkdir -p /geth")
            if err != nil {
                log.Println(res)
                log.Println(err)
                return nil, err
            }

            err = clients[i].DockerCp(j,"/home/appo/passwd","/geth/")
            if err != nil {
                log.Println(err)
                return nil, err
            }
        }
        buildState.IncrementBuildProgress()
    }
    


    /**Create the wallets**/
    wallets := []string{}
    rawWallets := []string{}
    buildState.SetBuildStage("Creating the wallets")

    for i, server := range servers {
        for j, _ := range server.Ips {
            gethResults,err := clients[i].DockerExec(j,"geth --datadir /geth/ --password /geth/passwd account new")
            if err != nil {
                log.Println(gethResults)
                log.Println(err)
                return nil,err
            }

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

            res,err := clients[i].DockerExec(j,"bash -c 'cat /geth/keystore/*'")
            if err != nil {
                log.Println(res)
                log.Println(err)
                return nil, err
            }
            rawWallets = append(rawWallets,strings.Replace(res,"\"","\\\"",-1)) 
        }
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
    node := 0
    for i, server := range servers {
        err = clients[i].Scp("./CustomGenesis.json", "/home/appo/CustomGenesis.json")
        if err != nil {
            log.Println(err)
            return nil, err
        }
        defer clients[i].Run("rm /home/appo/CustomGenesis.json")

        for j, _ := range server.Ips {
            err = clients[i].DockerCp(j,"/home/appo/CustomGenesis.json","/geth/")
            if err != nil {
                log.Println(err)
                return nil, err
            }
            for k,rawWallet := range rawWallets {
                if k == node {
                    continue
                }
                _,err = clients[i].DockerExec(j,fmt.Sprintf("bash -c 'echo \"%s\">>/geth/keystore/account%d'",rawWallet,k))
                if err != nil {
                    log.Println(err)
                    return nil, err
                }
            }
            node++
        }
    }

    static_nodes := []string{}
    node = 0
    for i,server := range servers {
        for j,ip := range server.Ips {
            //fmt.Printf("---------------------  CREATING block directory for NODE-%d ---------------------\n",i)
            //Load the CustomGenesis file
            res,err := clients[i].DockerExec(j,
                            fmt.Sprintf("geth --datadir /geth/ --networkid %d init /geth/CustomGenesis.json",ethconf.NetworkId))
            if err != nil {
                log.Println(res)
                log.Println(err)
                return nil,err
            }
            fmt.Printf("---------------------  CREATING block directory for NODE-%d ---------------------\n",node)
            gethResults,err := clients[i].DockerExec(j,
                fmt.Sprintf("bash -c 'echo -e \"admin.nodeInfo.enode\\nexit\\n\" |  geth --rpc --datadir /geth/ --networkid %d console'",ethconf.NetworkId))
            if err != nil{
                log.Println(err)
                return nil,err
            }
            //fmt.Printf("RAWWWWWWWWWWWW%s\n\n\n",gethResults)
            enodePattern := regexp.MustCompile(`enode:\/\/[A-z|0-9]+@(\[\:\:\]|([0-9]|\.)+)\:[0-9]+`)
            enode := enodePattern.FindAllString(gethResults,1)[0]
            //fmt.Printf("ENODE fetched is: %s\n",enode)
            enodeAddressPattern := regexp.MustCompile(`\[\:\:\]|([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})`)
            enode = enodeAddressPattern.ReplaceAllString(enode,ip)
            static_nodes = append(static_nodes,enode)
            node++
        }
    }
    out, err := json.Marshal(static_nodes)
    //fmt.Printf("-----Static Nodes.json------\n%+v\n\n",static_nodes)
    if err != nil {
        log.Println(err)
        return nil,err
    }

    defer util.Rm("static-nodes.json")
    err = util.Write("static-nodes.json",string(out))
    if err != nil {
        log.Println(err)
        return nil,err
    }

    buildState.IncrementBuildProgress()
    buildState.SetBuildStage("Starting geth")
    node = 0
    for i, server := range servers {
        err = clients[i].Scp("./static-nodes.json", "/home/appo/static-nodes.json")
        if err != nil {
            log.Println(err)
            return nil, err
        }

        for j, ip := range server.Ips{
            sem.Acquire(ctx,1)
            fmt.Printf("-----------------------------  Starting NODE-%d  -----------------------------\n",node)

            go func(networkId int64,node int,server string,num int,unlock string,nodeIP string, i int){
                defer sem.Release(1)

                //name := fmt.Sprintf("whiteblock-node%d",num)
                
                buildState.IncrementBuildProgress() 

                gethCmd := fmt.Sprintf(
                    `geth --datadir /geth/ --maxpeers %d --networkid %d --rpc --rpcaddr %s`+
                        ` --rpcapi "web3,db,eth,net,personal,miner,txpool" --rpccorsdomain "0.0.0.0" --mine --unlock="%s"`+
                        ` --password /geth/passwd --etherbase %s console  2>&1 | tee output.log`,
                            ethconf.MaxPeers,
                            networkId,
                            nodeIP,
                            unlock,
                            wallets[node])
                
                clients[i].DockerCp(num,"/home/appo/static-nodes.json","/geth/")
                clients[i].DockerExecd(num,"tmux new -s whiteblock -d")
                clients[i].DockerExecd(num,fmt.Sprintf("tmux send-keys -t whiteblock '%s' C-m",gethCmd))
                
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }
                
                buildState.IncrementBuildProgress() 
            }(ethconf.NetworkId,node,server.Addr,j,unlock,ip,i)
            node++
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
                absName := fmt.Sprintf("whiteblock-node%d",absNum)
                sedCmd := fmt.Sprintf(`sed -i -r 's/"INSTANCE_NAME"(\s)*:(\s)*"(\S)*"/"INSTANCE_NAME"\t: "%s"/g' /eth-net-intelligence-api/app.json`,absName)
                sedCmd2 := fmt.Sprintf(`sed -i -r 's/"WS_SERVER"(\s)*:(\s)*"(\S)*"/"WS_SERVER"\t: "http:\/\/%s:%d"/g' /eth-net-intelligence-api/app.json`,ethnetIP,ETH_NET_STATS_PORT)
                sedCmd3 := fmt.Sprintf(`sed -i -r 's/"RPC_HOST"(\s)*:(\s)*"(\S)*"/"RPC_HOST"\t: "%s"/g' /eth-net-intelligence-api/app.json`,nodeIP)

                //sedCmd3 := fmt.Sprintf("docker exec -it %s sed -i 's/\"WS_SECRET\"(\\s)*:(\\s)*\"[A-Z|a-z|0-9| ]*\"/\"WS_SECRET\"\\t: \"second\"/g' /eth-net-intelligence-api/app.json",container)
                res,err := clients[i].DockerExecd(relNum,"tmux new -s ethnet -d")
                if err != nil {
                    log.Println(err)
                    log.Println(res)
                    buildState.ReportError(err)
                    return
                }
                res,err = clients[i].DockerExec(relNum,sedCmd)
                if err != nil {
                    log.Println(err)
                    log.Println(res)
                    buildState.ReportError(err)
                    return
                }
                res,err = clients[i].DockerExec(relNum,sedCmd2)
                if err != nil {
                    log.Println(err)
                    log.Println(res)
                    buildState.ReportError(err)
                    return
                }
                _,err = clients[i].DockerExec(relNum,sedCmd3)
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
            }(i,ip,util.GetGateway(server.ServerID,node),node,j)
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

    file,err := os.Create("CustomGenesis.json")
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
    "nonce": "0x0000000000000042",
    "difficulty": "0x0%X",
    "mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "coinbase": "0x0000000000000000000000000000000000000000",
    "timestamp": "0x00",
    "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "extraData": "0x1100000000000000000000000000000000000000000000000000000000000000",
    "gasLimit": "0x0%X",
    "alloc": {`,
    ethconf.NetworkId,
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
 * Setup Eth Net Stats on a server
 * @param  string    ip     The servers config
 */
func setupEthNetStats(client *util.SshClient) error {
    _,err := client.Run(fmt.Sprintf(
        "docker exec -d wb_service0 bash -c 'cd /eth-netstats && WS_SECRET=second PORT=%d npm start'",ETH_NET_STATS_PORT))
    if err != nil {
        log.Println(err)
        return err
    }
    return nil
}