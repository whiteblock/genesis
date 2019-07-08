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
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
)

const (
	//P2PPort is the standard p2p port for all the ethereum clients
	P2PPort = 30303
	//RPCPort is the standard json rpc over http port
	RPCPort = 8545
)

//CreatePasswordFile turns the process of creating a password file into a single function call
func CreatePasswordFile(tn *testnet.TestNet, password string, dest string) error {
	var data string
	for i := 1; i <= tn.LDD.Nodes; i++ {
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
