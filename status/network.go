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

package status

import (
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/util"
)

// GetLatestServers gets the servers used in the latest testnet, populated with the
// ips of all the nodes
func GetLatestServers(testnetID string) ([]db.Server, error) {
	nodes, err := db.GetAllNodesByTestNet(testnetID)
	if err != nil {
		return nil, util.LogError(err)
	}
	servers, err := db.GetServers(db.GetUniqueServerIDs(nodes))
	return servers, util.LogError(err)
}
