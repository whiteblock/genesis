package main

import (
    "fmt"
    "log"
    "errors"
    db "./db"
    deploy "./deploy"
    util "./util"
    eos "./blockchains/eos"
    eth "./blockchains/ethereum"
    sys "./blockchains/syscoin"
    rchain "./blockchains/rchain"
    state "./state"
)

type DeploymentDetails struct {
    Servers     []int                   `json:"servers"`
    Blockchain  string                  `json:"blockchain"`
    Nodes       int                     `json:"nodes"`
    Image       string                  `json:"image"`
    Params      map[string]interface{}  `json:"params"`
    Resources   deploy.Resources        `json:"resources"`
}


func AddTestNet(details DeploymentDetails) error {
    defer state.DoneBuilding()

    //STEP 1: FETCH THE SERVERS
    servers, err := db.GetServers(details.Servers)
    if err != nil {
        log.Println(err.Error())
        state.ReportError(err)
        return err
    }
    fmt.Println("Got the Servers")

    //STEP 2: OPEN UP THE RELEVANT SSH CONNECTIONS
    clients := make([]*util.SshClient,len(servers))
    defer func(clients []*util.SshClient){
        for _,client := range clients {
            client.Close()
        }
    }(clients)

    for i,server := range servers {
        clients[i],err = util.NewSshClient(server.Addr)
        if err != nil {
            log.Println(err)
            return err
        }
    }
    
    //STEP 3: GET THE SERVICES
    var services []util.Service
    switch(details.Blockchain){
        case "ethereum":
            services = eth.GetServices()
        case "eos":
            services = eos.GetServices()
        case "syscoin": 
            services = sys.GetServices()
        case "rchain":
            services = rchain.GetServices()
    }

    //STEP 4: BUILD OUT THE DOCKER CONTAINERS AND THE NETWORK
    config := deploy.Config{Nodes: details.Nodes, Image: details.Image, Servers: details.Servers}
    fmt.Printf("Created the build configuration : %+v \n",config)

    newServerData,err := deploy.Build(&config,servers,details.Resources,clients,services) //TODO: Restructure distribution of nodes over servers
    if err != nil {
        log.Println(err)
        state.ReportError(err)
        return err
    }
    fmt.Println("Built the docker containers")

    var labels []string = nil

    switch(details.Blockchain){
        case "eos":
            labels,err = eos.Build(details.Params,details.Nodes,newServerData,clients);
            if err != nil {
                state.ReportError(err)
                log.Println(err)
                return err
            }
        case "ethereum":
            labels,err = eth.Build(details.Params,details.Nodes,newServerData,clients)
            if err != nil {
                state.ReportError(err)
                log.Println(err)
                return err
            }
        case "syscoin":
            labels,err = sys.RegTest(details.Params,details.Nodes,newServerData,clients)
            if err != nil {
                state.ReportError(err)
                log.Println(err)
                return err
            }
        case "rchain":
            labels,err = rchain.Build(details.Params,details.Nodes,newServerData,clients)
            if err != nil {
                state.ReportError(err)
                log.Println(err)
                return err
            }
        case "generic":
            log.Println("Built in generic mode")
        default:
            state.ReportError(errors.New("Unknown blockchain"))
            return errors.New("Unknown blockchain")
    }

    testNetId,err := db.InsertTestNet(db.TestNet{Id: -1, Blockchain: details.Blockchain, Nodes: details.Nodes, Image: details.Image})
    if err != nil{
        log.Println(err)
        state.ReportError(err);
        return err
    }
    i := 0
    for _, server := range newServerData {
        db.UpdateServerNodes(server.Id,0)
        for _, ip := range server.Ips {
            node := db.Node{Id: -1, TestNetId: testNetId, Server: server.Id, LocalId: i, Ip: ip}
            if labels != nil {
                node.Label = labels[i]
            }
            _,err := db.InsertNode(node)
            if err != nil {
                log.Println(err.Error())
            }
            i++
        }
    }
    return nil
}

func GetLastTestNetId() (int,error) {
    testNets,err := db.GetAllTestNets()
    if err != nil{
        log.Println(err)
        return 0,err
    }
    highestId := -1

    for _, testNet := range testNets {
        if testNet.Id > highestId {
            highestId = testNet.Id
        }
    }
    return highestId,nil
}

func GetNextTestNetId() (string,error) {
    highestId,err := GetLastTestNetId()
    return fmt.Sprintf("%d",highestId+1),err
}

func RebuildTestNet(id int) {
    panic("Not Implemented")
}

func RemoveTestNet(id int) error {
    nodes, err := db.GetAllNodesByTestNet(id)
    if err != nil {
        return err
    }
    for _, node := range nodes {
        server, _, err := db.GetServer(node.Server)
        if err != nil {
            log.Println(err)
            return err
        }
        util.SshExec(server.Addr, fmt.Sprintf("~/local_deploy/deploy --kill=%d", node.LocalId))
    }
    return nil
}


func GetParams(blockchain string) string {
    switch blockchain{
        case "ethereum":
            return eth.GetParams()
        case "syscoin":
            return sys.GetParams()
        case "eos":
            return eos.GetParams()
        case "rchain":
            return rchain.GetParams()
        default:
            return "[]"
    }
}

func GetDefaults(blockchain string) string {
    switch blockchain {
        case "ethereum":
            return eth.GetDefaults()
        case "syscoin":
            return sys.GetDefaults()
        case "eos":
            return eos.GetDefaults()
        case "rchain":
            return rchain.GetDefaults()
        default:
            return "{}"
    }
}