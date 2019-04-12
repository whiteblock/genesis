package helpers

import (
	db "../../db"
	state "../../state"
	util "../../util"
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"log"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

/*
	fn func(serverNum int,localNodeNum int,absoluteNodeNum int)(error)
*/
func AllNodeExecCon(servers []db.Server, buildState *state.BuildState,
	fn func(serverNum int, localNodeNum int, absoluteNodeNum int) error) error {

	sem := semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()
	node := 0
	for i, server := range servers {
		for j, _ := range server.Ips {
			sem.Acquire(ctx, 1)
			go func(i int, j int, node int) {
				defer sem.Release(1)
				err := fn(i, j, node)
				if err != nil {
					log.Println(err)
					buildState.ReportError(err)
					return
				}
			}(i, j, node)
			node++
		}
	}

	sem.Acquire(ctx, conf.ThreadLimit)
	sem.Release(conf.ThreadLimit)
	if !buildState.ErrorFree() {
		return buildState.GetError()
	}
	return nil
}

func AllServerExecCon(servers []db.Server, buildState *state.BuildState,
	fn func(serverNum int, server *db.Server) error) error {

	sem := semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()
	for i, server := range servers {
		sem.Acquire(ctx, 1)
		go func(serverNum int, server *db.Server) {
			defer sem.Release(1)
			err := fn(serverNum, server)
			if err != nil {
				log.Println(err)
				buildState.ReportError(err)
				return
			}
		}(i, &server)
	}

	sem.Acquire(ctx, conf.ThreadLimit)
	sem.Release(conf.ThreadLimit)
	if !buildState.ErrorFree() {
		return buildState.GetError()
	}
	return nil
}

func CopyToServers(servers []db.Server, clients []*util.SshClient, buildState *state.BuildState, src string, dst string) error {
	return AllServerExecCon(servers, buildState, func(serverNum int, server *db.Server) error {
		buildState.Defer(func() { clients[serverNum].Run(fmt.Sprintf("rm -rf %s", dst)) })
		return clients[serverNum].Scp(src, dst)
	})

}

func CopyAllToServers(servers []db.Server, clients []*util.SshClient, buildState *state.BuildState, srcDst ...string) error {
	if len(srcDst)%2 != 0 {
		return fmt.Errorf("Invalid number of variadic arguments, must be given an even number of them")
	}
	sem := semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()

	for i, _ := range servers {
		for j := 0; j < len(srcDst)/2; j++ {
			sem.Acquire(ctx, 1)
			go func(i int, j int) {
				defer sem.Release(1)
				buildState.Defer(func() { clients[i].Run(fmt.Sprintf("rm -rf %s", srcDst[2*j+1])) })
				err := clients[i].Scp(srcDst[2*j], srcDst[2*j+1])
				if err != nil {
					log.Println(err)
					buildState.ReportError(err)
					return
				}
			}(i, j)

		}
	}

	sem.Acquire(ctx, conf.ThreadLimit)
	sem.Release(conf.ThreadLimit)
	if !buildState.ErrorFree() {
		return buildState.GetError()
	}
	return nil
}
