package pantheon

import (
    "fmt"
    "log"
    "sync"
    "context"
    "golang.org/x/sync/semaphore"
    "github.com/Whiteblock/mustache"
    db "../../db"
    util "../../util"
    state "../../state"
)

var conf *util.Config

func init() {
    conf = util.GetConfig()
}

func Build(details db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
    buildState *state.BuildState) ([]string, error) {

    sem := semaphore.NewWeighted(conf.ThreadLimit)
    ctx := context.TODO()
    mux := sync.Mutex{}

    panconf, err := NewConf(details.Params)
    if err != nil {
        log.Println(err)
        return nil, err
    }

    buildState.SetBuildSteps(6 * details.Nodes + 2)
    buildState.IncrementBuildProgress()

    addresses := make([]string,details.Nodes)
    pubKeys := make([]string,details.Nodes)
    privKeys := make([]string,details.Nodes)

    buildState.SetBuildStage("Setting Up Accounts")
    node := 0
    for i, server := range servers {
        for localId, _ := range server.Ips {
            sem.Acquire(ctx,1)
            go func(i int,localId int,node int){
                defer sem.Release(1)
                res, err := clients[i].DockerExec(localId, "pantheon --data-path=/pantheon/data public-key export-address --to=/pantheon/data/nodeAddress")
                if err != nil {
                    log.Println(err)
                    log.Println(res)
                    buildState.ReportError(err)
                    return
                }
                buildState.IncrementBuildProgress()
                _, err = clients[i].DockerExec(localId, "pantheon --data-path=/pantheon/data public-key export --to=/pantheon/data/publicKey")
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }

                addr, err := clients[i].DockerExec(localId, "cat /pantheon/data/nodeAddress")
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }

                addrs := string(addr[2:])
                
                mux.Lock()
                addresses[node] = addrs
                mux.Unlock()

                key, err := clients[i].DockerExec(localId, "cat /pantheon/data/publicKey")
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }
                buildState.IncrementBuildProgress()
                keys := string(key[2:])

                mux.Lock()
                pubKeys[node] = keys
                mux.Unlock()


                privKey, err := clients[i].DockerExec(localId, "cat /pantheon/data/key")
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }
                mux.Lock()
                privKeys[node] = privKey
                mux.Unlock()

                res, err = clients[i].DockerExec(localId, "bash -c 'echo \"[\\\"" + addrs + "\\\"]\" >> /pantheon/data/toEncode.json'")
                if err != nil {
                    log.Println(err)
                    log.Println(res)
                    buildState.ReportError(err)
                    return
                }

                _, err = clients[i].DockerExec(localId, "mkdir /pantheon/genesis")
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }

                // used for IBFT2 extraData
                _, err = clients[i].DockerExec(localId, "pantheon rlp encode --from=/pantheon/data/toEncode.json --to=/pantheon/rlpEncodedExtraData")
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }
                buildState.IncrementBuildProgress()
            }(i,localId,node)
            node++
        }
    }

    err = sem.Acquire(ctx,conf.ThreadLimit)
    if err != nil{
        log.Println(err)
        return nil,err
    }
    sem.Release(conf.ThreadLimit)

    if !buildState.ErrorFree(){
        return nil,buildState.GetError()
    }

        /* Create Genesis File */
    buildState.SetBuildStage("Generating Genesis File")
    err = createGenesisfile(panconf,details,addresses)
    if err != nil{
        log.Println(err)
        return nil,err
    }
    defer util.Rm("./genesis.json")

    p2pPort := 30303
    enodes := "["
    var enodeAddress string
    for _, server := range servers {
        for i, ip := range server.Ips {
            enodeAddress = fmt.Sprintf("enode://%s@%s:%d",
            pubKeys[i],
            ip,
            p2pPort)
            if i < len(pubKeys)-1 {
                enodes = enodes + "\"" + enodeAddress + "\"" + ","
            } else {
                enodes = enodes + "\"" + enodeAddress + "\""
            }
            buildState.IncrementBuildProgress()
        }
    }
    enodes = enodes + "]"

    /* Create Static Nodes File */
    buildState.SetBuildStage("Setting Up Static Peers")
    buildState.IncrementBuildProgress()
    err = createStaticNodesFile(enodes)
    if err != nil{
        log.Println(err)
        return nil,err
    }
    defer util.Rm("./static-nodes.json")

    /* Copy static-nodes & genesis files to each node */
    buildState.SetBuildStage("Distributing Files")
    for i, server := range servers {
        err = clients[i].Scp("./static-nodes.json", "/home/appo/static-nodes.json")
        if err != nil {
            log.Println(err)
            return nil, err
        }
        defer clients[i].Run("rm /home/appo/static-nodes.json")

        err = clients[i].Scp("./genesis.json", "/home/appo/genesis.json")
        if err != nil {
            log.Println(err)
            return nil, err
        }
        defer clients[i].Run("rm /home/appo/genesis.json")

        for localId, _ := range server.Ips {
            sem.Acquire(ctx,1)
            go func(i int,localId int){
                defer sem.Release(1)
                err := clients[i].DockerCp(localId,"/home/appo/static-nodes.json","/pantheon/data/static-nodes.json")
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }
                err = clients[i].DockerCp(localId,"/home/appo/genesis.json","/pantheon/genesis/genesis.json")
                if err != nil {
                    log.Println(err)
                    buildState.ReportError(err)
                    return
                }
                buildState.IncrementBuildProgress()
            }(i,localId)
        }

    }

    err = sem.Acquire(ctx,conf.ThreadLimit)
    if err != nil{
        log.Println(err)
        return nil,err
    }
    sem.Release(conf.ThreadLimit)

    if !buildState.ErrorFree(){
        return nil,buildState.GetError()
    }

    /* Start the nodes */
    buildState.SetBuildStage("Starting Pantheon")
    httpPort := 8545
    for i, server := range servers {
        for localId, _ := range server.Ips {
            pantheonCmd := fmt.Sprintf(
                `pantheon --data-path /pantheon/data --genesis-file=/pantheon/genesis/genesis.json --rpc-http-enabled --rpc-http-api="ADMIN,CLIQUE,DEBUG,EEA,ETH,IBFT,MINER,NET,WEB3" ` +
                    ` --p2p-port=%d --rpc-http-port=%d --rpc-http-host="0.0.0.0" --host-whitelist=all --rpc-http-cors-origins="*"`,
                p2pPort,
                httpPort,
                )
            err := clients[i].DockerExecdLog(localId, pantheonCmd)
            if err != nil {
                log.Println(err)
                return nil, err
            }
            buildState.IncrementBuildProgress()
        }
    }
    
    return privKeys, nil
}

