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

//Package geth handles geth specific functionality
package geth

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
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

var conf = util.GetConfig()

const (
	alias           = "ethereum"
	blockchain      = "geth"
	password        = "password"
	passwordFile    = "/geth/passwd"
	defaultMode     = "default"
	expansionMode   = "expand"
	genesisFileName = "CustomGenesis.json"
)

func init() {

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
	ethconf, err := newConf(tn)
	if err != nil {
		return util.LogError(err)
	}

	err = loadForExpand(tn, ethconf) //Prepare everything if it is large state
	if err != nil {
		return util.LogError(err)
	}

	validFlags := checkFlagsExist(tn)

	tn.BuildState.SetBuildSteps(8 + (5 * tn.LDD.Nodes))

	tn.BuildState.IncrementBuildProgress()

	tn.BuildState.SetBuildStage("Distributing secrets")

	helpers.MkdirAllNodes(tn, "/geth")

	err = ethereum.CreatePasswordFile(tn, password, passwordFile)
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.IncrementBuildProgress()

	/**Create the wallets**/
	tn.BuildState.SetBuildStage("Creating the wallets")

	accounts, err := getAccountPool(tn, ethconf, int(ethconf.ExtraAccounts)+tn.LDD.Nodes)
	if err != nil {
		return util.LogError(err)
	}
	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		for i, account := range accounts[:tn.LDD.Nodes] {
			_, err := client.DockerExec(node, fmt.Sprintf("sh -c 'echo \"%s\" > /geth/pk%d'", account.HexPrivateKey(), i))
			if err != nil {
				return util.LogError(err)
			}
			_, err = client.DockerExec(node,
				fmt.Sprintf("geth --datadir /geth/ --password /geth/passwd account import /geth/pk%d", i))
			if err != nil {
				util.LogError(err) //dont report the error
			}
		}
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.IncrementBuildProgress()

	err = handleGenesisFileDist(tn, ethconf, accounts)
	if err != nil {
		return util.LogError(err)
	}

	staticNodes := getEnodes(tn, accounts)

	tn.BuildState.SetBuildStage("Initializing geth")

	out, err := json.Marshal(staticNodes)
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Starting geth")
	//Copy static-nodes to every server
	err = helpers.CopyBytesToAllNodes(tn, string(out), "/geth/static-nodes.json")
	if err != nil {
		return util.LogError(err)
	}

	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		tn.BuildState.IncrementBuildProgress()
		account := accounts[node.GetAbsoluteNumber()]
		gethCmd := fmt.Sprintf(
			`geth --datadir /geth/ %s --rpc --nodiscover --rpcaddr 0.0.0.0`+
				` --miner.gasprice=1 --rpcapi "admin,web3,db,eth,net,personal,miner,txpool" --rpccorsdomain "0.0.0.0" --mine`+
				` --txpool.nolocals --rpcvhosts=* --port %d console  2>&1 | tee %s`,
			getExtraFlags(ethconf, account, validFlags[node.GetAbsoluteNumber()]), ethereum.P2PPort, conf.DockerOutputFile)

		_, err := client.DockerExecdit(node, fmt.Sprintf("sh -ic '%s'", gethCmd))
		tn.BuildState.IncrementBuildProgress()
		return util.LogError(err)
	})
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.Set("staticNodes", staticNodes)
	tn.BuildState.Set("geth-conf", *ethconf)

	tn.BuildState.SetExt("networkID", ethconf.NetworkID)
	tn.BuildState.SetExt("port", ethereum.RPCPort)
	helpers.SetFunctionalityGroup(tn, "eth")
	ethereum.ExposeAccounts(tn, accounts)
	ethereum.ExposeEnodes(tn, staticNodes)

	return nil
}

/***************************************************************************************************************************/

