package deploy

import (
	db "../db"
	testnet "../testnet"
	util "../util"
	"fmt"
	"log"
	"sync"
)

/*
   Add nodes to the network instead of building independently. Functions similarly to build, except that it
   does not destroy the previous network when building.
*/
func AddNodes(tn *testnet.TestNet) error {

	tn.BuildState.SetDeploySteps(2 * tn.LDD.Nodes)
	defer tn.BuildState.FinishDeploy()
	wg := sync.WaitGroup{}

	fmt.Println("-------------Building The Docker Containers-------------")

	tn.BuildState.SetBuildStage("Provisioning the nodes")

	availibleServers := make([]int, len(tn.Servers))
	for i, _ := range availibleServers {
		availibleServers[i] = i
	}
	index := 0

	for i := 0; i < tn.LDD.Nodes; i++ {
		serverIndex := availibleServers[index]
		serverID := tn.Servers[serverIndex].Id

		if tn.Servers[serverIndex].Max <= tn.Servers[serverIndex].Nodes {
			if len(availibleServers) == 1 {
				return fmt.Errorf("Cannot build that many nodes with the availible resources")
			}
			availibleServers = append(availibleServers[:serverIndex], availibleServers[serverIndex+1:]...)
			i--
			index++
			index = index % len(availibleServers)
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
			LocalID: tn.Servers[serverIndex].Nodes, Ip: util.GetNodeIP(tn.Servers[serverIndex].SubnetID, len(tn.Nodes))})

		tn.Servers[serverIndex].Nodes++

		wg.Add(1)
		go func(serverID int, subnetID int, absNum int, relNum int) {
			defer wg.Done()
			err := DockerNetworkCreate(tn, serverID, subnetID, relNum)
			if err != nil {
				log.Println(err)
				tn.BuildState.ReportError(err)
				return
			}
			tn.BuildState.IncrementDeployProgress()
			image := tn.LDD.Images[0]
			resource := tn.LDD.Resources[0]
			if len(tn.LDD.Resources) > absNum {
				resource = tn.LDD.Resources[absNum]
			}
			if len(tn.LDD.Images) > absNum {
				image = tn.LDD.Images[absNum]
			}

			var env map[string]string = nil
			if tn.LDD.Environments != nil && len(tn.LDD.Environments) > absNum && tn.LDD.Environments[absNum] != nil {
				env = tn.LDD.Environments[absNum]
			}

			err = DockerRun(tn, serverID, subnetID, resource, relNum, image, env)
			if err != nil {
				log.Println(err)
				tn.BuildState.ReportError(err)
				return
			}
			tn.BuildState.IncrementDeployProgress()
		}(serverID, tn.Servers[serverIndex].SubnetID, absNum, relNum)

		index++
		index = index % len(availibleServers)
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
		client.Run("sudo iptables --flush DOCKER-ISOLATION-STAGE-1")
	}
	wg.Wait()

	log.Println("Finished adding nodes into the network")
	return tn.BuildState.GetError()
}
