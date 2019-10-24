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
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"sync"
)

// AddNodes adds nodes to the network instead of building independently. Functions similarly to build, except that it
// does not destroy the previous network when building.
func AddNodes(tn *testnet.TestNet) error {

	tn.BuildState.SetDeploySteps(2 * tn.LDD.Nodes)
	defer tn.BuildState.FinishDeploy()
	wg := sync.WaitGroup{}

	tn.BuildState.SetBuildStage("Provisioning the nodes")

	availableServers := make([]int, len(tn.Servers))
	for i := range availableServers {
		availableServers[i] = i
	}
	index := 0

	for i := 0; i < tn.LDD.Nodes; i++ {
		serverIndex := availableServers[index]
		serverID := tn.Servers[serverIndex].ID

		if tn.Servers[serverIndex].Max <= tn.Servers[serverIndex].Nodes {
			if len(availableServers) == 1 {
				return fmt.Errorf("cannot build that many nodes with the available resources")
			}
			availableServers = append(availableServers[:serverIndex], availableServers[serverIndex+1:]...)
			i--
			index = (index + 1) % len(availableServers)
			continue
		}

		nodeID, err := util.GetUUIDString()
		if err != nil {
			return util.LogError(err)
		}

		nodeIP, err := util.GetNodeIP(tn.Servers[serverIndex].SubnetID, len(tn.Nodes), 0)
		if err != nil {
			return util.LogError(err)
		}

		node := tn.AddNode(db.Node{
			ID: nodeID, TestNetID: tn.TestNetID, Server: serverID,
			LocalID: tn.Servers[serverIndex].Nodes, IP: nodeIP, Protocol: tn.LDD.Blockchain})

		tn.Servers[serverIndex].Nodes++

		wg.Add(1)
		go func(server *db.Server, node *db.Node) {
			defer wg.Done()
			BuildNode(tn, server, node)
		}(&tn.Servers[serverIndex], node)

		index = (index + 1) % len(availableServers)
	}
	wg.Wait()

	tn.BuildState.SetBuildStage("Setting up services")

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := finalizeNewNodes(tn)
		if err != nil {
			tn.BuildState.ReportError(err)
			return
		}
	}()

	for _, client := range tn.Clients {
		//noinspection SpellCheckingInspection
		client.Run("sudo -n iptables --flush DOCKER-ISOLATION-STAGE-1")
	}
	wg.Wait()

	log.Info("finished adding nodes into the network")
	return tn.BuildState.GetError()
}
