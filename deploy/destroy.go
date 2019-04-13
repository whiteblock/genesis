package deploy

import (
	db "../db"
	netem "../net"
	state "../state"
	util "../util"
	"context"
	"golang.org/x/sync/semaphore"
	"log"
)

/*
PurgeTestNetwork goes into each given ssh client and removes all the nodes and the networks.
Increments the build state len(clients) * 2 times and sets it stag to tearing down network,
if buildState is non nil.
*/
func PurgeTestNetwork(servers []db.Server, clients []*util.SshClient, buildState *state.BuildState) {
	var sem = semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()
	if buildState != nil {
		buildState.SetBuildStage("Tearing down the previous testnet")
	}
	for i := range clients {
		sem.Acquire(ctx, 1)
		go func(i int) {
			defer sem.Release(1)
			DockerKillAll(clients[i])
			if buildState != nil {
				buildState.IncrementDeployProgress()
			}
			DockerNetworkDestroyAll(clients[i])
			if buildState != nil {
				buildState.IncrementDeployProgress()
			}
		}(i)
	}

	for i := range servers {
		sem.Acquire(ctx, 1)
		go func(i int) {
			defer sem.Release(1)
			netem.RemoveAllOnServer(clients[i], servers[i].Nodes)
		}(i)
	}

	for i := range servers {
		sem.Acquire(ctx, 1)
		go func(i int) {
			defer sem.Release(1)
			DockerStopServices(clients[i])
		}(i)
	}

	sem.Acquire(ctx, conf.ThreadLimit)
	sem.Release(conf.ThreadLimit)
}

func Destroy(buildConf *db.DeploymentDetails, clients []*util.SshClient) error {
	var sem = semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()
	for _, client := range clients {
		sem.Acquire(ctx, 1)
		go func(client *util.SshClient) {
			defer sem.Release(1)
			DockerKillAll(client)
			DockerNetworkDestroyAll(client)
		}(client)
	}

	for _, client := range clients {
		sem.Acquire(ctx, 1)
		go func(client *util.SshClient) {
			defer sem.Release(1)
			DockerStopServices(client)
		}(client)
	}

	err := sem.Acquire(ctx, conf.ThreadLimit)
	if err != nil {
		log.Println(err)
		return err
	}
	sem.Release(conf.ThreadLimit)
	return nil
}
