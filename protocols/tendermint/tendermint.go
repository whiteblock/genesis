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

//Package tendermint handles tendermint specific functionality
package tendermint

import (
	"encoding/json"
	"fmt"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/registrar"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"strings"
	"sync"
	"time"
)

type validatorPubKey struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type validator struct {
	Address string          `json:"address"`
	PubKey  validatorPubKey `json:"pub_key"`
	Power   string          `json:"power"`
	Name    string          `json:"name"`
}

var conf *util.Config

const blockchain = "tendermint"

func init() {
	conf = util.GetConfig()
	registrar.RegisterBuild(blockchain, Build)
	registrar.RegisterAddNodes(blockchain, Add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, helpers.DefaultGetDefaultsFn(blockchain))
	registrar.RegisterParams(blockchain, helpers.DefaultGetParamsFn(blockchain))
}

//ExecStart=/usr/bin/tendermint node --proxy_app=kvstore --p2p.persistent_peers=167b80242c300bf0ccfb3ced3dec60dc2a81776e@165.227.41.206:26656,3c7a5920811550c04bf7a0b2f1e02ab52317b5e6@165.227.43.146:26656,303a1a4312c30525c99ba66522dd81cca56a361a@159.89.115.32:26656,b686c2a7f4b1b46dca96af3a0f31a6a7beae0be4@159.89.119.125:26656

//Build builds out a fresh new tendermint test network
func Build(tn *testnet.TestNet) error {
	//Ensure that genesis file has same chain_id
	peers := []string{}
	validators := []validator{}
	tn.BuildState.SetBuildSteps(1 + (tn.LDD.Nodes * 4))
	tn.BuildState.SetBuildStage("Initializing the nodes")

	mux := sync.Mutex{}
	err := helpers.AllNodeExecCon(tn, func(client ssh.Client, server *db.Server, node ssh.Node) error {
		//init everything
		_, err := client.DockerExec(node, "tendermint init")
		if err != nil {
			return util.LogError(err)
		}

		//Get the node id
		res, err := client.DockerExec(node, "tendermint show_node_id")
		if err != nil {
			return util.LogError(err)
		}
		nodeID := res[:len(res)-1]

		mux.Lock()
		peers = append(peers, fmt.Sprintf("%s@%s:26656", nodeID, node.GetIP()))
		mux.Unlock()

		//Get the validators
		res, err = client.DockerExec(node, "cat /root/.tendermint/config/genesis.json")
		if err != nil {
			return util.LogError(err)
		}
		tn.BuildState.IncrementBuildProgress()
		var genesis map[string]interface{}
		err = json.Unmarshal([]byte(res), &genesis)
		if err != nil {
			return util.LogError(err)
		}
		validatorsRaw := genesis["validators"].([]interface{})
		for _, validatorRaw := range validatorsRaw {
			vdtr := validator{}

			validatorData := validatorRaw.(map[string]interface{})

			err = util.GetJSONString(validatorData, "address", &vdtr.Address)
			if err != nil {
				return util.LogError(err)
			}

			validatorPubKeyData := validatorData["pub_key"].(map[string]interface{})

			err = util.GetJSONString(validatorPubKeyData, "type", &vdtr.PubKey.Type)
			if err != nil {
				return util.LogError(err)
			}

			err = util.GetJSONString(validatorPubKeyData, "value", &vdtr.PubKey.Value)
			if err != nil {
				return util.LogError(err)
			}

			err = util.GetJSONString(validatorData, "power", &vdtr.Power)
			if err != nil {
				return util.LogError(err)
			}

			err = util.GetJSONString(validatorData, "name", &vdtr.Name)
			if err != nil {
				return util.LogError(err)
			}
			mux.Lock()
			validators = append(validators, vdtr)
			mux.Unlock()
		}
		tn.BuildState.IncrementBuildProgress()
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.SetBuildStage("Propogating the genesis file")

	//distribute the created genensis file among the nodes
	err = helpers.CopyBytesToAllNodes(tn, getGenesisFile(validators), "/root/.tendermint/config/genesis.json")
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.SetBuildStage("Starting tendermint")
	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, server *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		peersCpy := make([]string, len(peers))
		copy(peersCpy, peers)
		return client.DockerRunMainDaemon(node, fmt.Sprintf("tendermint node --proxy_app=kvstore --p2p.persistent_peers=%s",
			strings.Join(append(peersCpy[:node.GetAbsoluteNumber()], peersCpy[node.GetAbsoluteNumber()+1:]...), ",")))
	})
	return util.LogError(err)
}

// Add handles adding a node to the tendermint testnet
// TODO
func Add(tn *testnet.TestNet) error {
	return nil
}

func getGenesisFile(vdtrs []validator) string {
	validatorsStr, _ := json.Marshal(vdtrs)
	return fmt.Sprintf(`{
	  "genesis_time": "%s",
	  "chain_id": "whiteblock",
	  "consensus_params": {
		"block_size": {
		  "max_bytes": "22020096",
		  "max_gas": "-1"
		},
		"evidence": {
		  "max_age": "100000"
		},
		"validator": {
		  "pub_key_types": [
			"ed25519"
		  ]
		}
	  },
	  "validators": %s,
	  "app_hash": "" 
	}`, time.Now().Format("2006-01-02T15:04:05.000000000Z"),
		validatorsStr)
}
