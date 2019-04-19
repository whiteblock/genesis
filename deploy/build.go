package deploy

import (
	db "../db"
	ssh "../ssh"
	state "../state"
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
func Build(buildConf *db.DeploymentDetails, servers []db.Server, clients []*ssh.Client,
	services []util.Service, buildState *state.BuildState) ([]db.Server, error) {

	buildState.SetDeploySteps(3*buildConf.Nodes + 2 + len(services))
	defer buildState.FinishDeploy()
	wg := sync.WaitGroup{}

	buildState.SetBuildStage("Initializing build")

	err := handlePreBuildExtras(buildConf, clients, buildState)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	PurgeTestNetwork(servers, clients, buildState)

	buildState.SetBuildStage("Provisioning the nodes")

	availibleServers := make([]int, len(servers))
	for i, _ := range availibleServers {
		availibleServers[i] = i
	}

	index := 0
	for i := 0; i < buildConf.Nodes; i++ {
		serverIndex := availibleServers[index]
		if servers[serverIndex].Max <= servers[serverIndex].Nodes {
			if len(availibleServers) == 1 {
				return nil, fmt.Errorf("Cannot build that many nodes with the availible resources")
			}
			availibleServers = append(availibleServers[:serverIndex], availibleServers[serverIndex+1:]...)
			i--
			index++
			index = index % len(availibleServers)
			continue
		}
		relNum := len(servers[serverIndex].Ips)
		servers[serverIndex].Ips = append(servers[serverIndex].Ips, util.GetNodeIP(servers[serverIndex].SubnetID, i))
		servers[serverIndex].Nodes++

		wg.Add(1)
		go func(server db.Server, serverIndex int, absNum int, relNum int) {
			defer wg.Done()
			err := DockerNetworkCreate(server, clients[serverIndex], relNum) //RACE
			if err != nil {
				log.Println(err)
				buildState.ReportError(err)
				return
			}
			buildState.IncrementDeployProgress()

			resource := buildConf.Resources[0]
			image := buildConf.Images[0]
			var env map[string]string = nil

			if len(buildConf.Resources) > absNum {
				resource = buildConf.Resources[absNum]
			}
			if len(buildConf.Images) > absNum {
				image = buildConf.Images[absNum]
			}

			if buildConf.Environments != nil && len(buildConf.Environments) > absNum && buildConf.Environments[absNum] != nil {
				env = buildConf.Environments[absNum]
			}

			err = DockerRun(server, clients[serverIndex], resource, relNum, image, env)
			if err != nil {
				log.Println(err)
				buildState.ReportError(err)
				return
			}
			buildState.IncrementDeployProgress()
		}(servers[serverIndex], serverIndex, i, relNum)

		index++
		index = index % len(availibleServers)
	}

	if services != nil { //Maybe distribute the services over multiple servers
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := DockerStartServices(servers[0], clients[0], services, buildState)
			if err != nil {
				log.Println(err)
				buildState.ReportError(err)
				return
			}
		}()
	}
	wg.Wait()

	buildState.SetBuildStage("Setting up services")

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = finalize(servers, clients, buildState)
		if err != nil {
			log.Println(err)
			buildState.ReportError(err)
			return
		}
	}()

	for _, client := range clients {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.Run("sudo iptables --flush DOCKER-ISOLATION-STAGE-1")
		}()

	}
	distributeNibbler(servers, clients, buildState)
	//Acquire all of the resources here, then release and destroy
	wg.Wait()

	//Check if we should freeze
	if buildConf.Extras != nil {
		shouldFreezeI, ok := buildConf.Extras["freezeAfterInfrastructure"]
		if ok {
			shouldFreeze, ok := shouldFreezeI.(bool)
			if ok && shouldFreeze {
				buildState.Freeze()
			}
		}
	}

	return servers, buildState.GetError()
}
