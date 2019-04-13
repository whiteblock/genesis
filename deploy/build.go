package deploy

import (
	db "../db"
	state "../state"
	util "../util"
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"log"
)

var conf *util.Config = util.GetConfig()

/*
   Build out the given docker network infrastructure according to the given parameters, and return
   the given array of servers, with ips updated for the nodes added to that server
*/
func Build(buildConf *db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
	services []util.Service, buildState *state.BuildState) ([]db.Server, error) {

	buildState.SetDeploySteps(3*buildConf.Nodes + 2 + len(services))
	defer buildState.FinishDeploy()

	var sem = semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()

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

		servers[serverIndex].Ips = append(servers[serverIndex].Ips, util.GetNodeIP(servers[serverIndex].SubnetID, i))
		servers[serverIndex].Nodes++

		sem.Acquire(ctx, 1)
		go func(serverIndex int, i int) {
			defer sem.Release(1)
			err := DockerNetworkCreate(servers[serverIndex], clients[serverIndex], i)
			if err != nil {
				log.Println(err)
				buildState.ReportError(err)
				return
			}
			buildState.IncrementDeployProgress()

			resource := buildConf.Resources[0]
			var env map[string]string = nil

			if len(buildConf.Resources) > i {
				resource = buildConf.Resources[i]
			}
			if buildConf.Environments != nil && len(buildConf.Environments) > i && buildConf.Environments[i] != nil {
				env = buildConf.Environments[i]
			}

			err = DockerRun(servers[serverIndex], clients[serverIndex], resource, i, buildConf.Image, env)
			if err != nil {
				log.Println(err)
				buildState.ReportError(err)
				return
			}
			buildState.IncrementDeployProgress()
		}(serverIndex, i)

		index++
		index = index % len(availibleServers)
	}

	if services != nil { //Maybe distribute the services over multiple servers
		sem.Acquire(ctx, 1)
		go func() {
			defer sem.Release(1)
			err := DockerStartServices(servers[0], clients[0], services, buildState)
			if err != nil {
				log.Println(err)
				buildState.ReportError(err)
				return
			}
		}()
	}

	err = sem.Acquire(ctx, conf.ThreadLimit)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	sem.Release(conf.ThreadLimit)

	buildState.SetBuildStage("Setting up services")

	sem.Acquire(ctx, 1)
	go func() {
		defer sem.Release(1)
		err = finalize(servers, clients, buildState)
		if err != nil {
			log.Println(err)
			buildState.ReportError(err)
			return
		}
	}()

	for _, client := range clients {
		client.Run("sudo iptables --flush DOCKER-ISOLATION-STAGE-1")
	}

	//Acquire all of the resources here, then release and destroy
	err = sem.Acquire(ctx, conf.ThreadLimit)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	sem.Release(conf.ThreadLimit)

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