func createGenesisfile(panconf *PanConf, details db.DeploymentDetails, address []string) error {
    genesis := map[string]interface{}{
        "chainId":              panconf.NetworkId,
        "difficulty":           fmt.Sprintf("0x0%X", panconf.Difficulty),
        "gasLimit":             fmt.Sprintf("0x0%X", panconf.GasLimit),
        "blockPeriodSeconds":   panconf.BlockPeriodSeconds,
        "epoch":                panconf.Epoch,
    }
    alloc := map[string]map[string]string{}
    for _, addr := range address {
        alloc[addr] = map[string]string{
            "balance": panconf.InitBalance,
        }
    }
    extraData := "0x0000000000000000000000000000000000000000000000000000000000000000"
    for _,addr := range address {
        extraData += addr
    }
    extraData += "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
    genesis["extraData"] = extraData
    genesis["alloc"] = alloc
    dat, err := util.GetBlockchainConfig("pantheon", "genesis.json", details.Files)
    if err != nil {
        log.Println(err)
        return err
    }

    data, err := mustache.Render(string(dat), util.ConvertToStringMap(genesis))
    if err != nil {
        log.Println(err)
        return err
    }
    fmt.Println("Writing Genesis File Locally")
    return util.Write("genesis.json", data)

}

func createStaticNodesFile(list string) error {
    return util.Write("static-nodes.json", list)
}