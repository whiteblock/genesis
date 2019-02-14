/*
    Contains functions for managing the testnets. 
    Handles creating test nets, adding/removing nodes from testnets, and keeps track of the
    ssh clients for each server
*/
package testnet

import (
    "errors"
    "fmt"
    "log"

    beam "../blockchains/beam"
    eos "../blockchains/eos"
    eth "../blockchains/ethereum"
    rchain "../blockchains/rchain"
    sys "../blockchains/syscoin"
    tendermint "../blockchains/tendermint"
    cosmos "../blockchains/cosmos"

    db "../db"
    deploy "../deploy"
    state "../state"
    status "../status"
    util "../util"
)

var conf *util.Config

func init() {
    conf = util.GetConfig()
}


// AddTestNet implements the build command. All blockchains Build command must be 
// implemented here, other it will not be called during the build process. 
func AddTestNet(details db.DeploymentDetails) error {
    buildState := state.GetBuildStateByServerId(details.Servers[0])
    buildState.SetDeploySteps(3*details.Nodes + 2 )
    defer buildState.DoneBuilding()
    //STEP 0: VALIDATE
    err := details.Resources.ValidateAndSetDefaults()
    if err != nil {
        log.Println(err.Error())
        buildState.ReportError(err)
        return err
    }
    if details.Nodes > conf.MaxNodes {
        buildState.ReportError(errors.New("Too many nodes"))
        return errors.New("Too many nodes")
    }
    //STEP 1: FETCH THE SERVERS
    servers, err := db.GetServers(details.Servers)
    if err != nil {
        log.Println(err.Error())
        buildState.ReportError(err)
        return err
    }
    fmt.Println("Got the Servers")

    //STEP 2: OPEN UP THE RELEVANT SSH CONNECTIONS
    clients,err :=  GetClients(details.Servers) 
    if err != nil {
        log.Println(err)
        buildState.ReportError(err)
        return err
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
        case "beam":
            services = beam.GetServices()
        case "tendermint":
            services = tendermint.GetServices()
        case "cosmos":
            services = cosmos.GetServices()
    }

    //STEP 4: BUILD OUT THE DOCKER CONTAINERS AND THE NETWORK
    config := deploy.Config{Nodes: details.Nodes, Image: details.Image, Servers: details.Servers}
    fmt.Printf("Created the build configuration : %+v \n",config)

    newServerData,err := deploy.Build(&config,servers,details.Resources,clients,services,buildState) //TODO: Restructure distribution of nodes over servers
    if err != nil {
        log.Println(err)
        buildState.ReportError(err)
        return err
    }
    fmt.Println("Built the docker containers")

    var labels []string = nil

    switch(details.Blockchain){
        case "eos":
            labels,err = eos.Build(details.Params,details.Nodes,newServerData,clients,buildState);
        case "ethereum":
            labels,err = eth.Build(details.Params,details.Nodes,newServerData,clients,buildState)
        case "syscoin":
            labels,err = sys.RegTest(details.Params,details.Nodes,newServerData,clients,buildState)
        case "rchain":
            labels,err = rchain.Build(details.Params,details.Nodes,newServerData,clients,buildState)
        case "beam":
            labels, err = beam.Build(details.Params, details.Nodes, newServerData, clients,buildState)
        case "tendermint":
            labels, err = tendermint.Build(details.Params, details.Nodes, newServerData, clients,buildState)
        case "cosmos":
            labels, err = cosmos.Build(details.Params, details.Nodes, newServerData, clients,buildState)
        case "generic":
            log.Println("Built in generic mode")
        default:
            buildState.ReportError(errors.New("Unknown blockchain"))
            return errors.New("Unknown blockchain")
    }
    if err != nil {
        buildState.ReportError(err)
        log.Println(err)
        return err
    }
    testNetId,err := db.InsertTestNet(db.TestNet{Id: -1, Blockchain: details.Blockchain, Nodes: details.Nodes, Image: details.Image})
    if err != nil{
        log.Println(err)
        buildState.ReportError(err);
        return err
    }
    err = db.InsertBuild(details,testNetId)
    if err != nil{
        log.Println(err)
        buildState.ReportError(err);
        return err
    }
    i := 0
    for _, server := range newServerData {
        err = db.UpdateServerNodes(server.Id,0)
        if err != nil{
            log.Println(err)
            panic(err)
        }
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

/*
    GetNextTestNetId gets the next testnet id. Used for
    getting the id of a testnet that is in progress of being built
 */
func GetNextTestNetId() (string, error) {
    highestId, err := status.GetLastTestNetId()
    return fmt.Sprintf("%d", highestId+1), err
}

/*
    GetParams fetches the name and type of each availible
    blockchain specific parameter for the given blockchain. 
    Ensure that the blockchain you have implemented is included 
    in the switch statement.
 */
func GetParams(blockchain string) string {
    switch blockchain {
    case "ethereum":
        return eth.GetParams()
    case "syscoin":
        return sys.GetParams()
    case "eos":
        return eos.GetParams()
    case "rchain":
        return rchain.GetParams()
    case "beam":
        return beam.GetParams()
    case "tendermint":
        return tendermint.GetParams()
    case "cosmos":
        return cosmos.GetParams()
    default:
        return "[]"
    }
}

/*
    GetDefaults gets the default parameters for a blockchain. Ensure that 
    the blockchain you have implemented is included in the switch
    statement.
 */
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
    case "beam":
        return beam.GetDefaults()
    case "tendermint":
        return tendermint.GetDefaults()
    case "cosmos":
        return cosmos.GetDefaults()
    default:
        return "{}"
    }
}
