/*
	Copyright 2019 Whiteblock Inc.
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
	"../../db"
	"../../ssh"
	"../../testnet"
	"../../util"
	"../ethereum"
	"../helpers"
	"../registrar"
	"encoding/json"
	"fmt"
	"github.com/Whiteblock/mustache"
	"regexp"
	"sync"
)

var conf *util.Config

const blockchain = "geth"

func init() {
	conf = util.GetConfig()
	alias := "ethereum"
	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterBuild(alias, build) //ethereum default to geth

	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterAddNodes(alias, add)

	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterServices(alias, GetServices)

	registrar.RegisterDefaults(blockchain, GetDefaults)
	registrar.RegisterDefaults(alias, GetDefaults)

	registrar.RegisterParams(blockchain, GetParams)
	registrar.RegisterParams(alias, GetParams)
}

const ethNetStatsPort = 3338

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
	/**Copy over the password file**/
	helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error { //ignore err
		_, err := client.DockerExec(node, "mkdir -p /geth")
		return err
	})

	/**Create the Password files**/
	{
		var data string
		for i := 1; i <= tn.LDD.Nodes; i++ {
			data += "password\n"
		}
		err = helpers.CopyBytesToAllNodes(tn, data, "/geth/passwd")
		if err != nil {
			return util.LogError(err)
		}
	}

	tn.BuildState.IncrementBuildProgress()

	/**Create the wallets**/
	tn.BuildState.SetBuildStage("Creating the wallets")

	accounts, err := ethereum.GenerateAccounts(tn.LDD.Nodes)
	if err != nil {
		return util.LogError(err)
	}
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		for i, account := range accounts {
			_, err := client.DockerExec(node, fmt.Sprintf("bash -c 'echo \"%s\" >> /geth/pk%d'", account.HexPrivateKey(), i))
			if err != nil {
				return util.LogError(err)
			}
			_, err = client.DockerExec(node,
				fmt.Sprintf("geth --datadir /geth/ --password /geth/passwd account import --password /geth/passwd /geth/pk%d", i))
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

	for i, account := range accounts {
		if i != 0 {
			unlock += ","
		}
		unlock += account.HexAddress()
	}

	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Creating the genesis block")
	err = createGenesisfile(ethconf, tn, accounts)
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Bootstrapping network")

	err = helpers.CopyToAllNodes(tn, "CustomGenesis.json", "/geth/")
	if err != nil {
		return util.LogError(err)
	}

	staticNodes := make([]string, tn.LDD.Nodes)

	tn.BuildState.SetBuildStage("Initializing geth")

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		//Load the CustomGenesis file
		_, err := client.DockerExec(node,
			fmt.Sprintf("geth --datadir /geth/ --networkid %d init /geth/CustomGenesis.json", ethconf.NetworkID))
		if err != nil {
			return util.LogError(err)
		}
		fmt.Printf("---------------------  CREATING block directory for NODE-%d ---------------------\n", node.GetAbsoluteNumber())
		gethResults, err := client.DockerExec(node,
			fmt.Sprintf("bash -c 'echo -e \"admin.nodeInfo.enode\\nexit\\n\" | "+
				"geth --rpc --datadir /geth/ --networkid %d console'", ethconf.NetworkID))
		if err != nil {
			return util.LogError(err)
		}
		//fmt.Printf("RAWWWWWWWWWWWW%s\n\n\n",gethResults)
		enodePattern := regexp.MustCompile(`enode:\/\/[A-z|0-9]+@(\[\:\:\]|([0-9]|\.)+)\:[0-9]+`)
		enode := enodePattern.FindAllString(gethResults, 1)[0]
		//fmt.Printf("ENODE fetched is: %s\n",enode)
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

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		tn.BuildState.IncrementBuildProgress()

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

		tn.BuildState.IncrementBuildProgress()
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.IncrementBuildProgress()

	err = setupEthNetStats(tn.GetFlatClients()[0])
	if err != nil {
		return util.LogError(err)
	}

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

/**
 * Setup Eth Net Stats on a server
 * @param  string    ip     The servers config
 */
func setupEthNetStats(client *ssh.Client) error {
	_, err := client.Run(fmt.Sprintf(
		"docker exec -d wb_service0 bash -c 'cd /eth-netstats && WS_SECRET=second PORT=%d npm start'", ethNetStatsPort))
	if err != nil {
		return util.LogError(err)
	}
	return nil
}

func setupEthNetIntelligenceAPI(tn *testnet.TestNet) error {
	return helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, node ssh.Node) error {
		absName := fmt.Sprintf("%s%d", conf.NodePrefix, node.GetAbsoluteNumber())
		sedCmd := fmt.Sprintf(`sed -i -r 's/"INSTANCE_NAME"(\s)*:(\s)*"(\S)*"/"INSTANCE_NAME"\t: "%s"/g' /eth-net-intelligence-api/app.json`, absName)
		sedCmd2 := fmt.Sprintf(`sed -i -r 's/"WS_SERVER"(\s)*:(\s)*"(\S)*"/"WS_SERVER"\t: "http:\/\/%s:%d"/g' /eth-net-intelligence-api/app.json`,
			util.GetGateway(server.SubnetID, node.GetAbsoluteNumber()), ethNetStatsPort)
		sedCmd3 := fmt.Sprintf(`sed -i -r 's/"RPC_HOST"(\s)*:(\s)*"(\S)*"/"RPC_HOST"\t: "%s"/g' /eth-net-intelligence-api/app.json`, node.GetIP())

		//sedCmd3 := fmt.Sprintf("docker exec -it %s sed -i 's/\"WS_SECRET\"(\\s)*:(\\s)*\"[A-Z|a-z|0-9| ]*\"/\"WS_SECRET\"\\t: \"second\"/g' /eth-net-intelligence-api/app.json",container)
		_, err := client.DockerMultiExec(node, []string{
			sedCmd,
			sedCmd2,
			sedCmd3})

		if err != nil {
			return util.LogError(err)
		}
		_, err = client.DockerExecd(node, "bash -c 'cd /eth-net-intelligence-api && pm2 start app.json'")
		if err != nil {
			return util.LogError(err)
		}
		tn.BuildState.IncrementBuildProgress()
		return nil
	})
}
