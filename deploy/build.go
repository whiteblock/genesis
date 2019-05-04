// Package deploy provides functions for building out nodes and test networks
package deploy

import (
	"../blockchains/registrar"
	"../db"
	"../docker"
	"../testnet"
	"../util"
	"fmt"
	"log"
	"sync"
)

var conf = util.GetConfig()

func buildSideCars(tn *testnet.TestNet, server *db.Server, node *db.Node) {
	sidecars, err := registrar.GetBlockchainSideCars(tn.LDD.Blockchain)
	if err != nil {
		log.Println(err) //do not report
		return
	}

	for i, sidecar := range sidecars {
		sideCarDetails, err := registrar.GetSideCar(sidecar)
		if err != nil {
			tn.BuildState.ReportError(err)
			return
		}

		sidecarIP, err := util.GetNodeIP(server.SubnetID, node.LocalID, i+1)
		if err != nil {
			tn.BuildState.ReportError(err)
			return
		}

		scNode := db.SideCar{
			NodeID:          node.ID,
			AbsoluteNodeNum: node.AbsoluteNum,
			TestnetID:       node.TestNetID,
			Server:          node.Server,
			LocalID:         node.LocalID,
			NetworkIndex:    i + 1,
			IP:              sidecarIP,
			Image:           sideCarDetails.Image,
			Type:            sidecar,
		}
		tn.AddSideCar(scNode, i)
		err = docker.Run(tn, server.ID, docker.NewSideCarContainer(&scNode, nil, util.Resources{}, server.SubnetID))
		if err != nil {
			tn.BuildState.ReportError(err)
			return
		}
	}
}

// BuildNode builds out a single node in a testnet
func BuildNode(tn *testnet.TestNet, server *db.Server, node *db.Node) {
	/*tn.BuildState.OnError(func() {
		docker.Kill(tn.Clients[server.ID], node.LocalID)
		docker.NetworkDestroy(tn.Clients[server.ID], node.LocalID)
	})*/
	defer buildSideCars(tn, server, node) //Needs to be handled better
	err := docker.NetworkCreate(tn, server.ID, server.SubnetID, node.LocalID)
	if err != nil {
		tn.BuildState.ReportError(err)
		return
	}
	tn.BuildState.IncrementDeployProgress()

	resource := tn.LDD.Resources[0]
	node.Image = tn.LDD.Images[0]
	var env map[string]string

	if len(tn.LDD.Resources) > node.AbsoluteNum {
		resource = tn.LDD.Resources[node.AbsoluteNum]
	}
	if len(tn.LDD.Images) > node.AbsoluteNum {
		node.Image = tn.LDD.Images[node.AbsoluteNum]
	}

	if tn.LDD.Environments != nil && len(tn.LDD.Environments) > node.AbsoluteNum && tn.LDD.Environments[node.AbsoluteNum] != nil {
		env = tn.LDD.Environments[node.AbsoluteNum]
	}

	err = docker.Run(tn, server.ID, docker.NewNodeContainer(node, env, resource, server.SubnetID))
	if err != nil {
		tn.BuildState.ReportError(err)
		return
	}

	tn.BuildState.IncrementDeployProgress()
}

// Build builds out the given docker network infrastructure according to the given parameters, and return
// the given array of servers, with ips updated for the nodes added to that server
func Build(tn *testnet.TestNet, services []util.Service) error {
	tn.BuildState.SetDeploySteps(3*tn.LDD.Nodes + 2 + len(services))
	defer tn.BuildState.FinishDeploy()
	wg := sync.WaitGroup{}

	tn.BuildState.SetBuildStage("Initializing build")

	err := handlePreBuildExtras(tn)
	if err != nil {
		return util.LogError(err)
	}
	PurgeTestNetwork(tn)

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

		tn.Servers[serverIndex].Ips = append(tn.Servers[serverIndex].Ips, nodeIP) //TODO: REMOVE
		tn.Servers[serverIndex].Nodes++

		wg.Add(1)
		go func(server *db.Server, node *db.Node) {
			defer wg.Done()
			BuildNode(tn, server, node)
		}(&tn.Servers[serverIndex], node)

		index = (index + 1) % len(availableServers)
	}

	if services != nil { //Maybe distribute the services over multiple servers
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := docker.StartServices(tn, services)
			if err != nil {
				tn.BuildState.ReportError(err)
				return
			}
		}()
	}
	wg.Wait()

	tn.BuildState.SetBuildStage("Setting up services")

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = finalize(tn)
		if err != nil {
			tn.BuildState.ReportError(err)
			return
		}
	}()

	for _, client := range tn.Clients {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.Run("sudo iptables --flush DOCKER-ISOLATION-STAGE-1")
		}()

	}
	distributeNibbler(tn)
	//Acquire all of the resources here, then release and destroy
	wg.Wait()

	//Check if we should freeze
	if tn.LDD.Extras != nil {
		shouldFreezeI, ok := tn.LDD.Extras["freezeAfterInfrastructure"]
		if ok {
			shouldFreeze, ok := shouldFreezeI.(bool)
			if ok && shouldFreeze {
				tn.BuildState.Freeze()
			}
		}
	}
	return tn.BuildState.GetError()
}
