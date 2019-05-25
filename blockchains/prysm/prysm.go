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

//Package prysm handles prysm specific functionality
package prysm

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
	blockchain = "prysm"
	p2pPort    = 9000
)

func init() {
	conf = util.GetConfig()

	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, helpers.DefaultGetDefaultsFn(blockchain))
	registrar.RegisterParams(blockchain, helpers.DefaultGetParamsFn(blockchain))
}

// build builds out a fresh new prysm test network
func build(tn *testnet.TestNet) error {
	_, err := newConf(tn.LDD.Params)
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.SetBuildSteps(0 + (tn.LDD.Nodes * 4))

	peers := ""
	for _, node := range tn.Nodes {
		peers += fmt.Sprintf(" --peer=/dns4/whiteblock-node%d@%s/tcp/%d", node.LocalID, node.IP, p2pPort)
		tn.BuildState.IncrementBuildProgress()
	}

	tn.BuildState.SetBuildStage("Starting prysm")
	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		return client.DockerExecdLog(node, "/beacon-chain --no-discovery "+peers)
	})
	return util.LogError(err)
}

// add handles adding nodes to the testnet
func add(tn *testnet.TestNet) error {
	return nil
}
