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
	"regexp"
	"sync"
)

var conf *util.Config

const (
	blockchain      = "geth"
	ethNetStatsPort = 3338
	password        = "password"
	defaultMode     = "default"
	expansionMode   = "expand"
)

func init() {
	conf = util.GetConfig()
	alias := "ethereum"
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

	err = loadForExpand(tn, ethconf) //Prepare everything if it is large state
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.SetBuildSteps(8 + (5 * tn.LDD.Nodes))

	tn.BuildState.IncrementBuildProgress()

	tn.BuildState.SetBuildStage("Distributing secrets")

	helpers.MkdirAllNodes(tn, "/geth")

	{
		/**Create the Password files**/
		var data string
		for i := 1; i <= tn.LDD.Nodes; i++ {
			data += password + "\n"
		}
		/**Copy over the password file**/
		err = helpers.CopyBytesToAllNodes(tn, data, "/geth/passwd")
		if err != nil {
			return util.LogError(err)
		}
	}

	tn.BuildState.IncrementBuildProgress()

	/**Create the wallets**/
	tn.BuildState.SetBuildStage("Creating the wallets")

	accounts, err := getAccountPool(tn, int(ethconf.ExtraAccounts)+tn.LDD.Nodes)
	if err != nil {
		return util.LogError(err)
	}
	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		for i, account := range accounts[:tn.LDD.Nodes] {
			_, err := client.DockerExec(node, fmt.Sprintf("bash -c 'echo \"%s\" > /geth/pk%d'", account.HexPrivateKey(), i))
			if err != nil {
				return util.LogError(err)
			}
			_, err = client.DockerExec(node,
				fmt.Sprintf("geth --datadir /geth/ --password /geth/passwd account import /geth/pk%d", i))
			if err != nil {
				return util.LogError(err)
			}
		}
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.IncrementBuildProgress()
	unlock := ""

	for i, account := range accounts[:tn.LDD.Nodes] {
		if i != 0 {
			unlock += ","
		}
		unlock += account.HexAddress()
	}
	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Creating the genesis block")
	if ethconf.Mode != expansionMode {
		err = createGenesisfile(ethconf, tn, accounts)
		if err != nil {
			return util.LogError(err)
		}
	}
	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Bootstrapping network")

	if ethconf.Mode != expansionMode {
		err = helpers.CopyToAllNodes(tn, "CustomGenesis.json", "/geth/")
		if err != nil {
			return util.LogError(err)
		}
	}

	staticNodes := make([]string, tn.LDD.Nodes)

	tn.BuildState.SetBuildStage("Initializing geth")

	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		//Load the CustomGenesis file
		_, err := client.DockerExec(node,
			fmt.Sprintf("geth --datadir /geth/ --networkid %d init /geth/CustomGenesis.json", ethconf.NetworkID))
		if err != nil {
			return util.LogError(err)
		}
		log.WithFields(log.Fields{"node": node.GetAbsoluteNumber()}).Trace("creating block directory")
		gethResults, err := client.DockerExec(node,
			fmt.Sprintf("bash -c 'echo -e \"admin.nodeInfo.enode\\nexit\\n\" | "+
				"geth --rpc --datadir /geth/ --networkid %d console'", ethconf.NetworkID))
		if err != nil {
			return util.LogError(err)
		}
		log.WithFields(log.Fields{"raw": gethResults}).Trace("grabbed raw enode info")
		enodePattern := regexp.MustCompile(`enode:\/\/[A-z|0-9]+@(\[\:\:\]|([0-9]|\.)+)\:[0-9]+`)
		enode := enodePattern.FindAllString(gethResults, 1)[0]
		log.WithFields(log.Fields{"enode": enode}).Trace("parsed the enode")
		enodeAddressPattern := regexp.MustCompile(`\[\:\:\]|([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})`)

		enode = enodeAddressPattern.ReplaceAllString(enode, node.GetIP())

		mux.Lock()
		staticNodes[node.GetAbsoluteNumber()] = enode
		mux.Unlock()

		tn.BuildState.IncrementBuildProgress()
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}

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

		gethCmd := fmt.Sprintf(
			`geth --datadir /geth/ --maxpeers %d --networkid %d --rpc --nodiscover --rpcaddr %s`+
				` --rpcapi "admin,web3,db,eth,net,personal,miner,txpool" --rpccorsdomain "0.0.0.0" --mine --unlock="%s"`+
				` --password /geth/passwd --etherbase %s console  2>&1 | tee %s`,
			ethconf.MaxPeers,
			ethconf.NetworkID,
			node.GetIP(),
			unlock,
			accounts[node.GetAbsoluteNumber()].HexAddress(),
			conf.DockerOutputFile)

		_, err := client.DockerExecdit(node, fmt.Sprintf("bash -ic '%s'", gethCmd))
		tn.BuildState.IncrementBuildProgress()
		return util.LogError(err)
	})
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.IncrementBuildProgress()

	err = setupEthNetStats(tn.GetFlatClients()[0])
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.SetExt("networkID", ethconf.NetworkID)
	tn.BuildState.SetExt("accounts", ethereum.ExtractAddresses(accounts))
	tn.BuildState.SetExt("port", 8545)

	for _, account := range accounts {
		tn.BuildState.SetExt(account.HexAddress(), map[string]string{
			"privateKey": account.HexPrivateKey(),
			"publicKey":  account.HexPublicKey(),
		})
	}

	return setupEthNetIntelligenceAPI(tn)
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
	dat, err := helpers.GetBlockchainConfig(blockchain, 0, "genesis.json", tn.LDD)
	if err != nil {
		return util.LogError(err)
	}

	data, err := mustache.Render(string(dat), util.ConvertToStringMap(genesis))
	if err != nil {
		return util.LogError(err)
	}
	return tn.BuildState.Write("CustomGenesis.json", data)
}

