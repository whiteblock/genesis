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

//Package polkadot handles polkadot specific functionality
package polkadot

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/blockchains/ethereum"
	"github.com/whiteblock/genesis/blockchains/helpers"
	"github.com/whiteblock/genesis/blockchains/registrar"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"github.com/whiteblock/mustache"
	// "regexp"
	"sync"
)

var conf *util.Config

const blockchain = "polkadot"

func init() {
	conf = util.GetConfig()
	alias := "polkadot"
	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterBuild(alias, build) //ethereum default to geth

	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterAddNodes(alias, add)

	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterServices(alias, GetServices)

	registrar.RegisterDefaults(blockchain, helpers.DefaultGetDefaultsFn(blockchain))
	registrar.RegisterDefaults(alias, helpers.DefaultGetDefaultsFn(blockchain))

	registrar.RegisterParams(blockchain, helpers.DefaultGetParamsFn(blockchain))
	registrar.RegisterParams(alias, helpers.DefaultGetParamsFn(blockchain))
}

// build builds out a fresh new ethereum test network using geth
func build(tn *testnet.TestNet) error {
	mux := sync.Mutex{}
	ethconf, err := newConf(tn.LDD.Params)
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.SetBuildSteps(8 + (5 * tn.LDD.Nodes))

	tn.BuildState.IncrementBuildProgress()

	tn.BuildState.SetBuildStage("Distributing secrets")

	helpers.MkdirAllNodes(tn, "/polkadot")

	// {
	// 	/**Create the Static Nodes files**/
	// 	var nodeID string
	// 	for i := 1; i <= tn.LDD.Nodes; i++ {
	// 		data += "password\n"
	// 	}
	// 	/**Copy over the password file**/
	// 	err = helpers.CopyBytesToAllNodes(tn, data, "/geth/passwd")
	// 	if err != nil {
	// 		return util.LogError(err)
	// 	}
	// }

	var nodeIDList []string
	var nodeID string

	tn.BuildState.IncrementBuildProgress()

	/**Create the wallets**/
	tn.BuildState.SetBuildStage("Creating the wallets")

	accounts, err := ethereum.GenerateAccounts(tn.LDD.Nodes)
	if err != nil {
		return util.LogError(err)
	}
	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		_, err := client.DockerExecd(node, fmt.Sprintf("polkadot"))
		if err != nil {
			return util.LogError(err)
		}
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.IncrementBuildProgress()

	//read from directory to get nodeID

	// ('/ip4/%s/tcp/30333/p2p/%s', ip, id)

	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		nID, err := client.DockerExec(node, fmt.Sprintf("bash -c 'echo *'"))
		if err != nil {
			return util.LogError(err)
		}
		nodeIDList = append(nodeIDList, nID)
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}

	fmt.Println(nodeIDList)

	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Bootstrapping network")


	tn.BuildState.SetBuildStage("Initializing polkadot")

	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		//Load the CustomGenesis file
		_, err := client.DockerExec(node,
			fmt.Sprintf("polkadot %d", ethconf.NetworkID))
		if err != nil {
			return util.LogError(err)
		}
		log.WithFields(log.Fields{"node": node.GetAbsoluteNumber()}).Trace("creating block directory")

		// log.WithFields(log.Fields{"raw": gethResults}).Trace("grabbed raw enode info")
		// enodePattern := regexp.MustCompile(`(?m)Local node identity is: (.*)`)
		// enode := enodePattern.FindAllString(gethResults, 1)[0]
		// log.WithFields(log.Fields{"enode": enode}).Trace("parsed the enode")

		// enode = enodeAddressPattern.ReplaceAllString(enode, node.GetIP())

		mux.Lock()
		nodeIDList[node.GetAbsoluteNumber()] = nodeID
		mux.Unlock()

		tn.BuildState.IncrementBuildProgress()
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}

	out, err := json.Marshal(nodeIDList)
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Starting polkadot")
	// Copy static-nodes to every server
	err = helpers.CopyBytesToAllNodes(tn, string(out), "/geth/static-nodes.json")
	if err != nil {
		return util.LogError(err)
	}

	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		tn.BuildState.IncrementBuildProgress()

		/*
		gethCmd := fmt.Sprintf(
			`geth --datadir /geth/ --maxpeers %d --networkid %d --rpc --nodiscover --rpcaddr %s`+
				` --rpcapi "web3,db,eth,net,personal,miner,txpool" --rpccorsdomain "0.0.0.0" --mine --unlock="%s"`+
				` --password /geth/passwd --etherbase %s console  2>&1 | tee %s`,
			ethconf.MaxPeers,
			ethconf.NetworkID,
			node.GetIP(),
			unlock,
			accounts[node.GetAbsoluteNumber()].HexAddress(),
			conf.DockerOutputFile)

		_, err := client.DockerExecdit(node, fmt.Sprintf("bash -ic '%s'", gethCmd))
		if err != nil {
			return util.LogError(err)
		}
		*/

		tn.BuildState.IncrementBuildProgress()
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.IncrementBuildProgress()

	tn.BuildState.SetExt("accounts", ethereum.ExtractAddresses(accounts))

	for _, account := range accounts {
		tn.BuildState.SetExt(account.HexAddress(), map[string]string{
			"privateKey": account.HexPrivateKey(),
			"publicKey":  account.HexPublicKey(),
		})
	}
	return nil
}

/***************************************************************************************************************************/

// Add handles adding a node to the geth testnet
// TODO
func add(tn *testnet.TestNet) error {
	return nil
}

// MakeFakeAccounts creates ethereum addresses which can be marked as funded to produce a
// larger initial state
func MakeFakeAccounts(accs int) []string {
	out := make([]string, accs)
	for i := 1; i <= accs; i++ {
		out[i-1] = fmt.Sprintf("0x%.40x", i)
	}
	return out
}

/**
 * Create the custom genesis file for Ethereum
 * @param  *ethConf ethconf     The chain configuration
 * @param  []string wallets     The wallets to be allocated a balance
 */

func createGenesisfile(ethconf *ethConf, tn *testnet.TestNet, accounts []*ethereum.Account) error {

	genesis := map[string]interface{}{
		"chainId":        ethconf.NetworkID,
		"homesteadBlock": ethconf.HomesteadBlock,
		"eip155Block":    ethconf.Eip155Block,
		"eip158Block":    ethconf.Eip158Block,
		"difficulty":     fmt.Sprintf("0x0%X", ethconf.Difficulty),
		"gasLimit":       fmt.Sprintf("0x0%X", ethconf.GasLimit),
	}
	alloc := map[string]map[string]string{}
	for _, account := range accounts {
		alloc[account.HexAddress()] = map[string]string{
			"balance": ethconf.InitBalance,
		}
	}

	accs := MakeFakeAccounts(int(ethconf.ExtraAccounts))

	for _, wallet := range accs {
		alloc[wallet] = map[string]string{
			"balance": ethconf.InitBalance,
		}
	}
	genesis["alloc"] = alloc
	dat, err := helpers.GetBlockchainConfig("geth", 0, "genesis.json", tn.LDD)
	if err != nil {
		return util.LogError(err)
	}

	data, err := mustache.Render(string(dat), util.ConvertToStringMap(genesis))
	if err != nil {
		return util.LogError(err)
	}
	return tn.BuildState.Write("CustomGenesis.json", data)
}

