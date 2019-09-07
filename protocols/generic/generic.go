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

package generic

import (
	"fmt"

	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/registrar"
	"github.com/whiteblock/genesis/protocols/services"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
)

const (
	blockchain     = "generic"
)

func init() {
	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterServices(blockchain, func() []services.Service { return nil })
	registrar.RegisterDefaults(blockchain, helpers.DefaultGetDefaultsFn(blockchain))
	registrar.RegisterParams(blockchain, helpers.DefaultGetParamsFn(blockchain))
}

// build builds out a fresh new ethereum test network using geth
func build(tn *testnet.TestNet) error {
	tn.BuildState.SetBuildSteps(3 + tn.LDD.Nodes)

	tn.BuildState.IncrementBuildProgress()

	tn.BuildState.SetBuildStage("Creating the genesis block")

	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Bootstrapping network")

	staticNodes := make([]string, tn.LDD.Nodes)

	for i, node := range tn.Nodes {
		staticNodes[i] = node.GetIP()
	}

	tn.BuildState.IncrementBuildProgress()

	err := helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		var params string

		for key, param := range tn.LDD.Params {
			params += fmt.Sprintf(" --%s %v", key, param)
		}

		params += "--peers"
		for _, peer := range staticNodes {
			params += fmt.Sprintf(" %v", peer)
		}

		_, err := client.DockerExec(node, fmt.Sprintf("sh /start.sh %s", params))
		if err != nil {
			return util.LogError(err)
		}

		tn.BuildState.IncrementBuildProgress()
		return nil
	})

	return err
}

func add(tn *testnet.TestNet) error {
	err := build(tn)

	return err
}
