package deploy

import (
	db "../db"
	util "../util"
	"context"
	"golang.org/x/sync/semaphore"
	"log"
)

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
