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

//Package cosmos handles cosmos specific functionality
package cosmos

import (
	"fmt"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/registrar"
	"github.com/whiteblock/genesis/protocols/services"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"strings"
	"sync"
)

var conf = util.GetConfig()

const blockchain = "cosmos"

func init() {
	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterServices(blockchain, func() []services.Service { return nil })
	registrar.RegisterDefaults(blockchain, helpers.DefaultGetDefaultsFn(blockchain))
	registrar.RegisterParams(blockchain, helpers.DefaultGetParamsFn(blockchain))
}

// build builds out a fresh new cosmos test network
func build(tn *testnet.TestNet) error {
	tn.BuildState.SetBuildSteps(4 + (tn.LDD.Nodes * 2))

	tn.BuildState.SetBuildStage("Setting up the first node")
	/**
	 * Set up first node
	 */
	_, err := helpers.FirstNodeExec(tn, "gaiad init --chain-id=whiteblock whiteblock")
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.IncrementBuildProgress()
	_, err = helpers.FirstNodeExec(tn, "sh -c 'echo \"password\\n\" | gaiacli keys add validator -ojson'")
	if err != nil {
		return util.LogError(err)
	}

	res, err := helpers.FirstNodeExec(tn, "gaiacli keys show validator -a")
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.IncrementBuildProgress()
	_, err = helpers.FirstNodeExec(tn, fmt.Sprintf("gaiad add-genesis-account %s 100000000stake,100000000validatortoken",
		res[:len(res)-1]))
	if err != nil {
		return util.LogError(err)
	}

	_, err = helpers.FirstNodeExec(tn, "sh -c 'echo \"password\\n\" | gaiad gentx --name validator'")
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.IncrementBuildProgress()
	_, err = helpers.FirstNodeExec(tn, "gaiad collect-gentxs")
	if err != nil {
		return util.LogError(err)
	}
	genesisFile, err := helpers.FirstNodeExec(tn, "cat /root/.gaiad/config/genesis.json")
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Initializing the rest of the nodes")
	peers := make([]string, tn.LDD.Nodes)
	mux := sync.Mutex{}

	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, server *db.Server, node ssh.Node) error {
		ip := tn.Nodes[node.GetAbsoluteNumber()].IP
		if node.GetAbsoluteNumber() != 0 {
			//init everything
			_, err := client.DockerExec(node, "gaiad init --chain-id=whiteblock whiteblock")
			if err != nil {
				return util.LogError(err)
			}
		}

		//Get the node id
		res, err := client.DockerExec(node, "gaiad tendermint show-node-id")
		if err != nil {
			return util.LogError(err)
		}
		nodeID := res[:len(res)-1]
		mux.Lock()
		peers[node.GetAbsoluteNumber()] = fmt.Sprintf("%s@%s:26656", nodeID, ip)
		mux.Unlock()
		tn.BuildState.IncrementBuildProgress()
		return nil
	})

	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.SetBuildStage("Copying the genesis file to each node")

	err = helpers.CopyBytesToAllNodes(tn, genesisFile, "/root/.gaiad/config/genesis.json")
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.SetBuildStage("Starting cosmos")

	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, server *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		peersCpy := make([]string, len(peers))
		copy(peersCpy, peers)
		_, err := client.DockerExecd(node, fmt.Sprintf("gaiad start --p2p.persistent_peers=%s",
			strings.Join(append(peersCpy[:node.GetAbsoluteNumber()], peersCpy[node.GetAbsoluteNumber()+1:]...), ",")))
		return err
	})
	return err
}

// Add handles adding a node to the cosmos testnet
// TODO
func add(tn *testnet.TestNet) error {
	return nil
}
