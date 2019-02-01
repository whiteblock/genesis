package testnet

import(
    "errors"
    "log"
    "fmt"
    beam "../blockchains/beam"
    eos "../blockchains/eos"
    eth "../blockchains/ethereum"
    rchain "../blockchains/rchain"
    sys "../blockchains/syscoin"
    db "../db"
    deploy "../deploy"
    state "../state"
    status "../status"
    //util "../util"
)

func AddNodes(details db.DeploymentDetails) error {
    defer state.DoneBuilding()

    //STEP 1: MERGE IN MISSING INFO FROM ORIGINAL BUILD
    prevDetails,err := status.GetLatestBuild()
    if err != nil {
        log.Println(err.Error())
        state.ReportError(err)
        return err
    }
    if details.Servers == nil || len(details.Servers) == 0 {
        details.Servers = prevDetails.Servers
    }

    if len(details.Blockchain) == 0 {
        details.Blockchain = prevDetails.Blockchain
    }

    if len(details.Image) == 0 {
        details.Image = prevDetails.Image
    }

    if details.Params == nil {
        details.Params = prevDetails.Params
    }
    
    //STEP 2: VALIDATE
    err = details.Resources.ValidateAndSetDefaults()
    if err != nil {
        log.Println(err.Error())
        state.ReportError(err)
        return err
    }
    if details.Nodes > conf.MaxNodes {
        state.ReportError(errors.New("Too many nodes"))
        return errors.New("Too many nodes")
    }
    //STEP 3: FETCH THE SERVERS
    servers, err := status.GetLatestServers()
    if err != nil {
        log.Println(err)
        state.ReportError(err)
        return err
    }
    fmt.Println("Got the Servers")

    //STEP 4: OPEN UP THE RELEVANT SSH CONNECTIONS
    clients,err :=  GetClients(details.Servers) 
    if err != nil {
        log.Println(err)
        state.ReportError(err)
        return err
    }

    config := deploy.Config{Nodes: details.Nodes, Image: details.Image, Servers: details.Servers}


    nodes,err := deploy.AddNodes(&config, servers,details.Resources,clients)
    if err != nil {
        log.Println(err)
        state.ReportError(err)
        return err
    }
    var labels []string = nil
    switch(details.Blockchain){
        case "eos":
            labels,err = eos.Add(details.Params,details.Nodes,servers,clients,nodes);
            if err != nil {
                state.ReportError(err)
                log.Println(err)
                return err
            }
        case "ethereum":
            labels,err = eth.Add(details.Params,details.Nodes,servers,clients,nodes)
            if err != nil {
                state.ReportError(err)
                log.Println(err)
                return err
            }
        case "syscoin":
            labels,err = sys.Add(details.Params,details.Nodes,servers,clients,nodes)
            if err != nil {
                state.ReportError(err)
                log.Println(err)
                return err
            }
        case "rchain":
            labels,err = rchain.Add(details.Params,details.Nodes,servers,clients,nodes)
            if err != nil {
                state.ReportError(err)
                log.Println(err)
                return err
            }
        case "beam":
            labels, err = beam.Add(details.Params, details.Nodes, servers, clients,nodes)
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
    
    testNetId,err := status.GetLastTestNetId()
    i := 0
    for serverId,ips := range nodes {
        for _,ip := range ips{
            node := db.Node{Id: -1, TestNetId: testNetId, Server: serverId, LocalId: i, Ip: ip}
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