// Add handles adding a node to the geth testnet
func add(tn *testnet.TestNet) error {
	var ethconf *ethConf
	ok := tn.BuildState.GetP("geth-conf", ethconf)
	if !ok {
		var err error
		ethconf, err = newConf(tn)
		if err != nil {
			return util.LogError(err)
		}
	}

	helpers.MkdirAllNewNodes(tn, "/geth")

	err := ethereum.CreatePasswordFile(tn, password, passwordFile)
	if err != nil {
		return util.LogError(err)
	}

	validFlags := checkFlagsExist(tn)

	accounts, err := getAccountPool(tn, ethconf, int(ethconf.ExtraAccounts)+len(tn.Nodes))
	if err != nil {
		return util.LogError(err)
	}

	err = handleGenesisFileDist(tn, ethconf, accounts)
	if err != nil {
		return util.LogError(err)
	}

	staticNodes := getEnodes(tn, accounts)
	log.WithFields(log.Fields{"enodes": staticNodes}).Debug("got the enodes")

	tn.BuildState.SetBuildStage("Initializing geth")

	out, err := json.Marshal(staticNodes)
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Starting geth")
	//Copy static-nodes to every server
	err = helpers.CopyBytesToAllNewNodes(tn, string(out), "/geth/static-nodes.json")
	if err != nil {
		return util.LogError(err)
	}

	err = helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		for i, account := range accounts[:len(tn.Nodes)] {
			_, err := client.DockerExec(node, fmt.Sprintf("sh -c 'echo \"%s\" > /geth/pk%d'", account.HexPrivateKey(), i))
			if err != nil {
				return util.LogError(err)
			}
			_, err = client.DockerExec(node,
				fmt.Sprintf("geth --datadir /geth/ --password /geth/passwd account import /geth/pk%d", i))
			if err != nil {
				util.LogError(err) //dont report the error
			}
		}
		return nil
	})

	err = helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		tn.BuildState.IncrementBuildProgress()
		account := accounts[node.GetAbsoluteNumber()]
		gethCmd := fmt.Sprintf(
			`geth --datadir /geth/ %s --rpc --nodiscover --rpcaddr 0.0.0.0`+
				` --miner.gasprice=1 --rpcapi "admin,web3,db,eth,net,personal,miner,txpool" --rpccorsdomain "*" --mine`+
				` --txpool.nolocals --port %d console  2>&1 | tee %s`,
			getExtraFlags(ethconf, account, validFlags[node.GetAbsoluteNumber()]), ethereum.P2PPort, conf.DockerOutputFile)

		_, err := client.DockerExecdit(node, fmt.Sprintf("sh -ic '%s'", gethCmd))
		tn.BuildState.IncrementBuildProgress()
		return util.LogError(err)
	})
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.Set("geth-conf", *ethconf)
	ethereum.ExposeEnodes(tn, staticNodes)
	ethereum.ExposeAccounts(tn, accounts)
	return nil
}

/**
 * Create the custom genesis file for Ethereum
 * @param  *ethConf ethconf     The chain configuration
 * @param  []string wallets     The wallets to be allocated a balance
 */

