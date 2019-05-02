// Package manager contains functions for managing the testnets.
// Handles creating test nets, adding/removing nodes from testnets
package manager

import (
	"../blockchains/helpers"
	"../blockchains/registrar"
	"../db"
	"../deploy"
	"../testnet"
	"../util"
	"encoding/json"
	"fmt"
	"log"
	//Put the relative path to your blockchain library below this line, otherwise it won't be compiled
	_ "../blockchains/artemis"
	_ "../blockchains/beam"
	_ "../blockchains/cosmos"
	_ "../blockchains/eos"
	_ "../blockchains/geth"
	_ "../blockchains/pantheon"
	_ "../blockchains/parity"
	_ "../blockchains/rchain"
	_ "../blockchains/syscoin"
	_ "../blockchains/tendermint"
	
	_ "../sidecars/geth"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

// AddTestNet implements the build command. All blockchains Build command must be
// implemented here, other it will not be called during the build process.
func AddTestNet(details *db.DeploymentDetails, testnetID string) error {
	if details.Servers == nil || len(details.Servers) == 0 {
		err := fmt.Errorf("missing servers")
		log.Println(err)
		return err
	}
	//STEP 1: SETUP THE TESTNET
	tn, err := testnet.NewTestNet(*details, testnetID)
	if err != nil {
		log.Println(err)
		return err
	}
	buildState := tn.BuildState
	defer tn.FinishedBuilding()

	//STEP 0: VALIDATE
	err = validate(details)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}

	buildState.Async(func() {
		declareTestnet(testnetID, details)
	})

	//STEP 3: GET THE SERVICES
	servicesFn, err := registrar.GetServiceFunc(details.Blockchain)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}
	services := servicesFn()
	//STEP 4: BUILD OUT THE DOCKER CONTAINERS AND THE NETWORK

	err = deploy.Build(tn, services)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}
	fmt.Println("Built the docker containers")

	buildFn, err := registrar.GetBuildFunc(details.Blockchain)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}
	labels, err := buildFn(tn)
	if err != nil {
		buildState.ReportError(err)
		log.Println(err)
		return err
	}

	err = db.InsertBuild(*details, testnetID)
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

func declareTestnet(testnetID string, details *db.DeploymentDetails) error {

	data := map[string]interface{}{
		"id":        testnetID,
		"kind":      details.Blockchain,
		"num_nodes": details.Nodes,
		"image":     details.Images[0],
	}
	rawData, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = util.JwtHTTPRequest("POST", "https://api.whiteblock.io/testnets", details.GetJwt(), string(rawData))
	return err
}

// DeleteTestNet destroys all of the nodes of a testnet
func DeleteTestNet(testnetID string) error {
	tn, err := testnet.RestoreTestNet(testnetID)
	if err != nil {
		log.Println(err)
		return err
	}

	return deploy.Destroy(tn)
}

// GetParams fetches the name and type of each availible
// blockchain specific parameter for the given blockchain.
// Ensure that the blockchain you have implemented is included
// in the switch statement.
func GetParams(blockchain string) ([]byte, error) {
	if blockchain == "ethereum" {
		return GetParams("geth")
	}
	return helpers.GetStaticBlockchainConfig(blockchain, "params.json")
}

// GetDefaults gets the default parameters for a blockchain. Ensure that
// the blockchain you have implemented is included in the switch
// statement.
func GetDefaults(blockchain string) ([]byte, error) {
	if blockchain == "ethereum" {
		return GetParams("geth")
	}
	return helpers.GetStaticBlockchainConfig(blockchain, "defaults.json")
}
