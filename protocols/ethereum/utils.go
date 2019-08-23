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
	log "github.com/sirupsen/logrus"
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

//KEYS

const (
	//NetworkIDKey is the key for the network id
	NetworkIDKey = "networkId"
	//ChainIDKey is the key for chain id
	ChainIDKey = "chainId"
	//HomesteadBlockKey relates to the homestead block
	HomesteadBlockKey = "homesteadBlock"
	//EIP150BlockKey maps to the eip 150 block
	EIP150BlockKey = "eip150Block"
	//EIP155BlockKey maps to the eip 155 block
	EIP155BlockKey = "eip155Block"
	//EIP158BlockKey maps to the eip 158 block
	EIP158BlockKey = "eip158Block"
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

func GetExistingAccounts(tn *testnet.TestNet) []*Account {
	var out []*Account
	tn.BuildState.GetP("accounts", &out)
	return out
}

// ExposeEnodes provides a simple way to expose the enode addresses of the current nodes.
func ExposeEnodes(tn *testnet.TestNet, enodes []string) {
	enodesToStore := enodes
	if len(tn.Details) > 1 {
		var old []string
		tn.BuildState.GetP(EnodeKey,&old)
		enodesToStore = append(enodesToStore,old...)
	}
	log.WithFields(log.Fields{"enodes": enodesToStore}).Debug("updating the enodes in the store")
	tn.BuildState.SetExt(EnodeKey, enodesToStore)
	tn.BuildState.Set(EnodeKey, enodesToStore)	
}

//UnlockAllAccounts calls personal_unlockAccount for each account on every node, using the given password
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

//GetEnodes returns the enode addresses based on the nodes in the given testnet and the
//given accounts
func GetPeers(tn *testnet.TestNet, accounts []*Account) [][]string {
	var enodes []string
	tn.BuildState.GetP(EnodeKey, &enodes)
	out := [][]string{}
	for i, node := range tn.Nodes {
		for j,_ := range tn.Nodes {

			if i == j {
				log.WithFields(log.Fields{"num": node.GetAbsoluteNumber()}).Debug(
					"skipping node because already have it's node id")
				continue
			}
			out[i] = append(out[i],fmt.Sprintf("enode://%s@%s:%d", accounts[j].HexPublicKey(), node.IP, P2PPort))
		}
	}
	return out
}

//GetEnodes returns the enode addresses based on the nodes in the given testnet and the
//given accounts
func GetEnodes(tn *testnet.TestNet, accounts []*Account) []string {
	var enodes []string
	tn.BuildState.GetP(EnodeKey, &enodes)

	for i, node := range tn.Nodes {
		if len(enodes) > i {
			log.WithFields(log.Fields{"num": node.GetAbsoluteNumber()}).Debug(
				"skipping node because already have it's enode id")
			continue
		}
		enodes = append(enodes, fmt.Sprintf("enode://%s@%s:%d", accounts[i].HexPublicKey(), node.IP, P2PPort))
	}
	log.WithFields(log.Fields{"nodes":len(tn.Nodes),"builds":len(tn.Details)}).Debug("fetched the enodes")
	return enodes
}


//GetEnodes returns the enode addresses based on the nodes in the given testnet and the
//given accounts
func GetPreviousEnodes(tn *testnet.TestNet) []string {
	var enodes []string
	tn.BuildState.GetP(EnodeKey, &enodes)
	return enodes
}