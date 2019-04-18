package helpers

import (
	db "../../db"
	state "../../state"
	util "../../util"
	"log"
	"sync"
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
	node := 0
	wg := sync.WaitGroup{}
	for i, server := range servers {
		for j := range server.Ips {
			wg.Add(1)
			go func(i int, j int, node int) {
				defer wg.Done()
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
	wg.Wait()
	return buildState.GetError()
}

func AllServerExecCon(servers []db.Server, buildState *state.BuildState,
	fn func(serverNum int, server *db.Server) error) error {

	wg := sync.WaitGroup{}
	for i, server := range servers {
		wg.Add(1)
		go func(serverNum int, server *db.Server) {
			defer wg.Done()
			err := fn(serverNum, server)
			if err != nil {
				log.Println(err)
				buildState.ReportError(err)
				return
			}
		}(i, &server)
	}
	wg.Wait()
	return buildState.GetError()
}
