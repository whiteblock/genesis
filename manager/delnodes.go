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

package manager

import (
	"fmt"
	"github.com/whiteblock/genesis/docker"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
)

// DelNodes simply attempts to remove the given number of nodes from the
// network.
func DelNodes(num int, testnetID string) error {
	tn, err := testnet.RestoreTestNet(testnetID)
	if err != nil {
		return util.LogError(err)
	}
	if num >= len(tn.Nodes) {
		return fmt.Errorf("can't remove more than all the nodes in the network")
	}
	defer tn.FinishedBuilding()

	for i := len(tn.Nodes) - 1; i >= (len(tn.Nodes) - num); i-- {
		node := tn.Nodes[i]
		client := tn.Clients[node.GetServerID()]
		err = docker.Kill(client, node.GetRelativeNumber())
		if err != nil {
			return util.LogError(err)
		}
		err = docker.NetworkDestroy(client, node.GetRelativeNumber())
		if err != nil {
			return util.LogError(err)
		}
	}
	tn.Nodes = tn.Nodes[:(len(tn.Nodes) - num)]
	return nil
}
