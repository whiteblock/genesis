/*
   Contains functions for managing the testnets.
   Handles creating test nets, adding/removing nodes from testnets, and keeps track of the
   ssh clients for each server
*/
package testnet

import (
	"fmt"
	"log"
	"time"

	artemis "../blockchains/artemis"
	beam "../blockchains/beam"
	cosmos "../blockchains/cosmos"
	eos "../blockchains/eos"
	geth "../blockchains/geth"
	pantheon "../blockchains/pantheon"
	parity "../blockchains/parity"
	rchain "../blockchains/rchain"
	sys "../blockchains/syscoin"
	tendermint "../blockchains/tendermint"

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
func AddTestNet(details db.DeploymentDetails, testNetId string) error {

	buildState := state.GetBuildStateByServerId(details.Servers[0])
	buildState.SetDeploySteps(3*details.Nodes + 2)
	defer buildState.DoneBuilding()
	//STEP 0: VALIDATE
	for i, res := range details.Resources {
		err := res.ValidateAndSetDefaults()
		if err != nil {
			log.Println(err)
			err = fmt.Errorf("%s. For node %d", err.Error(), i)
			buildState.ReportError(err)
			return err
		}
	}

	if details.Nodes > conf.MaxNodes {
		buildState.ReportError(fmt.Errorf("Too many nodes"))
		return fmt.Errorf("Too many nodes")
	}

	if details.Nodes < 1 {
		buildState.ReportError(fmt.Errorf("You must have atleast 1 node"))
		return fmt.Errorf("You must have atleast 1 node")
	}
	err := util.ValidateCommandLine(details.Image)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}

	err = util.ValidateCommandLine(details.Blockchain)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}

	if len(details.Image) == 0 {
		details.Image = "gcr.io/whiteblock/" + details.Blockchain + ":master"
	}

	//STEP 1: FETCH THE SERVERS
	servers, err := db.GetServers(details.Servers)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}
	fmt.Println("Got the Servers")

	//STEP 2: OPEN UP THE RELEVANT SSH CONNECTIONS
	clients, err := status.GetClients(details.Servers)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}

	//STEP 3: GET THE SERVICES
	services := GetServices(details.Blockchain)

	//STEP 4: BUILD OUT THE DOCKER CONTAINERS AND THE NETWORK

	newServerData, err := deploy.Build(&details, servers, clients, services, buildState) //TODO: Restructure distribution of nodes over servers
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}
	fmt.Println("Built the docker containers")

	var labels []string = nil

	switch details.Blockchain {
	case "eos":
		labels, err = eos.Build(details, newServerData, clients, buildState)
	case "ethereum":
		fallthrough
	case "geth":
		labels, err = geth.Build(details, newServerData, clients, buildState)
	case "parity":
		labels, err = parity.Build(details, newServerData, clients, buildState)
	case "artemis":
		labels, err = artemis.Build(details, newServerData, clients, buildState)
	case "pantheon":
		labels, err = pantheon.Build(details, newServerData, clients, buildState)
	case "syscoin":
		labels, err = sys.RegTest(details, newServerData, clients, buildState)
	case "rchain":
		labels, err = rchain.Build(details, newServerData, clients, buildState)
	case "beam":
		labels, err = beam.Build(details, newServerData, clients, buildState)
	case "tendermint":
		labels, err = tendermint.Build(details, newServerData, clients, buildState)
	case "cosmos":
		labels, err = cosmos.Build(details, newServerData, clients, buildState)
	case "generic":
		log.Println("Built in generic mode")
	default:
		buildState.ReportError(fmt.Errorf("Unknown blockchain"))
		return fmt.Errorf("Unknown blockchain")
	}
	if err != nil {
		buildState.ReportError(err)
		log.Println(err)
		return err
	}
	err = db.InsertTestNet(db.TestNet{
		Id: testNetId, Blockchain: details.Blockchain,
		Nodes: details.Nodes, Image: details.Image,
		Ts: time.Now().Unix()})
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}
	err = db.InsertBuild(details, testNetId)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}
	i := 0
	for _, server := range newServerData {
		err = db.UpdateServerNodes(server.Id, 0)
		if err != nil {
			log.Println(err)
			panic(err)
		}
		for _, ip := range server.Ips {
			id, err := util.GetUUIDString()
			if err != nil {
				log.Println(err)
				buildState.ReportError(err)
				return err
			}
			node := db.Node{Id: id, TestNetId: testNetId, Server: server.Id, LocalId: i, Ip: ip}
			if labels != nil {
				node.Label = labels[i]
			}
			_, err = db.InsertNode(node)
			if err != nil {
				log.Println(err)
			}
			i++
		}
	}
	return nil
}

func DeleteTestNet(testnetId string) error {
	details, err := db.GetBuildByTestnet(testnetId)
	if err != nil {
		log.Println(err)
		return err
	}
	err = state.AcquireBuilding(details.Servers, testnetId)
	if err != nil {
		log.Println(err)
		return err
	}
	buildState := state.GetBuildStateByServerId(details.Servers[0])
	defer buildState.DoneBuilding()

	clients, err := status.GetClients(details.Servers)
	if err != nil {
		log.Println(err)
		return err
	}
	return deploy.Destroy(&details, clients)
}

/*
   GetParams fetches the name and type of each availible
   blockchain specific parameter for the given blockchain.
   Ensure that the blockchain you have implemented is included
   in the switch statement.
*/
func GetParams(blockchain string) ([]byte, error) {
	if blockchain == "ethereum" {
		return GetParams("geth")
	}
	return util.GetBlockchainConfig(blockchain, "params.json", nil)
}

/*
   GetDefaults gets the default parameters for a blockchain. Ensure that
   the blockchain you have implemented is included in the switch
   statement.
*/
func GetDefaults(blockchain string) ([]byte, error) {
	if blockchain == "ethereum" {
		return GetParams("geth")
	}
	return util.GetBlockchainConfig(blockchain, "defaults.json", nil)
}

func GetServices(blockchain string) []util.Service {
	var services []util.Service
	switch blockchain {
	case "ethereum":
		fallthrough
	case "geth":
		services = geth.GetServices()
	case "parity":
		services = parity.GetServices()
	case "pantheon":
		services = pantheon.GetServices()
	case "artemis":
		services = artemis.GetServices()
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
	return services
}
