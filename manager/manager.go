/*
	Copyright 2019 whiteblock Inc.
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
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/deploy"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/registrar"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"sync"
	//Put the relative path to your blockchain/sidecar library below this line, otherwise it won't be compiled
	//blockchains
	_ "github.com/whiteblock/genesis/protocols/aion"
	_ "github.com/whiteblock/genesis/protocols/artemis"
	_ "github.com/whiteblock/genesis/protocols/beam"
	_ "github.com/whiteblock/genesis/protocols/cosmos"
	_ "github.com/whiteblock/genesis/protocols/eos"
	_ "github.com/whiteblock/genesis/protocols/ethclassic"
	_ "github.com/whiteblock/genesis/protocols/geth"
	_ "github.com/whiteblock/genesis/protocols/libp2p-test"
	_ "github.com/whiteblock/genesis/protocols/lighthouse"
	_ "github.com/whiteblock/genesis/protocols/lodestar"
	_ "github.com/whiteblock/genesis/protocols/multigeth"
	_ "github.com/whiteblock/genesis/protocols/pantheon"
	_ "github.com/whiteblock/genesis/protocols/parity"
	_ "github.com/whiteblock/genesis/protocols/plumtree"
	_ "github.com/whiteblock/genesis/protocols/polkadot"
	_ "github.com/whiteblock/genesis/protocols/prysm"
	_ "github.com/whiteblock/genesis/protocols/rchain"
	_ "github.com/whiteblock/genesis/protocols/syscoin"
	_ "github.com/whiteblock/genesis/protocols/tendermint"

	//side cars
	_ "github.com/whiteblock/genesis/sidecars/geth"
	_ "github.com/whiteblock/genesis/sidecars/orion"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

// AddTestNet implements the build command. All blockchains Build command must be
// implemented here, other it will not be called during the build process.
func AddTestNet(details *db.DeploymentDetails, testnetID string) error {
	if details.Servers == nil || len(details.Servers) == 0 {
		log.WithFields(log.Fields{"build": testnetID}).Error("build request doesn't have any servers")
		return fmt.Errorf("missing servers")
	}
	//STEP 1: SETUP THE TESTNET
	tn, err := testnet.NewTestNet(*details, testnetID)
	if err != nil {
		log.WithFields(log.Fields{"build": testnetID, "error": err}).Error("failed to create new testnet")
		return err
	}
	buildState := tn.BuildState
	defer tn.FinishedBuilding()

	//STEP 0: VALIDATE
	err = validate(details)
	if err != nil {
		log.WithFields(log.Fields{"details": details}).Error("invalid build details")
		tn.BuildState.ReportError(err)
		return err
	}

	tn.BuildState.Async(func() {
		declareTestnet(testnetID, details)
	})

	//STEP 3: GET THE SERVICES
	servicesFn, err := registrar.GetServiceFunc(details.Blockchain)
	if err != nil {
		tn.BuildState.ReportError(err)
		return err
	}
	services := servicesFn()
	//STEP 4: BUILD OUT THE DOCKER CONTAINERS AND THE NETWORK

	err = deploy.Build(tn, services)
	if err != nil {
		tn.BuildState.ReportError(err)
		return err
	}
	log.WithFields(log.Fields{"build": testnetID}).Trace("Built the docker containers")

	buildFn, err := registrar.GetBuildFunc(details.Blockchain)
	if err != nil {
		buildState.ReportError(err)
		return err
	}
	sidecars, err := registrar.GetBlockchainSideCars(tn.LDD.Blockchain)
	if err == nil && len(sidecars) > 0 {
		tn.BuildState.SetSidecars(len(sidecars))
	}

	err = buildFn(tn)
	if err != nil {
		buildState.ReportError(err)
		return err
	}

	if len(sidecars) > 0 {
		tn.BuildState.SetBuildStage("setting up the sidecars")
		steps := 0
		for _, sidecarName := range sidecars {
			sidecar, err := registrar.GetSideCar(sidecarName)
			if err != nil {
				buildState.ReportError(err)
				return err
			}
			if sidecar.BuildStepsCalc != nil {
				steps += sidecar.BuildStepsCalc(tn.LDD.Nodes, len(tn.Servers))
			}
		}
		tn.BuildState.SetSidecarSteps(steps)
		tn.BuildState.FinishMainBuild()
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
	if len(details.GetJwt()) == 0 || conf.DisableTestnetReporting {
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
	_, err = util.JwtHTTPRequest("POST", conf.APIEndpoint+"/testnets", details.GetJwt(), string(rawData))
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

// GetParams fetches the name and type of each available
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
