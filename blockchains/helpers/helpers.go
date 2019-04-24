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
	fn func(serverId int,localNodeNum int,absoluteNodeNum int)(error)
*/
func AllNodeExecCon(tn *testnet.TestNet,fn func(*ssh.Client,int,int,int) error) error {
	wg := sync.WaitGroup{}
	for _, node := range tn.Nodes {
		
		wg.Add(1)
		go func(node *db.Node) {
			defer wg.Done()
			err := fn(tn.Clients[node.Server],node.Server, node.LocalID, node.AbsoluteNum)
			if err != nil {
				log.Println(err)
				tn.BuildState.ReportError(err)
				return
			}
		}(&node)
		
	}
	wg.Wait()
	return tn.BuildState.GetError()
}

func AllServerExecCon(tn *testnet.TestNet,fn func(*ssh.Client,*db.Server) error) error {

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
		}( &server)
	}
	wg.Wait()
	return tn.BuildState.GetError()
}
