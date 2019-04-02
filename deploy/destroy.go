package deploy

import (
    "context"
    "log"
    "golang.org/x/sync/semaphore"
    db "../db"
    util "../util"
)

func Destroy(buildConf *db.DeploymentDetails,clients []*util.SshClient) error {
	var sem = semaphore.NewWeighted(conf.ThreadLimit)
    ctx := context.TODO()
    for i,_ := range clients {
        sem.Acquire(ctx,1)
        go func(i int){
            defer sem.Release(1)
            DockerKillAll(clients[i])
            DockerNetworkDestroyAll(clients[i])
        }(i)
    }

    for i,_ := range clients {
        sem.Acquire(ctx,1)
        go func(i int){
            defer sem.Release(1)
            DockerStopServices(clients[i])
        }(i)
    }

    err := sem.Acquire(ctx,conf.ThreadLimit)
    if err != nil {
        log.Println(err)
        return err
    }
    sem.Release(conf.ThreadLimit)
  	return nil
}