func createGenesisfile(ethconf *ethConf, tn *testnet.TestNet, accounts []*ethereum.Account) (string, error) {
	var out string
	if ok := tn.BuildState.GetP("genesis-file", &out); ok {
		log.Debug("fetching the old genesis file")
		return out, nil
	}
	alloc := map[string]map[string]string{}
	for _, account := range accounts {
		alloc[account.HexAddress()] = map[string]string{
			"balance": ethconf.InitBalance,
		}
	}

	consensusParams := map[string]interface{}{}
	switch ethconf.Consensus {
	case "clique":
		consensusParams["period"] = ethconf.BlockPeriodSeconds
		consensusParams["epoch"] = ethconf.Epoch
	case "ethash":
		consensusParams["difficulty"] = ethconf.Difficulty
	}

	genesis := map[string]interface{}{
		"chainId":        ethconf.NetworkID,
		"homesteadBlock": ethconf.HomesteadBlock,
		"eip155Block":    ethconf.Eip155Block,
		"eip158Block":    ethconf.Eip158Block,
		"difficulty":     fmt.Sprintf("0x0%X", ethconf.Difficulty),
		"gasLimit":       fmt.Sprintf("0x0%X", ethconf.GasLimit),
		"consensus":      ethconf.Consensus,
	}

	switch ethconf.Consensus {
	case "clique":
		fallthrough
	case "ethash":
		extraData := "0x0000000000000000000000000000000000000000000000000000000000000000"
		//it does not work when there are multiple signers put into this extraData field
		/*
			for i := 0; i < len(accounts) && i < tn.LDD.Nodes; i++ {
				extraData += accounts[i].HexAddress()[2:]
			}
		*/
		extraData += accounts[0].HexAddress()[2:]
		extraData += "000000000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000"
		genesis["extraData"] = extraData
	}

	genesis["alloc"] = alloc
	genesis["consensusParams"] = consensusParams
	dat, err := helpers.GetGlobalBlockchainConfig(tn, "genesis.json")
	if err != nil {
		return "", util.LogError(err)
	}

	data, err := mustache.Render(string(dat), util.ConvertToStringMap(genesis))
	if err != nil {
		return "", util.LogError(err)
	}
	tn.BuildState.Set("genesis-file", data)
	return data, nil
}

func handleGenesisFileDist(tn *testnet.TestNet, ethconf *ethConf, accounts []*ethereum.Account) error {
	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Creating the genesis block")
	genesisFileLoc := fmt.Sprintf("/geth/%s", genesisFileName)

	genesisData, err := createGenesisfile(ethconf, tn, accounts)
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Bootstrapping network")
	hasGenesis := make([]bool, len(tn.Nodes))

	if ethconf.Mode != expansionMode {
		err := helpers.CopyBytesToAllNewNodes(tn, genesisData, genesisFileLoc)
		if err != nil {
			return util.LogError(err)
		}
	} else {
		//if it is expansion mode, do not create it if it does not exist

		mux := sync.Mutex{}

		helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
			_, err := client.DockerExec(node, fmt.Sprintf("test -f %s", genesisFileLoc))
			mux.Lock()
			hasGenesis[node.GetAbsoluteNumber()] = (err == nil)
			mux.Unlock()
			return nil
		})

		err := helpers.CreateConfigsNewNodes(tn, genesisFileLoc, func(node ssh.Node) ([]byte, error) {
			if hasGenesis[node.GetAbsoluteNumber()] {
				log.WithFields(log.Fields{"node": node.GetAbsoluteNumber()}).Debug("node already has a genesis file")
				return nil, nil
			}
			return []byte(genesisData), nil
		})
		if err != nil {
			return util.LogError(err)
		}
	}
	return helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		//Load the CustomGenesis file
		if ethconf.Mode != expansionMode {
			_, err := client.DockerExec(node,
				fmt.Sprintf("geth --datadir /geth/  init %s", genesisFileLoc))
			if err != nil {
				return util.LogError(err)
			}
		}
		log.WithFields(log.Fields{"node": node.GetAbsoluteNumber()}).Trace("creating block directory")
		tn.BuildState.IncrementBuildProgress()
		return nil
	})
}

func loadForExpand(tn *testnet.TestNet, ethconf *ethConf) error {
	if ethconf.Mode != expansionMode {
		return nil
	}
	masterNode := tn.Nodes[0]
	masterClient := tn.Clients[masterNode.Server]
	files := []string{"/important_data", "/important_info"}
	var data string
	var err error
	for _, file := range files {
		data, err = masterClient.DockerRead(masterNode, file, -1)
		if err == nil {
			break
		}

	}
	if err != nil {
		return util.LogError(err)
	}

	var clientData map[string]interface{}
	err = json.Unmarshal([]byte(data), &clientData)
	if err != nil {
		return util.LogError(err)
	}
	for key, value := range clientData {
		if key == "blockchain" || key == "blockchain_prefix" {
			continue
		}
		tn.BuildState.Set(key, value)
	}
	return nil
}

