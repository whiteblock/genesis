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

package manager

import (
	"fmt"
	"github.com/Whiteblock/genesis/db"
	"github.com/Whiteblock/genesis/docker"
	"github.com/Whiteblock/genesis/status"
	"github.com/Whiteblock/genesis/util"
)

// DelNodes simply attempts to remove the given number of nodes from the
// network.
func DelNodes(num int, testnetID string) error {
	//buildState := state.GetBuildStateByServerId(details.Servers[0])
	//defer buildState.DoneBuilding()

	nodes, err := db.GetAllNodesByTestNet(testnetID)
	if err != nil {
		//buildState.ReportError(err)
		return util.LogError(err)
	}

	if num >= len(nodes) {
		err = fmt.Errorf("can't remove more than all the nodes in the network")
		//buildState.ReportError(err)
		return err
	}

	servers, err := status.GetLatestServers(testnetID)
	if err != nil {
		//buildState.ReportError(err)
		return util.LogError(err)
	}

	toRemove := num
	for _, server := range servers {
		client, err := status.GetClient(server.ID)
		if err != nil {
			//buildState.ReportError(err)
			return util.LogError(err)
		}
		for i := len(server.Ips); i > 0; i++ {
			err = docker.Kill(client, i)
			if err != nil {
				//buildState.ReportError(err)
				return util.LogError(err)
			}

			err = docker.NetworkDestroy(client, i)
			if err != nil {
				//buildState.ReportError(err)
				return util.LogError(err)
			}

			toRemove--
			if toRemove == 0 {
				break
			}
		}
		if toRemove == 0 {
			break
		}
	}
	return nil
}
