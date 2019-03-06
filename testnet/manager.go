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
    "time"
    beam "../blockchains/beam"
    eos "../blockchains/eos"
    geth "../blockchains/geth"
    rchain "../blockchains/rchain"
    sys "../blockchains/syscoin"
    tendermint "../blockchains/tendermint"
    cosmos "../blockchains/cosmos"
    parity "../blockchains/parity"

    db "../db"
    deploy "../deploy"
    state "../state"
    //status "../status"
    util "../util"
)

var conf *util.Config

func init() {
    conf = util.GetConfig()
}


// AddTestNet implements the build command. All blockchains Build command must be 
// implemented here, other it will not be called during the build process. 
func AddTestNet(details db.DeploymentDetails,testNetId string) error {

    buildState := state.GetBuildStateByServerId(details.Servers[0])
    buildState.SetDeploySteps(3*details.Nodes + 2 )
    defer buildState.DoneBuilding()
    //STEP 0: VALIDATE
    for i,res := range details.Resources {
        err := res.ValidateAndSetDefaults()
        if err != nil {
            log.Println(err.Error())
            err = errors.New(fmt.Sprintf("%s. For node %d",err.Error(),i))
            buildState.ReportError(err)
            return err
        }
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
            fallthrough
        case "geth":
            services = geth.GetServices()
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
        case "parity":
            services = parity.GetServices()
    }

    //STEP 4: BUILD OUT THE DOCKER CONTAINERS AND THE NETWORK

    newServerData,err := deploy.Build(&details,servers,clients,services,buildState) //TODO: Restructure distribution of nodes over servers
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
            fallthrough
        case "geth":
            labels,err = geth.Build(details.Params,details.Nodes,newServerData,clients,buildState)
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
        case "parity":
            labels, err = parity.Build(details.Params, details.Nodes, newServerData, clients,buildState)
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
    err = db.InsertTestNet(db.TestNet{
                Id: testNetId, Blockchain: details.Blockchain, 
                Nodes: details.Nodes, Image: details.Image,
                Ts:time.Now().Unix()})
    if err != nil{
        log.Println(err)
        buildState.ReportError(err)
        return err
    }
    err = db.InsertBuild(details,testNetId)
    if err != nil{
        log.Println(err)
        buildState.ReportError(err)
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
            id,err := util.GetUUIDString()
            if err != nil {
                log.Println(err.Error())
                buildState.ReportError(err)
                return err
            }
            node := db.Node{Id:id, TestNetId: testNetId, Server: server.Id, LocalId: i, Ip: ip}
            if labels != nil {
                node.Label = labels[i]
            }
            _,err = db.InsertNode(node)
            if err != nil {
                log.Println(err.Error())
            }
            i++
        }
    }
    return nil
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
            fallthrough
        case "geth":
            return geth.GetParams()
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
        case "parity":
            return parity.GetParams()
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
            fallthrough
        case "geth":
            return geth.GetDefaults()
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
        case "parity":
            return parity.GetDefaults()
        default:
            return "{}"
        }
}
