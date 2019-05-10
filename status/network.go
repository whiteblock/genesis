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

package status

import (
	"github.com/Whiteblock/genesis/db"
	"log"
)

// GetLatestServers gets the servers used in the latest testnet, populated with the
// ips of all the nodes
func GetLatestServers(testnetID string) ([]db.Server, error) {
	nodes, err := db.GetAllNodesByTestNet(testnetID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	serverIDs := db.GetUniqueServerIDs(nodes)

	servers, err := db.GetServers(serverIDs)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	for _, node := range nodes {
		for i := range servers {
			if servers[i].Ips == nil {
				servers[i].Ips = []string{}
			}
			if node.Server == servers[i].ID {
				servers[i].Ips = append(servers[i].Ips, node.IP)
			}
			servers[i].Nodes++
		}
	}
	return servers, nil
}