func getAccountPool(tn *testnet.TestNet, ethconf *ethConf, numOfAccounts int) ([]*ethereum.Account, error) {
	accounts := []*ethereum.Account{}
	rawPreGen, err := helpers.FetchPreGeneratedPrivateKeys(tn)
	if err != nil {
		log.Debug("There are not any pregenerated accounts availible")
	} else {
		accounts, err = ethereum.ImportAccounts(rawPreGen)
		if err != nil {
			return nil, util.LogError(err)
		}
	}
	if len(accounts) >= numOfAccounts {
		return accounts, nil
	}
	if ethconf.Mode == expansionMode && len(tn.Details) == 1 { //first build
		var tmp []string
		tn.BuildState.GetP("accounts", &tmp)
		for _, addr := range tmp {
			var accountData map[string]string
			ok := tn.BuildState.GetP(addr, &accountData)
			if !ok {
				log.WithFields(log.Fields{"address": addr}).Trace("skipping address without entry")
				continue
			}
			acc, err := ethereum.CreateAccountFromHex(accountData["privateKey"])
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Info("there was an error with the given private key")
			} else {
				accounts = append(accounts, acc)
			}
		}
	} else {
		var accs []*ethereum.Account
		tn.BuildState.GetP("accounts", &accs)
		accounts = append(accounts, accs...)
	}

	if len(accounts) >= numOfAccounts {
		log.Info("Fetched all the accounts from the build state store")
		return accounts, nil
	}
	fillerAccounts, err := ethereum.GenerateAccounts(numOfAccounts - len(accounts))
	if err != nil {
		return nil, util.LogError(err)
	}
	return append(accounts, fillerAccounts...), nil
}

func getExtraFlags(ethconf *ethConf, account *ethereum.Account, validFlags map[string]bool) string {
	out := fmt.Sprintf("--maxpeers %d --nodekeyhex %s",
		ethconf.MaxPeers, account.HexPrivateKey())
	out += fmt.Sprintf(" --verbosity %d", ethconf.Verbosity)

	if ethconf.Consensus == "ethash" {
		out += fmt.Sprintf(" --miner.gaslimit %d", ethconf.GasLimit)
		out += fmt.Sprintf(" --miner.gastarget %d", ethconf.GasLimit)
		out += fmt.Sprintf(" --miner.etherbase %s", account.HexAddress())
	}

	if ethconf.Mode == expansionMode {
		out += " --syncmode full"
	}

	if ethconf.Unlock {
		out += fmt.Sprintf(` --unlock="%s" --password /geth/passwd`, account.HexAddress())
		if validFlags["--allow-insecure-unlock"] {
			out += " --allow-insecure-unlock"
		}
	}
	if len(ethconf.Flags) > 0 {
		out += fmt.Sprintf(" %s",ethconf.Flags)
	}

	return out
}

func checkFlagsExist(tn *testnet.TestNet) []map[string]bool {
	flagsToCheck := []string{"--allow-insecure-unlock"}

	out := make([]map[string]bool, len(tn.Nodes))
	for i := range tn.Nodes {
		out[i] = map[string]bool{}
	}
	mux := sync.Mutex{}

	helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		for _, flag := range flagsToCheck {
			_, err := client.DockerExec(node, fmt.Sprintf("geth --help | grep -- '%s'", flag))
			mux.Lock()
			out[node.GetAbsoluteNumber()][flag] = (err == nil)
			mux.Unlock()
		}
		return nil
	})
	return out
}

func getEnodes(tn *testnet.TestNet, accounts []*ethereum.Account) []string {
	return ethereum.GetEnodes(tn, accounts)
}
