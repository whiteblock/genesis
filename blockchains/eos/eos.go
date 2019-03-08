package eos

import (
    "context"
    "fmt"
    "golang.org/x/sync/semaphore"
    "math/rand"
    "log"
    "errors"
    db "../../db"
    util "../../util"
    state "../../state"
)

var conf *util.Config

func init(){
    conf = util.GetConfig()
}

/**
 * Setup the EOS test net
 * @param  int      nodes       The number of producers to make
 * @param  []Server servers     The list of relevant servers
 */
func Build(data map[string]interface{},nodes int,servers []db.Server,clients []*util.SshClient,buildState *state.BuildState) ([]string,error) {

    if nodes < 2 {
        return nil, errors.New("Cannot build less than 2 nodes")
    }

    eosconf,err := NewConf(data)
    if err != nil {
        log.Println(err)
        return nil,err
    }
    if eosconf.BlockProducers < 2 {
        return nil, errors.New("Cannot build eos with only one BP")
    }
    eosconf.BlockProducers++
    state.SetEOSNumberOfAccounts(eosconf.UserAccounts)
    fmt.Println("-------------Setting Up EOS-------------")
    sem := semaphore.NewWeighted(conf.ThreadLimit)
    ctx := context.TODO()

    masterIP := servers[0].Ips[0]
    masterServerIP := servers[0].Addr

    clientPasswords := make(map[string]string)

    fmt.Println("\n*** Get Key Pairs ***")
    buildState.SetBuildSteps(17 + (nodes*(3)) + (int(eosconf.UserAccounts) * (2)) + ((int(eosconf.UserAccounts)/50)*nodes))
    

    contractAccounts := []string{
        "eosio.bpay",
        "eosio.msig",
        "eosio.names",
        "eosio.ram",
        "eosio.ramfee",
        "eosio.saving",
        "eosio.stake",
        "eosio.token",
        "eosio.vpay",
    }
    km,err := NewKeyMaster()
    if err != nil{
        log.Println(err)
        return nil,err
    }
    keyPairs,err := km.GetServerKeyPairs(servers,clients)
    if err != nil{
        log.Println(err)
        return nil,err
    }

    contractKeyPairs,err := km.GetMappedKeyPairs(contractAccounts,clients[0])
    if err != nil {
        log.Println(err)
        return nil,err
    }
    buildState.IncrementBuildProgress()

    masterKeyPair := keyPairs[servers[0].Ips[0]]

    var accountNames []string
    for i := 0; i < int(eosconf.UserAccounts); i++{
        accountNames = append(accountNames,eos_getRegularName(i))
    }
    accountKeyPairs,err := km.GetMappedKeyPairs(accountNames,clients[0])
    if err != nil {
        log.Println(err)
        return nil,err
    }
    buildState.SetBuildStage("Building genesis block")
    genesis,err := eosconf.GenerateGenesis(keyPairs[masterIP].PublicKey)
    if err != nil {
        log.Println(err)
        return nil,err
    }
    err = util.Write("genesis.json",genesis)
    if err != nil{
        log.Println(err)
        return nil,err
    }
    eosConfigIni,err := eosconf.GenerateConfig()
    if err != nil {
        log.Println(err)
        return nil,err
    }
    err = util.Write("config.ini",eosConfigIni )
    if err != nil{
        log.Println(err)
        return nil,err
    }
    buildState.IncrementBuildProgress() 
    /**Start keos and add all the key pairs for all the nodes**/
    {
        buildState.SetBuildStage("Generating key pairs")
        for i, server := range servers {
            for localId,ip := range server.Ips {
                /**Start keosd**/
                _,err = clients[i].DockerExecd(localId,"keosd --http-server-address 0.0.0.0:8900")
                if err != nil{
                    log.Println(err)
                    return nil,err
                }
                clientPasswords[ip],err = eos_createWallet(clients[i], localId)
                if err != nil {
                    log.Println(err)
                    return nil,err
                }
                sem.Acquire(ctx,1)

                go func(accountKeyPairs map[string]util.KeyPair,accountNames []string,localId int,server int){
                    defer sem.Release(1)
                    cmds := []string{}
                    for _,name := range accountNames {
                        if len(cmds) > 50 {
                            _,err := clients[server].KTDockerMultiExec(localId,cmds)
                            if err != nil {
                                log.Println(err)
                                buildState.ReportError(err)
                                return
                            }
                            buildState.IncrementBuildProgress()
                            cmds = []string{}
                        }
                        
                        cmds = append(cmds,fmt.Sprintf("cleos wallet import --private-key %s", accountKeyPairs[name].PrivateKey))
                    }
                    if len(cmds) > 0 {
                        _,err := clients[server].KTDockerMultiExec(localId,cmds)
                        if err != nil {
                            log.Println(err)
                            buildState.ReportError(err)
                            return
                        }
                    }
                    buildState.IncrementBuildProgress()
                }(accountKeyPairs,accountNames,localId,i)

            }
        }
        sem.Acquire(ctx,conf.ThreadLimit)
        sem.Release(conf.ThreadLimit)
        if !buildState.ErrorFree(){
            return nil, buildState.GetError()
        }
    }
    password := clientPasswords[servers[0].Ips[0]]
    passwordNormal := clientPasswords[servers[0].Ips[1]]
    buildState.IncrementBuildProgress() 
    {
        buildState.SetBuildStage("Building genesis block")
        node := 0
        for i, server := range servers {
            sem.Acquire(ctx,1)
            go func(i int, ips []string){
                defer sem.Release(1)
                err := clients[i].Scp("./genesis.json", "/home/appo/genesis.json")
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }
                err = clients[i].Scp("./config.ini", "/home/appo/config.ini")
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }

                for j := 0; j < len(ips); j++ {
                    _,err = clients[i].DockerExec(j,"mkdir /datadir/")
                    if err != nil {
                        log.Println(err)
                        buildState.ReportError(err)
                        return
                    }
                    _,err = clients[i].FastMultiRun(
                                    fmt.Sprintf("docker cp /home/appo/genesis.json whiteblock-node%d:/datadir/", j),
                                    fmt.Sprintf("docker cp /home/appo/config.ini whiteblock-node%d:/datadir/", j))
                    if err != nil {
                        log.Println(err)
                        buildState.ReportError(err)
                        return
                    }
                    node++
                    buildState.IncrementBuildProgress()
                }
                _,err = clients[i].Run("rm /home/appo/genesis.json")
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }
                _,err = clients[i].Run("rm /home/appo/config.ini")
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }
            }(i,server.Ips)
        }
        sem.Acquire(ctx,conf.ThreadLimit)
        sem.Release(conf.ThreadLimit)
        if !buildState.ErrorFree(){
            return nil, buildState.GetError()
        }
    }
    defer func(){
        util.Rm("./genesis.json")
        util.Rm("./config.ini")
    }()
    buildState.IncrementBuildProgress()
    /**Step 2d**/
    buildState.SetBuildStage("Starting EOS BIOS boot sequence")
    {
        
        res,err := clients[0].KeepTryDockerExec(0,fmt.Sprintf("cleos wallet import --private-key %s", 
                keyPairs[masterIP].PrivateKey))
        fmt.Println(res)
        if err != nil {
            log.Println(err)
            return nil,err
        }
        
        err = clients[0].DockerExecdLog(0,
                    fmt.Sprintf(`nodeos -e -p eosio --genesis-json /datadir/genesis.json --config-dir /datadir --data-dir /datadir %s %s`,
                        eos_getKeyPairFlag(keyPairs[masterIP]),
                        eos_getPTPFlags(servers, 0)))
        fmt.Println(res)
        if err != nil {
            log.Println(err)
            return nil,err
        }
        
    }
    
    buildState.IncrementBuildProgress() 
    /**Step 3**/
    {
        clients[0].Run(fmt.Sprintf("docker exec whiteblock-node0 cleos -u http://%s:8889 wallet unlock --password %s",
            masterIP, password))//Can fail

        for _, account := range contractAccounts {
            sem.Acquire(ctx,1)
            go func(masterIP string,account string,masterKeyPair util.KeyPair,contractKeyPair util.KeyPair){
                defer sem.Release(1)
                
                
                _,err = clients[0].KeepTryDockerExec(0,fmt.Sprintf("cleos wallet import --private-key %s", 
                            contractKeyPair.PrivateKey))
                if err != nil {
                    buildState.ReportError(err)
                    log.Println(err)
                    return
                }
                res,err := clients[0].KeepTryDockerExec(0,fmt.Sprintf("cleos -u http://%s:8889 create account eosio %s %s %s",
                            masterIP, account,masterKeyPair.PublicKey,contractKeyPair.PublicKey))
                if err != nil {
                    buildState.ReportError(err)
                    log.Println(err)
                    log.Println(res)
                    return
                }

                //log.Println("Finished creating account for "+account)
            }(masterIP,account,masterKeyPair,contractKeyPairs[account])

        }
        sem.Acquire(ctx,conf.ThreadLimit)
        sem.Release(conf.ThreadLimit)

        if !buildState.ErrorFree(){
            return nil, buildState.GetError()
        }
        
    }
    buildState.IncrementBuildProgress() 
    /**Steps 4 and 5**/
    {
        contracts := []string{"eosio.token","eosio.msig"}
        clients[0].KeepTryDockerExec(0, fmt.Sprintf("cleos -u http://%s:8889 wallet unlock --password %s",masterIP, password))//ign

        for _, contract := range contracts {
            
            _,err = clients[0].KeepTryDockerExec(0, fmt.Sprintf("cleos -u http://%s:8889 set contract %s /opt/eosio/contracts/%s",
                masterIP, contract, contract))
            if err != nil {
                log.Println(err)
                return nil,err
            }
        }
    }
    buildState.SetBuildStage("Creating the tokens")
    buildState.IncrementBuildProgress() 
    /**Step 6**/

    res,err := clients[0].KeepTryDockerExec(0, 
            fmt.Sprintf("cleos -u http://%s:8889 push action eosio.token create '[ \"eosio\", \"10000000000.0000 SYS\" ]' -p eosio.token@active",
                        masterIP))
    fmt.Println(res)
    if err != nil {
        log.Println(err)
        return nil,err
    }
    

    res,err = clients[0].KeepTryDockerExec(0, 
        fmt.Sprintf("cleos -u http://%s:8889 push action eosio.token issue '[ \"eosio\", \"1000000000.0000 SYS\", \"memo\" ]' -p eosio@active",
        masterIP))
    fmt.Println(res)
    if err != nil{
        log.Println(err)
        return nil,err
    }

    
    buildState.SetBuildStage("Setting up the system contract")
    clients[0].DockerExec(0,fmt.Sprintf("cleos -u http://%s:8889 wallet unlock --password %s",masterIP, password))//Ignore fail

    buildState.IncrementBuildProgress() 
    /**Step 7**/

    res, err = clients[0].KeepTryDockerExec(0, 
        fmt.Sprintf("cleos -u http://%s:8889 set contract -x 1000 eosio /opt/eosio/contracts/eosio.system",masterIP))
    
    fmt.Println(res)
    if(err != nil){
        log.Println(err)  
        return nil,err
    }
  
    
    buildState.IncrementBuildProgress() 
    /**Step 8**/

    
    res,err = clients[0].KeepTryDockerExec(0,
        fmt.Sprintf(`cleos -u http://%s:8889 push action eosio setpriv '["eosio.msig", 1]' -p eosio@active`,
            masterIP))
    
    fmt.Println(res)
    if err != nil{
        log.Println(err)
        return nil,err
    }

    res,err = clients[0].KeepTryDockerExec(0,
        fmt.Sprintf(`cleos -u http://%s:8889 push action eosio init '["0", "4,SYS"]' -p eosio@active`,masterIP))
    fmt.Println(res)
    if err != nil{
        log.Println(err)
        return nil,err
    }
    buildState.SetBuildStage("Creating the block producers")
    buildState.IncrementBuildProgress() 

    /**Step 10a**/
    {
        node := 0
        for _, server := range servers {
            for _, ip := range server.Ips {
                if node == 0 {
                    node++
                    continue
                }
                sem.Acquire(ctx,1)
                go func(masterServerIP string, masterKeyPair util.KeyPair, keyPair util.KeyPair,node int){
                    defer sem.Release(1)
                    if node > int(eosconf.BlockProducers) {
                        buildState.IncrementBuildProgress() 
                        return
                    }
                    clients[0].DockerExec(0,fmt.Sprintf("cleos wallet import --private-key %s",keyPair.PrivateKey))//ignore return
                    

                    res,err := clients[0].KeepTryDockerExec(0,
                        fmt.Sprintf(`cleos -u http://%s:8889 system newaccount eosio --transfer %s %s %s --stake-net "%d SYS" --stake-cpu "%d SYS" --buy-ram-kbytes %d`,
                                    masterIP,
                                    eos_getProducerName(node),
                                    masterKeyPair.PublicKey,
                                    keyPair.PublicKey,
                                    eosconf.BpNetStake,
                                    eosconf.BpCpuStake,
                                    eosconf.BpRam))
                    if err != nil {
                        log.Println(res)
                        log.Println(err)
                        buildState.ReportError(err)
                        return
                    }
                    
                    _,err = clients[0].KeepTryDockerExec(0,fmt.Sprintf(`cleos -u http://%s:8889 transfer eosio %s "%d SYS"`,
                                    masterIP,
                                    eos_getProducerName(node),
                                    eosconf.BpFunds))
                    if err != nil {
                        log.Println(err)
                        buildState.ReportError(err)
                        return
                    }
                    
                    buildState.IncrementBuildProgress() 
                }(masterServerIP,masterKeyPair,keyPairs[ip],node)
                node++
            }
        }
        sem.Acquire(ctx,conf.ThreadLimit)
        sem.Release(conf.ThreadLimit)
        if !buildState.ErrorFree() {
            return nil,buildState.GetError()
        }
    }
    buildState.IncrementBuildProgress() 
    buildState.SetBuildStage("Starting up the candidate block producers")
    /**Step 11c**/
    {
        node := 0
        for i, server := range servers {
            for j, ip := range server.Ips {
                
                if node == 0 {
                    node++
                    continue
                }
                sem.Acquire(ctx,1)

                go func(server int,servers []db.Server,node int,j int,kp util.KeyPair){
                    defer sem.Release(1)
                    clients[server].DockerExec(j,"mkdir -p /datadir/blocks")

                    p2pFlags := eos_getPTPFlags(servers,node)
                    prodFlags := ""

                    if node <= int(eosconf.BlockProducers) {
                        prodFlags = " -p "+eos_getProducerName(node)+" "
                    }

                    err := clients[server].DockerExecdLog(j,
                                    fmt.Sprintf(`nodeos --genesis-json /datadir/genesis.json --config-dir /datadir --data-dir /datadir %s %s %s`,
                                        prodFlags,
                                        eos_getKeyPairFlag(kp),
                                        p2pFlags))
                    //fmt.Println(res)
                    if err != nil{
                        log.Println(err)
                        buildState.ReportError(err)
                        return
                    }
                    
                    
                }(i,servers,node,j,keyPairs[ip])
                node++
            }
        }
        sem.Acquire(ctx,conf.ThreadLimit)
        sem.Release(conf.ThreadLimit)
        if !buildState.ErrorFree(){
            return nil,buildState.GetError()
        }
    }
    buildState.IncrementBuildProgress() 
    /**Step 11a**/
    {
        node := 0
        for _, server := range servers {
            for _, ip := range server.Ips {
                
                if node == 0 {
                    node++
                    continue
                }else if node > int(eosconf.BlockProducers) {
                    break
                }

                if node % 5 == 0{
                    clients[0].DockerExec(0,fmt.Sprintf("cleos -u http://%s:8889 wallet unlock --password %s",
                        masterIP, password))//ignore
                }

                
                res,err = clients[0].KeepTryDockerExec(0,
                            fmt.Sprintf("cleos --wallet-url http://%s:8900 -u http://%s:8889 system regproducer %s %s https://whiteblock.io/%s",
                                masterIP,
                                masterIP,
                                eos_getProducerName(node),
                                keyPairs[ip].PublicKey,
                                keyPairs[ip].PublicKey))
                fmt.Println(res)
                if err != nil{
                    log.Println(err)
                    return nil,err
                }                
                node++
            }
        }
    }
    buildState.IncrementBuildProgress() 
    /**Step 11b**/
    res,err = clients[0].DockerExec(0,fmt.Sprintf("cleos -u http://%s:8889 system listproducers",masterIP))
    fmt.Println(res)
    if err != nil{
        log.Println(err)
        return nil,err
    }
    fmt.Println(res)
    /**Create normal user accounts**/
    buildState.SetBuildStage("Creating funded accounts")
    for _, name := range accountNames {
        sem.Acquire(ctx,1)
        go func(masterServerIP string,name string,masterKeyPair util.KeyPair,accountKeyPair util.KeyPair){
            defer sem.Release(1)
            res,err := clients[0].KeepTryDockerExec(0,
                fmt.Sprintf(`cleos -u http://%s:8889 system newaccount eosio --transfer %s %s %s --stake-net "%d SYS" --stake-cpu "%d SYS" --buy-ram-kbytes %d`,
                            masterIP,
                            name,
                            masterKeyPair.PublicKey,
                            accountKeyPair.PublicKey,
                            eosconf.AccountNetStake,
                            eosconf.AccountCpuStake,
                            eosconf.AccountRam))
            fmt.Println(res)
            if err != nil{
                log.Println(err)
                buildState.ReportError(err)
                return
            }
        
            res,err = clients[0].KeepTryDockerExec(0,
                    fmt.Sprintf(`cleos -u http://%s:8889 transfer eosio %s "%d SYS"`,
                                masterIP,
                                name,
                                eosconf.AccountFunds))
            fmt.Println(res)
            if err != nil{
                log.Println(err)
                buildState.ReportError(err)
                return
            }
            buildState.IncrementBuildProgress()

        }(masterServerIP,name,masterKeyPair,accountKeyPairs[name])
    }
    sem.Acquire(ctx,conf.ThreadLimit)
    sem.Release(conf.ThreadLimit)
    if !buildState.ErrorFree() {
        return nil, buildState.GetError()
    }
    
    buildState.IncrementBuildProgress() 
    buildState.SetBuildStage("Voting in block producers")
    /**Vote in block producers**/
    {   
        node := 0
        for _, server := range servers {
            for range server.Ips {          
                node++
            }
        }
        if(node > int(eosconf.BlockProducers)){
            node = int(eosconf.BlockProducers)
        }
        clients[0].DockerExec(1, fmt.Sprintf("cleos -u http://%s:8889 wallet unlock --password %s",
                masterIP, passwordNormal))
        n := 0
        for _, name := range accountNames {
            prod := 0
            fmt.Printf("name=%sn=%d\n",name,n)
            if n > 0 {
                prod = rand.Intn(100) % n
            } 
        
            prod = prod % (node - 1)
            prod += 1
            sem.Acquire(ctx,1)
            go func(masterServerIP string,masterIP string,name string,prod int){
                defer sem.Release(1)
                
                    res,err := clients[0].KeepTryDockerExec(1,
                            fmt.Sprintf("cleos -u http://%s:8889 system voteproducer prods %s %s",
                                        masterIP,
                                        name,
                                        eos_getProducerName(prod)))
                        fmt.Println(res)
                        if err != nil{
                            log.Println(err)
                            buildState.ReportError(err)
                            return
                        }
                    
                buildState.IncrementBuildProgress()
            }(masterServerIP,masterIP,name,prod)
            n++;
        }
        sem.Acquire(ctx,conf.ThreadLimit)
        sem.Release(conf.ThreadLimit)
        if !buildState.ErrorFree() {
            return nil, buildState.GetError()
        }
    }
    buildState.IncrementBuildProgress()
    buildState.SetBuildStage("Initializing EOSIO") 
    /**Step 12**/
    
    _,err = clients[0].KeepTryDockerExec(0,
            fmt.Sprintf(
                `cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio.prods", "permission": "active"}}]}}' -p eosio@owner`,
                masterIP))
    if err != nil{
        log.Println(err)
        return nil,err
    }

    
    _,err = clients[0].KeepTryDockerExec(0,
            fmt.Sprintf(
                `cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio.prods", "permission": "active"}}]}}' -p eosio@active`,
                masterIP))
    if err != nil {
        log.Println(err)
        return nil,err
    }
    
    
    _,err = clients[0].KeepTryDockerExec(0,
        fmt.Sprintf(
            `cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.bpay", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.bpay@owner`,
            masterIP))

    if err != nil {
        log.Println(err)
        return nil,err
    }
    
    _,err = clients[0].KeepTryDockerExec(0,
        fmt.Sprintf(
            `cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.bpay", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.bpay@active`,
            masterIP))

    if err != nil{
        log.Println(err)
        return nil,err
    }

    _,err = clients[0].KeepTryDockerExec(0,
            fmt.Sprintf(
                `cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.msig", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.msig@owner`,
                masterIP))

    if err != nil {
        log.Println(err)
        return nil,err
    }

    _,err = clients[0].KeepTryDockerExec(0,
            fmt.Sprintf(
                `cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.msig", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.msig@active`,
                masterIP))

    if err != nil {
        log.Println(err)
        return nil,err
    }

    
    _,err = clients[0].KeepTryDockerExec(0,
            fmt.Sprintf(
                `cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.names", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.names@owner`,
                masterIP))

    if err != nil {
        log.Println(err)
        return nil,err
    }

    _,err = clients[0].KeepTryDockerExec(0,fmt.Sprintf(
                `cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.names", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.names@active`,
                masterIP))

    if err != nil {
        log.Println(err)
        return nil,err
    }

    
    _,err = clients[0].KeepTryDockerExec(0,fmt.Sprintf(
                `cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.ram", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.ram@owner`,
                masterIP))

    if err != nil {
        log.Println(err)
        return nil,err
    }

    _,err = clients[0].KeepTryDockerExec(0,fmt.Sprintf(
                `cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.ram", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.ram@active`,
                masterIP))

    if err != nil {
        log.Println(err)
        return nil,err
    }
    _,err = clients[0].KeepTryDockerExec(0,fmt.Sprintf(
                `cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.ramfee", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.ramfee@owner`,
                masterIP))
    if err != nil {
        log.Println(err)
        return nil,err
    }

    
    _,err = clients[0].KeepTryDockerExec(0,fmt.Sprintf(
                `cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.ramfee", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.ramfee@active`,
                masterIP))
    if err != nil {
        log.Println(err)
        return nil,err
    }

    out := []string{}

    for _, server := range servers {
        for _, ip := range server.Ips {
            out = append(out,clientPasswords[ip])
        }
    }
    buildState.IncrementBuildProgress() 
    return out,nil
}

