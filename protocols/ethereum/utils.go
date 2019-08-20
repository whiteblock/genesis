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

package ethereum

import (
	"fmt"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
)

const (
	//P2PPort is the standard p2p port for all the ethereum clients
	P2PPort = 30303
	//RPCPort is the standard json rpc over http port
	RPCPort = 8545
	//EnodeKey is the state key for enodes
	EnodeKey = "enodes"
)

//CreatePasswordFile turns the process of creating a password file into a single function call
func CreatePasswordFile(tn *testnet.TestNet, password string, dest string) error {
	return CreateNPasswordFile(tn, tn.LDD.Nodes, password, dest)
}

//CreateNPasswordFile turns the process of creating a password file into a single function call,
//while also allowing you to specify the number of passwords in the file
func CreateNPasswordFile(tn *testnet.TestNet, n int, password string, dest string) error {
	var data string
	for i := 0; i < n; i++ {
		data += fmt.Sprintf("%s\n", password)
	}
	return util.LogError(helpers.CopyBytesToAllNewNodes(tn, data, dest))
}

// ExposeAccounts exposes the given accounts to the external services which require this data in
// order to function correctly.
func ExposeAccounts(tn *testnet.TestNet, accounts []*Account) {
	tn.BuildState.SetExt("accounts", ExtractAddresses(accounts))
	tn.BuildState.Set("accounts", accounts)
	for _, account := range accounts {
		tn.BuildState.SetExt(account.HexAddress(), map[string]string{
			"privateKey": account.HexPrivateKey(),
			"publicKey":  account.HexPublicKey(),
		})
	}
}

// ExposeEnodes provides a simple way to expose the enode addresses of the current nodes.
func ExposeEnodes(tn *testnet.TestNet, enodes []string) {
	tn.BuildState.SetExt(EnodeKey, enodes)
	tn.BuildState.Set(EnodeKey, enodes)
}

func UnlockAllAccounts(tn *testnet.TestNet, accounts []*Account, password string) error {
	return helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		tn.BuildState.Defer(func() { //Can happen eventually
			for _, account := range accounts {

				client.Run( //Doesn't really need to succeed, it is a nice to have, but not required.
					fmt.Sprintf(
						`curl -sS -X POST http://%s:8545 -H "Content-Type: application/json"  -d `+
							`'{ "method": "personal_unlockAccount", "params": ["%s","%s",0], "id": 3, "jsonrpc": "2.0" }'`,
						node.GetIP(), account.HexAddress(), password))

			}
		})
		return nil
	})
}
