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

//Package plumtree handles plumtree specific functionality
package plumtree

import (
	"fmt"

	"github.com/whiteblock/genesis/blockchains/helpers"
	"github.com/whiteblock/genesis/blockchains/registrar"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
)

var conf *util.Config

const (
	blockchain = "plumtree"
	p2pPort    = 9000
)

func init() {
	conf = util.GetConfig()
	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterServices(blockchain, getServices)
	registrar.RegisterDefaults(blockchain, helpers.DefaultGetDefaultsFn(blockchain))
	registrar.RegisterParams(blockchain, helpers.DefaultGetParamsFn(blockchain))
	registrar.RegisterAdditionalLogs(blockchain, map[string]string{
		"json": "/plumtree/data/log.json"})
}

func getServices() []helpers.Service {
	return nil
}

// build builds out a fresh new plumtree test network
func build(tn *testnet.TestNet) error {

	tn.BuildState.SetBuildSteps(0 + (tn.LDD.Nodes * 1))
	tn.BuildState.SetBuildStage("Starting plumtree")

	return util.LogError(helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		peers := ""
		for _, peerNode := range tn.Nodes {
			if node.GetIP() != peerNode.GetIP() {
				peers += fmt.Sprintf(" --peer=tcp://whiteblock-node%d@%s:%d", peerNode.LocalID, peerNode.IP, p2pPort)
				tn.BuildState.IncrementBuildProgress()
			}
		}
		return client.DockerExecdLog(node, "gossip -n 0.0.0.0 -l 9000 -r 9001 -m /plumtree/data/log.json "+peers)
	}))
}

// Add handles adding a node to the plumtree testnet
// TODO
func add(tn *testnet.TestNet) error {
	return nil
}
