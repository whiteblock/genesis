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
			index++
			index = index % len(availableServers)
			continue
		}

		relNum := tn.Servers[serverIndex].Nodes

		nodeID, err := util.GetUUIDString()
		if err != nil {
			log.Println(err)
			return err
		}

		absNum := tn.AddNode(db.Node{
			ID: nodeID, TestNetID: tn.TestNetID, Server: serverID,
			LocalID: tn.Servers[serverIndex].Nodes, IP: util.GetNodeIP(tn.Servers[serverIndex].SubnetID, len(tn.Nodes))})

		tn.Servers[serverIndex].Nodes++

		wg.Add(1)
		go func(server *db.Server, absNum int, relNum int) {
			defer wg.Done()
			BuildNode(tn, server, absNum, relNum)
		}(&tn.Servers[serverIndex], absNum, relNum)

		index++
		index = index % len(availableServers)
	}
	wg.Wait()

	tn.BuildState.SetBuildStage("Setting up services")

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := finalizeNewNodes(tn)
		if err != nil {
			log.Println(err)
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