func Add(data map[string]interface{},nodes int,servers []db.Server,clients []*util.SshClient,
         newNodes map[int][]string,buildState *state.BuildState) ([]string,error) {
    return nil,nil
}

/*func eos_getKeyPair(serverIP string) (util.KeyPair,error){
    data,err := util.SshExec(serverIP, "docker exec whiteblock-node0 cleos create key --to-console | awk '{print $3}'")
    if err != nil {
        return util.KeyPair{},err
    }
    //fmt.Printf("RAW KEY DATA%s\n", data)
    keyPair := strings.Split(data, "\n")
    if(len(data) < 10){
        return util.KeyPair{},errors.New(fmt.Sprintf("Unexpected create key output %s\n",keyPair))
        panic(1)
    }
    return util.KeyPair{PrivateKey: keyPair[0], PublicKey: keyPair[1]},nil
}*/

/**
func eos_getKeyPairs(servers []db.Server,clients []*util.SshClient) (map[string]util.KeyPair,error) {
    keyPairs := make(map[string]util.KeyPair)
    //Get the key pairs for each nodeos account
    
    var wg sync.WaitGroup
    var mutex = &sync.Mutex{}

    for i, server := range servers {
        wg.Add(1)
        go func(server int,ips []string){
            defer wg.Done()
            for _, ip := range ips {
                data,err := clients[server].DockerExec(0,"cleos create key --to-console | awk '{print $3}'")
                if err != nil {
                    state.ReportError(err)
                    return
                }
                //fmt.Printf("RAW KEY DATA%s\n", data)
                keyPair := strings.Split(data, "\n")
                if(len(data) < 10){
                    fmt.Printf("Unexpected create key output %s\n",keyPair)
                    panic(1)
                }
                    
                mutex.Lock()
                keyPairs[ip] = util.KeyPair{PrivateKey: keyPair[0], PublicKey: keyPair[1]}
                mutex.Unlock()
            }
        }(i,server.Ips)
    }
    wg.Wait()
    if !state.ErrorFree() {
        return nil, state.GetError()
    }
    return keyPairs,nil
}


func eos_getContractKeyPairs(servers []db.Server,contractAccounts []string) (map[string]util.KeyPair,error) {

    keyPairs := make(map[string]util.KeyPair)
    server := servers[0]
    var err error

    for _,contractAccount := range contractAccounts {
        
        keyPairs[contractAccount],err = eos_getKeyPair(server.Addr)
        if err != nil {
            return keyPairs,err
        }
    }
    return keyPairs,nil
}

func eos_getUserAccountKeyPairs(client *util.SshClient,accountNames []string) (map[string]util.KeyPair,error) {

    keyPairs := make(map[string]util.KeyPair)
    sem := semaphore.NewWeighted(conf.ThreadLimit)
    var mutex = &sync.Mutex{}
    ctx := context.TODO()

    for _,name := range accountNames {
        sem.Acquire(ctx,1)
        go func(name string){
            defer sem.Release(1)
            data,err := client.DockerExec(0, "cleos create key --to-console | awk '{print $3}'")
            if err != nil{
                state.ReportError(err)
                return
            }
            //fmt.Printf("RAW KEY DATA%s\n", data)
            keyPair := strings.Split(data, "\n")
            if(len(data) < 10){
                fmt.Printf("Unexpected create key output %s\n",keyPair)
                panic(1)
            }
            mutex.Lock()
            keyPairs[name] = util.KeyPair{PrivateKey: keyPair[0], PublicKey: keyPair[1]}
            mutex.Unlock()
        }(name)
    }
    sem.Acquire(ctx,conf.ThreadLimit)
    sem.Release(conf.ThreadLimit)
    if !state.ErrorFree(){
        return keyPairs,state.GetError()
    }
    return keyPairs,nil
}
*/