/**
 * Setup Eth Net Stats on a server
 * @param  string    ip     The servers config
 */
func setupEthNetStats(client ssh.Client) error {
	_, err := client.Run(fmt.Sprintf(
		"docker exec -d wb_service0 bash -c 'cd /eth-netstats && WS_SECRET=second PORT=%d npm start'", ethNetStatsPort))
	if err != nil {
		return util.LogError(err)
	}
	return nil
}

func setupEthNetIntelligenceAPI(tn *testnet.TestNet) error {
	return helpers.AllNodeExecCon(tn, func(client ssh.Client, server *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()

		absName := fmt.Sprintf("%s%d", conf.NodePrefix, node.GetAbsoluteNumber())
		sedCmd := fmt.Sprintf(`sed -i -r 's/"INSTANCE_NAME"(\s)*:(\s)*"(\S)*"/"INSTANCE_NAME"\t: "%s"/g' /eth-net-intelligence-api/app.json`, absName)
		sedCmd2 := fmt.Sprintf(`sed -i -r 's/"WS_SERVER"(\s)*:(\s)*"(\S)*"/"WS_SERVER"\t: "http:\/\/%s:%d"/g' /eth-net-intelligence-api/app.json`,
			util.GetGateway(server.SubnetID, node.GetAbsoluteNumber()), ethNetStatsPort)
		sedCmd3 := fmt.Sprintf(`sed -i -r 's/"RPC_HOST"(\s)*:(\s)*"(\S)*"/"RPC_HOST"\t: "%s"/g' /eth-net-intelligence-api/app.json`, node.GetIP())

		//sedCmd3 := fmt.Sprintf("docker exec -it %s sed -i 's/\"WS_SECRET\"(\\s)*:(\\s)*\"[A-Z|a-z|0-9| ]*\"/\"WS_SECRET\"\\t: \"second\"/g' /eth-net-intelligence-api/app.json",container)
		_, err := client.DockerMultiExec(node, []string{sedCmd, sedCmd2, sedCmd3})

		if err != nil {
			return util.LogError(err)
		}
		_, err = client.DockerExecd(node, "bash -c 'cd /eth-net-intelligence-api && pm2 start app.json'")
		return util.LogError(err)
	})
}

func loadForExpand(tn *testnet.TestNet, ethconf *ethConf) error {
	if ethconf.Mode != expansionMode {
		return nil
	}
	masterNode := tn.Nodes[0]
	masterClient := tn.Clients[masterNode.Server]
	data, err := masterClient.DockerRead(masterNode, "/important_data", -1)
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

func getAccountPool(tn *testnet.TestNet, numOfAccounts int) ([]*ethereum.Account, error) {
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
		return accounts[:numOfAccounts], nil
	}
	var tmp []string
	tn.BuildState.GetP("accounts", &tmp)
	newAccs, err := ethereum.ImportAccounts(tmp)
	if err == nil {
		accounts = append(accounts, newAccs...)
	}
	if len(accounts) >= numOfAccounts {
		return accounts[:numOfAccounts], nil
	}
	fillerAccounts, err := ethereum.GenerateAccounts(numOfAccounts - len(accounts))
	if err != nil {
		return nil, util.LogError(err)
	}
	return append(accounts, fillerAccounts...), nil
}
