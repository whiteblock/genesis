package deploy

import (
	"../db"
	"../testnet"
	"../util"
	"fmt"
	"log"
	"sync"
)

var conf = util.GetConfig()

func BuildNode(tn *testnet.TestNet, server *db.Server, absNum int, relNum int) {
	tn.BuildState.OnError(func() {
		DockerKill(tn.Clients[server.Id], relNum)
		DockerNetworkDestroy(tn.Clients[server.Id], relNum)
	})
	err := DockerNetworkCreate(tn, server.Id, server.SubnetID, relNum) //RACE
	if err != nil {
		log.Println(err)
		tn.BuildState.ReportError(err)
		return
	}
	tn.BuildState.IncrementDeployProgress()

	resource := tn.LDD.Resources[0]
	image := tn.LDD.Images[0]
	var env map[string]string = nil

	if len(tn.LDD.Resources) > absNum {
		resource = tn.LDD.Resources[absNum]
	}
	if len(tn.LDD.Images) > absNum {
		image = tn.LDD.Images[absNum]
	}

	if tn.LDD.Environments != nil && len(tn.LDD.Environments) > absNum && tn.LDD.Environments[absNum] != nil {
		env = tn.LDD.Environments[absNum]
	}

	err = DockerRun(tn, server.Id, server.SubnetID, resource, relNum, image, env)
	if err != nil {
		log.Println(err)
		tn.BuildState.ReportError(err)
		return
	}

	tn.BuildState.IncrementDeployProgress()
}

/*
   Build out the given docker network infrastructure according to the given parameters, and return
   the given array of servers, with ips updated for the nodes added to that server
*/
func Build(tn *testnet.TestNet, services []util.Service) error {
	tn.BuildState.SetDeploySteps(3*tn.LDD.Nodes + 2 + len(services))
	defer tn.BuildState.FinishDeploy()
	wg := sync.WaitGroup{}

	tn.BuildState.SetBuildStage("Initializing build")

	err := handlePreBuildExtras(tn)
	if err != nil {
		log.Println(err)
		return err
	}
	PurgeTestNetwork(tn)

	tn.BuildState.SetBuildStage("Provisioning the nodes")

	availibleServers := make([]int, len(tn.Servers))
	for i := range availibleServers {
		availibleServers[i] = i
	}

	index := 0
	for i := 0; i < tn.LDD.Nodes; i++ {
		serverIndex := availibleServers[index]
		serverID := tn.Servers[serverIndex].Id

		if tn.Servers[serverIndex].Max <= tn.Servers[serverIndex].Nodes {
			if len(availibleServers) == 1 {
				return fmt.Errorf("cannot build that many nodes with the availible resources")
			}
			availibleServers = append(availibleServers[:serverIndex], availibleServers[serverIndex+1:]...)
			i--
			index++
			index = index % len(availibleServers)
			continue
		}
		relNum := len(tn.Servers[serverIndex].Ips)

		nodeID, err := util.GetUUIDString()
		if err != nil {
			log.Println(err)
			return err
		}

		tn.AddNode(db.Node{
			ID: nodeID, TestNetID: tn.TestNetID, Server: serverID,
			LocalID: tn.Servers[serverIndex].Nodes, IP: util.GetNodeIP(tn.Servers[serverIndex].SubnetID, len(tn.Nodes))})
		tn.Servers[serverIndex].Ips = append(tn.Servers[serverIndex].Ips, util.GetNodeIP(tn.Servers[serverIndex].SubnetID, len(tn.Nodes))) //TODO: REMOVE
		tn.Servers[serverIndex].Nodes++

		wg.Add(1)
		go func(server *db.Server, absNum int, relNum int) {
			defer wg.Done()
			BuildNode(tn, server, absNum, relNum)
		}(&tn.Servers[serverIndex], i, relNum)

		index++
		index = index % len(availibleServers)
	}

	if services != nil { //Maybe distribute the services over multiple servers
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := DockerStartServices(tn, services)
			if err != nil {
				log.Println(err)
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
			log.Println(err)
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
