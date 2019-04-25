package helpers

import (
	db "../../db"
	ssh "../../ssh"
	testnet "../../testnet"
	util "../../util"
	"log"
	"sync"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

/*
	fn func(client *ssh.Client, server &db.Server,localNodeNum int,absoluteNodeNum int)(error)
*/
func allNodeExecCon(tn *testnet.TestNet, useNew bool, fn func(*ssh.Client, *db.Server, int, int) error) error {
	nodes := tn.Nodes
	if useNew {
		nodes = tn.NewlyBuiltNodes
	}
	wg := sync.WaitGroup{}
	for _, node := range nodes {

		wg.Add(1)
		go func(client *ssh.Client, server *db.Server, localID int, absNum int) {
			defer wg.Done()
			err := fn(client, server, localID, absNum)
			if err != nil {
				log.Println(err)
				tn.BuildState.ReportError(err)
				return
			}
		}(tn.Clients[node.Server], tn.GetServer(node.Server), node.LocalID, node.AbsoluteNum)

	}
	wg.Wait()
	return tn.BuildState.GetError()
}

func AllNodeExecCon(tn *testnet.TestNet, fn func(*ssh.Client, *db.Server, int, int) error) error {
	return allNodeExecCon(tn, false, fn)
}

func AllNewNodeExecCon(tn *testnet.TestNet, fn func(*ssh.Client, *db.Server, int, int) error) error {
	return allNodeExecCon(tn, true, fn)
}

func AllServerExecCon(tn *testnet.TestNet, fn func(*ssh.Client, *db.Server) error) error {

	wg := sync.WaitGroup{}
	for _, server := range tn.Servers {
		wg.Add(1)
		go func(server *db.Server) {
			defer wg.Done()
			err := fn(tn.Clients[server.Id], server)
			if err != nil {
				log.Println(err)
				tn.BuildState.ReportError(err)
				return
			}
		}(&server)
	}
	wg.Wait()
	return tn.BuildState.GetError()
}
