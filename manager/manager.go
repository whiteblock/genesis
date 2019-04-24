/*
   Contains functions for managing the testnets.
   Handles creating test nets, adding/removing nodes from testnets, and keeps track of the
   ssh clients for each server
*/
package manager

import (
	artemis "../blockchains/artemis"
	beam "../blockchains/beam"
	cosmos "../blockchains/cosmos"
	eos "../blockchains/eos"
	geth "../blockchains/geth"
	helpers "../blockchains/helpers"
	pantheon "../blockchains/pantheon"
	parity "../blockchains/parity"
	rchain "../blockchains/rchain"
	sys "../blockchains/syscoin"
	tendermint "../blockchains/tendermint"
	"encoding/json"
	"fmt"
	"log"
	"time"

	db "../db"
	deploy "../deploy"
	testnet "../testnet"
	util "../util"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

// AddTestNet implements the build command. All blockchains Build command must be
// implemented here, other it will not be called during the build process.
func AddTestNet(details *db.DeploymentDetails, testNetId string) error {
	if details.Servers == nil || len(details.Servers) == 0 {
		err := fmt.Errorf("Missing servers")
		log.Println(err)
		return err
	}
	//STEP 1: SETUP THE TESTNET
	tn, err := testnet.NewTestNet(*details, testNetId)
	if err != nil {
		log.Println(err)
		return err
	}
	buildState := tn.BuildState
	defer buildState.DoneBuilding()
	defer tn.FinishedBuilding()

	//STEP 0: VALIDATE
	err = validate(details)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}

	buildState.Async(func() {
		declareTestnet(testNetId, details)
	})

	//STEP 3: GET THE SERVICES
	services := GetServices(details.Blockchain)

	//STEP 4: BUILD OUT THE DOCKER CONTAINERS AND THE NETWORK

	err = deploy.Build(tn, services)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}
	fmt.Println("Built the docker containers")

	var labels []string = nil

	switch details.Blockchain {
	case "eos":
		labels, err = eos.Build(tn)
	case "ethereum":
		fallthrough
	case "geth":
		labels, err = geth.Build(tn)
	case "parity":
		labels, err = parity.Build(tn)
	case "artemis":
		labels, err = artemis.Build(tn)
	case "pantheon":
		labels, err = pantheon.Build(tn)
	case "syscoin":
		labels, err = sys.RegTest(tn)
	case "rchain":
		labels, err = rchain.Build(tn)
	case "beam":
		labels, err = beam.Build(tn)
	case "tendermint":
		labels, err = tendermint.Build(tn)
	case "cosmos":
		labels, err = cosmos.Build(tn)
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
		Nodes: details.Nodes, Image: details.Images[0], //fix
		Ts: time.Now().Unix()})

	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}

	err = db.InsertBuild(*details, testNetId)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}
	err = tn.StoreNodes(labels)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}
	return nil
}

func declareTestnet(testnetId string, details *db.DeploymentDetails) error {

	data := map[string]interface{}{
		"id":        testnetId,
		"kind":      details.Blockchain,
		"num_nodes": details.Nodes,
		"image":     details.Images[0],
	}
	rawData, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = util.JwtHttpRequest("POST", "https://api.whiteblock.io/testnets", details.GetJwt(), string(rawData))
	return err
}

func DeleteTestNet(testnetID string) error {
	tn, err := testnet.RestoreTestNet(testnetID)
	if err != nil {
		log.Println(err)
		return err
	}

	return deploy.Destroy(tn)
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
	return helpers.GetStaticBlockchainConfig(blockchain, "params.json")
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
	return helpers.GetStaticBlockchainConfig(blockchain, "defaults.json")
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
