package helpers

import (
	db "../../db"
	state "../../state"
	util "../../util"
	"context"
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
		for j := range server.Ips {
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
	return buildState.GetError()
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

	return buildState.GetError()
}
