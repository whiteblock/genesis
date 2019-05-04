package deploy

import (
	"../db"
	"../testnet"
	"../util"
	"fmt"
	"log"
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
				return fmt.Errorf("cannot build that many nodes with the availible resources")
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
			LocalID: tn.Servers[serverIndex].Nodes, IP: nodeIP})

		tn.Servers[serverIndex].Nodes++

		wg.Add(1)
		go func(server *db.Server, node *db.Node) {
			defer wg.Done()
			BuildNode(tn, server, node)
		}(&tn.Servers[serverIndex], node)

		index = (index + 1) % len(availableServers)
	}
	wg.Wait()
	distributeNibbler(tn)
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
		client.Run("sudo iptables --flush DOCKER-ISOLATION-STAGE-1")
	}
	wg.Wait()

	log.Println("Finished adding nodes into the network")
	return tn.BuildState.GetError()
}
