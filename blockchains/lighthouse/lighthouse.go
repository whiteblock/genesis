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

//Package lighthouse handles lighthouse specific functionality
package lighthouse

import (
	"fmt"
	"github.com/whiteblock/genesis/blockchains/helpers"
	"github.com/whiteblock/genesis/blockchains/registrar"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"strings"
)

var conf *util.Config

const (
	blockchain = "lighthouse"
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

// build builds out a fresh new lighthouse test network
func build(tn *testnet.TestNet) error {
	_, err := newConf(tn.LDD.Params)
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.SetBuildSteps(1 + (tn.LDD.Nodes * 3))

	var bootNodes []string
	for _, node := range tn.Nodes {
		bootNodes = append(bootNodes, fmt.Sprintf("/dns4/whiteblock-node%d@%s/tcp/%d", node.LocalID, node.IP, p2pPort))
	}
	peers := fmt.Sprintf("--boot-nodes=%s", strings.Join(bootNodes, ","))
	tn.BuildState.IncrementBuildProgress()

	tn.BuildState.SetBuildStage("Starting lighthouse")
	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()

		lighthouseCmd := "RUST_LOG=libp2p=debug beacon_node --listen-address 0.0.0.0 --port 9000 " + peers + " 2>&1 | tee /output.log"
		return client.DockerExecdLog(node, lighthouseCmd)
	})
	return util.LogError(err)
}

// add handles adding nodes to the testnet
func add(tn *testnet.TestNet) error {
	return nil
}
