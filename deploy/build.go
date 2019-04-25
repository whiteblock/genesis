package deploy

import (
	db "../db"
	testnet "../testnet"
	util "../util"
	"fmt"
	"log"
	"sync"
)

var conf *util.Config = util.GetConfig()

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
		relNum := len(tn.Servers[serverIndex].Ips)

		nodeID, err := util.GetUUIDString()
		if err != nil {
			log.Println(err)
			return err
		}

		tn.AddNode(db.Node{
			ID: nodeID, TestNetID: tn.TestNetID, Server: serverID,
			LocalID: tn.Servers[serverIndex].Nodes, Ip: util.GetNodeIP(tn.Servers[serverIndex].SubnetID, len(tn.Nodes))})
		tn.Servers[serverIndex].Ips = append(tn.Servers[serverIndex].Ips, util.GetNodeIP(tn.Servers[serverIndex].SubnetID, len(tn.Nodes))) //TODO: REMOVE
		tn.Servers[serverIndex].Nodes++

		wg.Add(1)
		go func(serverID int, subnetID int, absNum int, relNum int) {
			defer wg.Done()
			err := DockerNetworkCreate(tn, serverID, subnetID, relNum) //RACE
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

			err = DockerRun(tn, serverID, subnetID, resource, relNum, image, env)
			if err != nil {
				log.Println(err)
				tn.BuildState.ReportError(err)
				return
			}
			tn.BuildState.IncrementDeployProgress()
		}(serverID, tn.Servers[serverIndex].SubnetID, i, relNum)

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
	fmt.Printf("TESTNET=%+v\n", tn)

	return tn.BuildState.GetError()
}
