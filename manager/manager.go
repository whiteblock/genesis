/*
	Copyright 2019 Whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Genesis is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

// Package manager contains functions for managing the testnets.
// Handles creating test nets, adding/removing nodes from testnets
package manager

import (
	"github.com/Whiteblock/genesis/blockchains/helpers"
	"github.com/Whiteblock/genesis/blockchains/registrar"
	"github.com/Whiteblock/genesis/db"
	"github.com/Whiteblock/genesis/deploy"
	"github.com/Whiteblock/genesis/testnet"
	"github.com/Whiteblock/genesis/util"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	//Put the relative path to your blockchain/sidecar library below this line, otherwise it won't be compiled
	//blockchains
	_ "github.com/Whiteblock/genesis/blockchains/artemis"
	_ "github.com/Whiteblock/genesis/blockchains/beam"
	_ "github.com/Whiteblock/genesis/blockchains/cosmos"
	_ "github.com/Whiteblock/genesis/blockchains/eos"
	_ "github.com/Whiteblock/genesis/blockchains/geth"
	_ "github.com/Whiteblock/genesis/blockchains/pantheon"
	_ "github.com/Whiteblock/genesis/blockchains/parity"
	_ "github.com/Whiteblock/genesis/blockchains/rchain"
	_ "github.com/Whiteblock/genesis/blockchains/syscoin"
	_ "github.com/Whiteblock/genesis/blockchains/tendermint"

	//side cars
	_ "github.com/Whiteblock/genesis/sidecars/geth"
	_ "github.com/Whiteblock/genesis/sidecars/orion"
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
		buildState.ReportError(err)
		return err
	}

	err = buildFn(tn)
	if err != nil {
		buildState.ReportError(err)
		return err
	}

	err = handleSideCars(tn, false)
	if err != nil {
		buildState.ReportError(err)
		return err
	}

	err = db.InsertBuild(*details, testnetID)
	if err != nil {
		buildState.ReportError(err)
		return err
	}
	err = tn.StoreNodes()
	if err != nil {
		buildState.ReportError(err)
		return err
	}
	return nil
}

func handleSideCars(tn *testnet.TestNet, append bool) error {
	sidecars, err := registrar.GetBlockchainSideCars(tn.LDD.Blockchain)
	if err != nil || sidecars == nil || len(sidecars) == 0 {
		return nil //Not an error, just means that the blockchain doesn't have any sidecars
	}
	wg := sync.WaitGroup{}
	wg.Add(len(sidecars))
	for i, sidecar := range sidecars { //In future, should probably check all the sidecars before running any builds
		var buildFn func(*testnet.Adjunct) error
		if append {
			buildFn, err = registrar.GetAddSideCar(sidecar)
		} else {
			buildFn, err = registrar.GetBuildSideCar(sidecar)
		}

		if err != nil {
			return util.LogError(err)
		}
		ad, err := tn.SpawnAdjunct(append, i)
		if err != nil {
			return util.LogError(err)
		}
		go func(i int) {
			defer wg.Done()
			err := buildFn(ad)
			if err != nil {
				tn.BuildState.ReportError(err)
			}
		}(i)
	}
	wg.Wait()
	return nil
}

func declareTestnet(testnetID string, details *db.DeploymentDetails) error {
	if len(details.GetJwt()) == 0 {
		return nil
	}
	data := map[string]interface{}{
		"id":        testnetID,
		"kind":      details.Blockchain,
		"num_nodes": details.Nodes,
		"image":     details.Images[0],
	}
	rawData, err := json.Marshal(data)
	if err != nil {
		return util.LogError(err)
	}
	_, err = util.JwtHTTPRequest("POST", "https://api.whiteblock.io/testnets", details.GetJwt(), string(rawData))
	return err
}

// DeleteTestNet destroys all of the nodes of a testnet
func DeleteTestNet(testnetID string) error {
	tn, err := testnet.RestoreTestNet(testnetID)
	if err != nil {
		return util.LogError(err)
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
