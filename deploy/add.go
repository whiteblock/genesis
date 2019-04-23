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

/*
   Add nodes to the network instead of building independently. Functions similarly to build, except that it
   does not destroy the previous network when building.
*/
func AddNodes(buildConf *db.DeploymentDetails, servers []db.Server, clients []*ssh.Client,
	buildState *state.BuildState) (map[int][]string, error) {

	buildState.SetDeploySteps(2 * buildConf.Nodes)
	defer buildState.FinishDeploy()
	wg := sync.WaitGroup{}

	fmt.Println("-------------Building The Docker Containers-------------")

	buildState.SetBuildStage("Provisioning the nodes")

	availibleServers := make([]int, len(servers))
	for i, _ := range availibleServers {
		availibleServers[i] = i
	}
	out := map[int][]string{}
	index := 0

	for i := 0; i < buildConf.Nodes; i++ {
		serverIndex := availibleServers[index]
		nodeNum := len(servers[serverIndex].Ips) + i
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
		out[servers[serverIndex].Id] = append(out[servers[serverIndex].Id], util.GetNodeIP(servers[serverIndex].SubnetID, nodeNum))

		wg.Add(1)
		go func(serverIndex int, i int) {
			defer wg.Done()
			err := DockerNetworkCreate(servers[serverIndex], clients[serverIndex], i)
			if err != nil {
				log.Println(err)
				buildState.ReportError(err)
				return
			}
			buildState.IncrementDeployProgress()
			image := buildConf.Images[0]
			resource := buildConf.Resources[0]
			if len(buildConf.Resources) > i {
				resource = buildConf.Resources[i]
			}
			if len(buildConf.Images) > i {
				image = buildConf.Images[i]
			}

			var env map[string]string = nil
			if buildConf.Environments != nil && len(buildConf.Environments) > i && buildConf.Environments[i] != nil {
				env = buildConf.Environments[i]
			}

			err = DockerRun(servers[serverIndex], clients[serverIndex], resource, i, image, env)
			if err != nil {
				log.Println(err)
				buildState.ReportError(err)
				return
			}
			buildState.IncrementDeployProgress()
		}(serverIndex, nodeNum)

		index++
		index = index % len(availibleServers)
	}
	wg.Wait()

	buildState.SetBuildStage("Setting up services")

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := finalizeNewNodes(servers, clients, out, buildState)
		if err != nil {
			log.Println(err)
			buildState.ReportError(err)
			return
		}
	}()

	for i, _ := range servers {
		clients[i].Run("sudo iptables --flush DOCKER-ISOLATION-STAGE-1")
	}
	wg.Wait()

	log.Println("Finished adding nodes into the network")
	return out, buildState.GetError()
}
