// Package helpers contains functions to help make the task of deploying a blockchain easier and faster
package helpers

import (
	"../../db"
	"../../ssh"
	"../../testnet"
	"../../util"
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
func allNodeExecCon(tn *testnet.TestNet, useNew bool,sideCar bool, fn func(*ssh.Client, *db.Server, ssh.Node) error) error {
	var nodes []ssh.Node
	if useNew {
		nodes = tn.GetNewSSHNodes(sideCar)
	}else{
		nodes = tn.GetSSHNodes(sideCar)
	}

	wg := sync.WaitGroup{}
	for _, node := range nodes {

		wg.Add(1)
		go func(client *ssh.Client, server *db.Server, node ssh.Node) {
			defer wg.Done()
			err := fn(client, server, node)
			if err != nil {
				log.Println(err)
				tn.BuildState.ReportError(err)
				return
			}
		}(tn.Clients[node.GetServerID()], tn.GetServer(node.GetServerID()), node)

	}
	wg.Wait()
	return tn.BuildState.GetError()
}

// AllNodeExecCon executes fn for every node concurrently. Will return once all of the calls to fn
// have been completely.
// Each call to fn is provided with, in order, the relevant ssh client, the server where the node exists, the local
// number of that node on the server and the absolute number of the node in the testnet. If any of the calls to fn
// return a non-nil error value, one of those errors will be returned. Currently there is no guarentee as to which one,
// however this should be implemented in the future.
func AllNodeExecCon(tn *testnet.TestNet, fn func(*ssh.Client, *db.Server, ssh.Node) error) error {
	return allNodeExecCon(tn, false,false, fn)
}

// AllNewNodeExecCon is AllNodeExecCon but executes only for new nodes
func AllNewNodeExecCon(tn *testnet.TestNet, fn func(*ssh.Client, *db.Server, ssh.Node) error) error {
	return allNodeExecCon(tn, true,false, fn)
}

func AllNodeExecConSC(tn *testnet.TestNet, fn func(*ssh.Client, *db.Server, ssh.Node) error) error {
	return allNodeExecCon(tn, false,true, fn)
}

func AllNewNodeExecConSC(tn *testnet.TestNet, fn func(*ssh.Client, *db.Server, ssh.Node) error) error {
	return allNodeExecCon(tn, true,true, fn)
}

// AllServerExecCon executes fn for every server in the testnet. Is sementatically similar to
// AllNodeExecCon. Every call to fn is provided with the relevant ssh client and server object.
func AllServerExecCon(tn *testnet.TestNet, fn func(*ssh.Client, *db.Server) error) error {

	wg := sync.WaitGroup{}
	for _, server := range tn.Servers {
		wg.Add(1)
		go func(server *db.Server) {
			defer wg.Done()
			err := fn(tn.Clients[server.ID], server)
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