func eos_createWallet(client *util.SshClient, node int) (string,error) {
    data,err := client.DockerExec(node,"cleos wallet create --to-console | tail -n 1")
    if err != nil{
        return "",err
    }
    //fmt.Printf("CREATE WALLET DATA %s\n",data)
    offset := 0
    for data[len(data) - (offset + 1)] != '"' {
        offset++
    }
    offset++
    data = data[1 : len(data) - offset]
    fmt.Printf("CREATE WALLET DATA %s\n",data)
    return data,nil
}

func eos_getKeyPairFlag(keyPair util.KeyPair) string {
    return fmt.Sprintf("--signature-provider %s=KEY:%s", keyPair.PublicKey, keyPair.PrivateKey)
}

func eos_getProducerName(num int) string {
    if num == 0 {
        return "eosio"
    }
    out := ""

    for i := num; i > 0; i = (i - (i % 4)) / 4{
        place := i % 4
        place++
        out = fmt.Sprintf("%d%s",place,out)//I hate this
    }
    for i := len(out); i < 5; i++ {
        out = "x"+out
    }

    return "prod"+out
}

func eos_getRegularName(num int) string {

    out := ""
    //num -= blockProducers

    for i := num; i > 0; i = (i - (i % 4)) / 4{
        place := i % 4
        place++
        out = fmt.Sprintf("%d%s",place,out)//I hate this
    }
    for i := len(out); i < 8; i++ {
        out = "x"+out
    }

    return "user"+out
}


func eos_getPTPFlags(servers []db.Server, exclude int) string {
    flags := ""
    node := 0
    for _, server := range servers {
        for _, ip := range server.Ips {
            if(node == exclude){
                node++
                continue
            }
            flags += fmt.Sprintf("--p2p-peer-address %s:8999 ", ip)

        }
    }
    return flags
}
