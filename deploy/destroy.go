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

package deploy

import (
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/docker"
	"github.com/whiteblock/genesis/protocols/helpers"
	//netem "github.com/whiteblock/genesis/net"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
)

// PurgeTestNetwork goes into each given ssh client and removes all the nodes and the networks.
// Increments the build state len(clients) * 2 times and sets it stag to tearing down network,
// if buildState is non nil.
func PurgeTestNetwork(tn *testnet.TestNet) error {
	if tn.BuildState != nil {
		tn.BuildState.SetBuildStage("Tearing down the previous testnet")
	}
	docker.StopServices(tn)
	return helpers.AllServerExecCon(tn, func(client ssh.Client, server *db.Server) error {
		docker.KillAll(client)
		if tn.BuildState != nil {
			tn.BuildState.IncrementDeployProgress()
		}
		docker.NetworkDestroyAll(client)
		if tn.BuildState != nil {
			tn.BuildState.IncrementDeployProgress()
		}
		//Redundant because the network is already destroy, so the tc rules are implicitly destroyed.
		//netem.RemoveAllOnServer(client, server.Nodes)

		return nil
	})
}

// Destroy is an alias of PurgeTestNetwork
func Destroy(tn *testnet.TestNet) error {
	return PurgeTestNetwork(tn)
}
