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

//Package geth handles the creation of the geth sidecar
package geth

import (
	"encoding/json"
	"fmt"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/protocols/ethereum"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/registrar"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"github.com/whiteblock/mustache"
	"sync"
)

var conf *util.Config

const sidecar = "geth"

func init() {
	conf = util.GetConfig()

	registrar.RegisterSideCar(sidecar, registrar.SideCar{
		Image: "gcr.io/whiteblock/geth:light",
		BuildStepsCalc: func(nodes int, _ int) int {
			return 5 * nodes
		},
	})
	registrar.RegisterBuildSideCar(sidecar, Build)
	registrar.RegisterAddSideCar(sidecar, Add)
}

// Build builds out a fresh new ethereum test network using geth
func Build(tn *testnet.Adjunct) error {
	var networkID int64
	var accounts []*ethereum.Account
	var mine bool
	var peers []string
	var conf map[string]interface{}
	var wallets []string

	tn.BuildState.GetP("networkID", &networkID)
	tn.BuildState.GetP("accounts", &accounts)
	tn.BuildState.GetP("mine", &mine)
	tn.BuildState.GetP("peers", &peers)
	tn.BuildState.GetP("gethConf", &conf)
	tn.BuildState.GetP("wallets", &wallets)

	err := helpers.AllNodeExecConSC(tn, func(client ssh.Client, server *db.Server, node ssh.Node) error {
		_, err := client.DockerExec(node, "mkdir -p /geth")
		return err
	})
	if err != nil {
		return util.LogError(err)
	}

	err = helpers.CreateConfigsSC(tn, "/geth/genesis.json", func(node ssh.Node) ([]byte, error) {
		defer tn.BuildState.IncrementSideCarProgress()
		if wallets != nil {
			return gethSpec(conf, wallets)
		}
		return gethSpec(conf, ethereum.ExtractAddresses(accounts))
	})
	if err != nil {
		return util.LogError(err)
	}

	/**Create the Password files**/
	{
		var data string
		for range accounts {
			data += "password\n"
		}
		err = helpers.CopyBytesToAllNodesSC(tn, data, "/geth/passwd")
		if err != nil {
			return util.LogError(err)
		}
	}

	out, err := json.Marshal(peers)
	if err != nil {
		return util.LogError(err)
	}

	//Copy static-nodes to every server
	err = helpers.CopyBytesToAllNodesSC(tn, string(out), "/geth/static-nodes.json")
	if err != nil {
		return util.LogError(err)
	}

	err = helpers.AllNodeExecConSC(tn, func(client ssh.Client, server *db.Server, node ssh.Node) error {
		_, err = client.DockerExec(node,
			fmt.Sprintf("geth --datadir /geth/ --networkid %d init /geth/genesis.json", networkID))
		if err != nil {
			return util.LogError(err)
		}
		tn.BuildState.IncrementSideCarProgress()

		wg := &sync.WaitGroup{}

		for i, account := range accounts {
			wg.Add(1)
			go func(account *ethereum.Account, i int) {
				defer wg.Done()

				_, err := client.DockerExec(node, fmt.Sprintf("sh -c 'echo \"%s\" >> /geth/pk%d'", account.HexPrivateKey(), i))
				if err != nil {
					tn.BuildState.ReportError(err)
					return
				}
				_, err = client.DockerExec(node,
					fmt.Sprintf("geth --datadir /geth/ --password /geth/passwd account import --password /geth/passwd /geth/pk%d", i))
				if err != nil {
					tn.BuildState.ReportError(err)
					return
				}

			}(account, i)
		}
		wg.Wait()
		tn.BuildState.IncrementSideCarProgress()
		unlock := ""
		for i, account := range accounts {
			if i != 0 {
				unlock += ","
			}
			unlock += account.HexAddress()
		}
		flags := ""

		if accounts != nil && len(accounts) > 0 {
			flags += fmt.Sprintf(" --etherbase %s", accounts[node.GetAbsoluteNumber()%len(accounts)].HexAddress())
		}

		_, err := client.DockerExecdit(node, fmt.Sprintf(` sh -ic 'geth --datadir /geth/ --rpc --rpcaddr 0.0.0.0`+
			` --rpcapi "admin,web3,miner,db,eth,net,personal,debug,txpool" --rpccorsdomain "0.0.0.0"%s --nodiscover --unlock="%s"`+
			` --password /geth/passwd --networkid %d --verbosity 5 console 2>&1 >> /output.log'`, flags, unlock, networkID))
		tn.BuildState.IncrementSideCarProgress()
		if err != nil {
			return util.LogError(err)
		}
		for i := 0; i < 10; i++ {
			if mine {
				_, err = client.KeepTryRun(
					fmt.Sprintf(`curl -sS -X POST http://%s:8545 -H "Content-Type: application/json" `+
						` -d '{ "method": "miner_start", "params": [1], "id": 3, "jsonrpc": "2.0" }'`, node.GetIP()))
			} else {
				_, err = client.KeepTryRun(
					fmt.Sprintf(`curl -sS -X POST http://%s:8545 -H "Content-Type: application/json" `+
						` -d '{ "method": "miner_stop", "params": [], "id": 3, "jsonrpc": "2.0" }'`, node.GetIP()))
			}
			if err == nil {
				break
			}
		}
		tn.BuildState.IncrementSideCarProgress()
		return util.LogError(err)
	})
	if err != nil {
		return util.LogError(err)
	}
	ips := make([]string, len(tn.Nodes))
	for i, node := range tn.Nodes {
		ips[i] = node.GetIP()
	}
	tn.BuildState.SetExt("geth", ips)

	return nil
}

/***************************************************************************************************************************/

// Add handles adding a node to the geth testnet
// TODO
func Add(tn *testnet.Adjunct) error {
	return nil
}

func gethSpec(conf map[string]interface{}, wallets []string) ([]byte, error) {
	accounts := make(map[string]interface{})
	for _, wallet := range wallets {
		accounts[wallet] = map[string]interface{}{
			"balance": conf["initBalance"],
		}
	}

	tmp := map[string]interface{}{
		"chainId":         conf["networkID"],
		"difficulty":      conf["difficulty"],
		"gasLimit":        conf["gasLimit"],
		"homesteadBlock":  0,
		"eip155Block":     10,
		"eip158Block":     10,
		"alloc":           accounts,
		"extraData":       conf["extraData"],
		"consensus":       conf["consensus"],
		"consensusParams": conf["consensusParams"],
	}
	filler := util.ConvertToStringMap(tmp)
	dat, err := helpers.GetStaticBlockchainConfig(sidecar, "genesis.json")
	if err != nil {
		return nil, util.LogError(err)
	}
	data, err := mustache.Render(string(dat), filler)
	return []byte(data), err
}
