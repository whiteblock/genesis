/*
	Copyright 2019 whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Genesis is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

// Package helpers contains functions to help make the task of deploying a blockchain easier and faster
package helpers

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"sync"
)

var conf = util.GetConfig()

/*
	fn func(client ssh.Client, server &db.Server,localNodeNum int,absoluteNodeNum int)(error)
*/
func allNodeExecCon(tn *testnet.TestNet, s settings, fn func(ssh.Client, *db.Server, ssh.Node) error) error {
	nodes := tn.GetSSHNodes(s.useNew, s.sidecar != -1, s.sidecar)
	wg := sync.WaitGroup{}
	for _, node := range nodes {

		wg.Add(1)
		go func(fwdClient ssh.Client, fwdServer *db.Server, fwdNode ssh.Node) {
			defer wg.Done()
			err := fn(fwdClient, fwdServer, fwdNode)
			if err != nil {
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
// return a non-nil error value, one of those errors will be returned. Currently there is no guarantee as to which one,
// however this should be implemented in the future.
func AllNodeExecCon(tn *testnet.TestNet, fn func(ssh.Client, *db.Server, ssh.Node) error) error {

	return allNodeExecCon(tn, settings{useNew: false, sidecar: -1, reportError: true}, fn)
}

// AllNewNodeExecCon is AllNodeExecCon but executes only for new nodes
func AllNewNodeExecCon(tn *testnet.TestNet, fn func(ssh.Client, *db.Server, ssh.Node) error) error {

	return allNodeExecCon(tn, settings{useNew: true, sidecar: -1, reportError: true}, fn)
}

// AllNewNodeExecConDR is AllNodeExecCon but executes only for new nodes, but only returns the error, doesn't
// report it to the build state automatically. (You most likely do NOT want this)
func AllNewNodeExecConDR(tn *testnet.TestNet, fn func(ssh.Client, *db.Server, ssh.Node) error) error {

	return allNodeExecCon(tn, settings{useNew: true, sidecar: -1, reportError: false}, fn)
}

// AllNodeExecConSC is AllNodeExecCon but executes only for sidecar nodes
func AllNodeExecConSC(ad *testnet.Adjunct, fn func(ssh.Client, *db.Server, ssh.Node) error) error {

	return allNodeExecCon(ad.Main, settings{useNew: false, sidecar: ad.Index, reportError: true}, fn)
}

// AllNewNodeExecConSC is AllNewNodeExecCon but executes only for sidecar nodes
func AllNewNodeExecConSC(ad *testnet.Adjunct, fn func(ssh.Client, *db.Server, ssh.Node) error) error {

	return allNodeExecCon(ad.Main, settings{useNew: true, sidecar: ad.Index, reportError: true}, fn)
}

// AllServerExecCon executes fn for every server in the testnet. Is sementatically similar to
// AllNodeExecCon. Every call to fn is provided with the relevant ssh client and server object.
func AllServerExecCon(tn *testnet.TestNet, fn func(ssh.Client, *db.Server) error) error {

	wg := sync.WaitGroup{}
	for _, server := range tn.Servers {
		wg.Add(1)
		go func(server *db.Server) {
			defer wg.Done()
			err := fn(tn.Clients[server.ID], server)
			if err != nil {
				tn.BuildState.ReportError(err)
				return
			}
		}(&server)
	}
	wg.Wait()
	return tn.BuildState.GetError()
}

func mkdirAllNodes(tn *testnet.TestNet, dir string, s settings) error {
	return allNodeExecCon(tn, s, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		_, err := client.DockerExec(node, fmt.Sprintf("mkdir -p %s", dir))
		return err
	})
}

// MkdirAllNodes makes a dir on all nodes
func MkdirAllNodes(tn *testnet.TestNet, dir string) error {
	return mkdirAllNodes(tn, dir, settings{useNew: false, sidecar: -1, reportError: true})
}

// MkdirAllNewNodes makes a dir on all new nodes
func MkdirAllNewNodes(tn *testnet.TestNet, dir string) error {
	return mkdirAllNodes(tn, dir, settings{useNew: true, sidecar: -1, reportError: true})
}

// AllServerExecConSC is like AllServerExecCon but for side cars
func AllServerExecConSC(ad *testnet.Adjunct, fn func(ssh.Client, *db.Server) error) error {
	return AllServerExecCon(ad.Main, fn)
}

// DefaultGetParamsFn creates the default function for getting a blockchains parameters
func DefaultGetParamsFn(blockchain string) func() string {
	return func() string {
		dat, err := GetStaticBlockchainConfig(blockchain, "params.json")
		if err != nil {
			//Missing required files is a fatal error
			log.WithFields(log.Fields{"blockchain": blockchain, "file": "params.json"}).Panic(err)
		}
		return string(dat)
	}
}

// DefaultGetDefaultsFn creates the default function for getting a blockchains default parameters
func DefaultGetDefaultsFn(blockchain string) func() string {
	return func() string {
		dat, err := GetStaticBlockchainConfig(blockchain, "defaults.json")
		if err != nil {
			//Missing required files is a fatal error
			log.WithFields(log.Fields{"blockchain": blockchain, "file": "defaults.json"}).Panic(err)
		}
		return string(dat)
	}
}

//JSONRPCAllNodes calls a JSON RPC call on all nodes and then returns the result
func JSONRPCAllNodes(tn *testnet.TestNet, call string, port int) ([]interface{}, error) {
	mux := sync.Mutex{}
	out := make([]interface{}, tn.LDD.Nodes)
	err := AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		for {
			res, err := client.KeepTryRun(
				fmt.Sprintf(
					`curl -sS -X POST http://%s:%d -H "Content-Type: application/json" `+
						` -d '{ "method": "%s", "params": [], "id": 1, "jsonrpc": "2.0" }'`,
					node.GetIP(), port, call))

			if err != nil {
				continue //could be infinite
			}
			var result map[string]interface{}

			err = json.Unmarshal([]byte(res), &result)
			if err != nil {
				return util.LogError(err)
			}
			_, ok := result["result"]
			if !ok {
				_, hasError := result["error"]
				if hasError {
					return fmt.Errorf("%v", result["error"])
				}
				return fmt.Errorf(res)

			}
			mux.Lock()
			out[node.GetAbsoluteNumber()] = result["result"]
			mux.Unlock()
			break
		}
		return nil
	})
	return out, err
}
