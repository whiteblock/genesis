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
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"sync"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

/*
	fn func(client *ssh.Client, server &db.Server,localNodeNum int,absoluteNodeNum int)(error)
*/
func allNodeExecCon(tn *testnet.TestNet, useNew bool, sideCar int, fn func(*ssh.Client, *db.Server, ssh.Node) error) error {
	nodes := tn.GetSSHNodes(useNew, sideCar != -1, sideCar)
	wg := sync.WaitGroup{}
	for _, node := range nodes {

		wg.Add(1)
		go func(client *ssh.Client, server *db.Server, node ssh.Node) {
			defer wg.Done()
			err := fn(client, server, node)
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
// return a non-nil error value, one of those errors will be returned. Currently there is no guarentee as to which one,
// however this should be implemented in the future.
func AllNodeExecCon(tn *testnet.TestNet, fn func(*ssh.Client, *db.Server, ssh.Node) error) error {
	return allNodeExecCon(tn, false, -1, fn)
}

// AllNewNodeExecCon is AllNodeExecCon but executes only for new nodes
func AllNewNodeExecCon(tn *testnet.TestNet, fn func(*ssh.Client, *db.Server, ssh.Node) error) error {
	return allNodeExecCon(tn, true, -1, fn)
}

// AllNodeExecConSC is AllNodeExecCon but executes only for sidecar nodes
func AllNodeExecConSC(ad *testnet.Adjunct, fn func(*ssh.Client, *db.Server, ssh.Node) error) error {
	return allNodeExecCon(ad.Main, false, ad.Index, fn)
}

// AllNewNodeExecConSC is AllNewNodeExecCon but executes only for sidecar nodes
func AllNewNodeExecConSC(ad *testnet.Adjunct, fn func(*ssh.Client, *db.Server, ssh.Node) error) error {
	return allNodeExecCon(ad.Main, true, ad.Index, fn)
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
				tn.BuildState.ReportError(err)
				return
			}
		}(&server)
	}
	wg.Wait()
	return tn.BuildState.GetError()
}

// AllServerExecConSC is like AllServerExecCon but for side cars
func AllServerExecConSC(ad *testnet.Adjunct, fn func(*ssh.Client, *db.Server) error) error {
	return AllServerExecCon(ad.Main, fn)
}
