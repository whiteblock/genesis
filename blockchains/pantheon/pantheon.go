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

//Package pantheon handles pantheon specific functionality
package pantheon

import (
	"fmt"
	"github.com/Whiteblock/genesis/blockchains/ethereum"
	"github.com/Whiteblock/genesis/blockchains/helpers"
	"github.com/Whiteblock/genesis/blockchains/registrar"
	"github.com/Whiteblock/genesis/db"
	"github.com/Whiteblock/genesis/ssh"
	"github.com/Whiteblock/genesis/state"
	"github.com/Whiteblock/genesis/testnet"
	"github.com/Whiteblock/genesis/util"
	"github.com/Whiteblock/mustache"
	"log"
	"sync"
)

var conf *util.Config

const blockchain = "pantheon"

func init() {
	conf = util.GetConfig()
	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, GetDefaults)
	registrar.RegisterParams(blockchain, GetParams)
	registrar.RegisterBlockchainSideCars(blockchain, []string{"geth", "orion"})

}

// build builds out a fresh new ethereum test network using pantheon
func build(tn *testnet.TestNet) error {
	mux := sync.Mutex{}

	panconf, err := newConf(tn.LDD.Params)
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.SetBuildSteps(6*tn.LDD.Nodes + 2)
	tn.BuildState.IncrementBuildProgress()

	rlpEncodedData := make([]string, tn.LDD.Nodes)
	accounts := make([]*ethereum.Account, tn.LDD.Nodes)

	extraAccounts, err := ethereum.GenerateAccounts(int(panconf.Accounts))
	if err != nil {
		return util.LogError(err)
	}
	accounts = append(accounts, extraAccounts...)

	tn.BuildState.SetBuildStage("Setting Up Accounts")

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {

		tn.BuildState.IncrementBuildProgress()
		_, err = client.DockerExec(node,
			"pantheon --data-path=/pantheon/data public-key export --to=/pantheon/data/publicKey")
		if err != nil {
			return util.LogError(err)
		}

		privKey, err := client.DockerRead(node, "/pantheon/data/key", -1)
		if err != nil {
			return util.LogError(err)
		}
		acc, err := ethereum.CreateAccountFromHex(privKey[2:])
		if err != nil {
			return util.LogError(err)
		}

		mux.Lock()
		accounts[node.GetAbsoluteNumber()] = acc
		mux.Unlock()
		tn.BuildState.IncrementBuildProgress()
		addr := acc.HexAddress()

		_, err = client.DockerExec(node, "bash -c 'echo \"[\\\""+addr[2:]+"\\\"]\" >> /pantheon/data/toEncode.json'")
		if err != nil {
			return util.LogError(err)
		}

		_, err = client.DockerExec(node, "mkdir /pantheon/genesis")
		if err != nil {
			return util.LogError(err)
		}

		// used for IBFT2 extraData
		_, err = client.DockerExec(node,
			"pantheon rlp encode --from=/pantheon/data/toEncode.json --to=/pantheon/rlpEncodedExtraData")
		if err != nil {
			return util.LogError(err)
		}

		rlpEncoded, err := client.DockerRead(node, "/pantheon/rlpEncodedExtraData", -1)
		if err != nil {
			return util.LogError(err)
		}

		mux.Lock()
		rlpEncodedData[node.GetAbsoluteNumber()] = rlpEncoded
		mux.Unlock()

		tn.BuildState.IncrementBuildProgress()
		return nil

	})
	if err != nil {
		return util.LogError(err)
	}

	/* Create Genesis File */
	tn.BuildState.SetBuildStage("Generating Genesis File")
	err = createGenesisfile(panconf, tn, accounts, rlpEncodedData[0])
	if err != nil {
		return util.LogError(err)
	}

	p2pPort := 30303
	enodes := "["
	for i, node := range tn.Nodes {
		enodeAddress := fmt.Sprintf("enode://%s@%s:%d",
			accounts[i].HexPublicKey(),
			node.IP,
			p2pPort)
		if i != 0 {
			enodes = enodes + ",\"" + enodeAddress + "\""
		} else {
			enodes = enodes + "\"" + enodeAddress + "\""
		}
		tn.BuildState.IncrementBuildProgress()
	}

	enodes = enodes + "]"

	/* Create Static Nodes File */
	tn.BuildState.SetBuildStage("Setting Up Static Peers")
	tn.BuildState.IncrementBuildProgress()
	err = helpers.CopyBytesToAllNodes(tn, enodes, "/pantheon/data/static-nodes.json")
	if err != nil {
		log.Println(err)
		return util.LogError(err)
	}

	err = helpers.CreateConfigs(tn, "/pantheon/config.toml", func(node ssh.Node) ([]byte, error) {
		return helpers.GetBlockchainConfig("pantheon", node.GetAbsoluteNumber(), "config.toml", tn.LDD)
	})
	if err != nil {
		return util.LogError(err)
	}
	/* Copy static-nodes & genesis files to each node */
	tn.BuildState.SetBuildStage("Distributing Files")
	err = helpers.CopyToAllNodes(tn, "genesis.json", "/pantheon/genesis/genesis.json")
	if err != nil {
		return util.LogError(err)
	}

	/* Start the nodes */
	tn.BuildState.SetBuildStage("Starting Pantheon")
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		flags, err := getExtraConfigurationFlags(tn, node, panconf)
		if err != nil {
			return util.LogError(err)
		}
		return client.DockerExecdLog(node, fmt.Sprintf(
			`pantheon --config-file=/pantheon/config.toml --logging=ALL --data-path=/pantheon/data --genesis-file=/pantheon/genesis/genesis.json  `+
				`--rpc-http-enabled --rpc-http-api="ADMIN,CLIQUE,DEBUG,EEA,ETH,IBFT,MINER,NET,TXPOOL,WEB3" `+
				` --p2p-port=%d --rpc-http-port=8545 --rpc-http-host="0.0.0.0" --host-whitelist=all --rpc-http-cors-origins="*"%s`,
			p2pPort, flags))
	})

	if err != nil {
		return util.LogError(err)
	}

	for _, account := range accounts {
		tn.BuildState.SetExt(account.HexAddress(), map[string]string{
			"privateKey": account.HexPrivateKey(),
			"publicKey":  account.HexPublicKey(),
		})
	}
	tn.BuildState.SetExt("accounts", ethereum.ExtractAddresses(accounts))
	tn.BuildState.Set("networkID", panconf.NetworkID)
	tn.BuildState.Set("accounts", accounts)
	tn.BuildState.Set("mine", false)
	tn.BuildState.Set("peers", []string{})

	tn.BuildState.Set("gethConf", map[string]interface{}{
		"networkID":   panconf.NetworkID,
		"initBalance": panconf.InitBalance,
		"difficulty":  fmt.Sprintf("0x%x", panconf.Difficulty),
		"gasLimit":    fmt.Sprintf("0x%x", panconf.GasLimit),
	})

	tn.BuildState.Set("wallets", ethereum.ExtractAddresses(accounts))
	return nil
}

func createGenesisfile(panconf *panConf, tn *testnet.TestNet, accounts []*ethereum.Account, ibftExtraData string) error {
	alloc := map[string]map[string]string{}
	for _, acc := range accounts {
		alloc[acc.HexAddress()] = map[string]string{
			"balance": panconf.InitBalance,
		}
	}
	consensusParams := map[string]interface{}{}
	switch panconf.Consensus {
	case "ibft2":
		fallthrough
	case "ibft":
		panconf.Consensus = "ibft2"
		fallthrough
	case "clique":
		consensusParams["blockPeriodSeconds"] = panconf.BlockPeriodSeconds
		consensusParams["epoch"] = panconf.Epoch
		consensusParams["requesttimeoutseconds"] = panconf.RequestTimeoutSeconds
	case "ethash":
		consensusParams["fixeddifficulty"] = panconf.FixedDifficulty
	}

	genesis := map[string]interface{}{
		"chainId":    panconf.NetworkID,
		"difficulty": fmt.Sprintf("0x0%X", panconf.Difficulty),
		"gasLimit":   fmt.Sprintf("0x0%X", panconf.GasLimit),
		"consensus":  panconf.Consensus,
	}

	switch panconf.Consensus {
	case "ibft2":
		genesis["extraData"] = ibftExtraData
	case "clique":
		fallthrough
	case "ethash":
		extraData := "0x0000000000000000000000000000000000000000000000000000000000000000"
		for i := 0; i < len(accounts) && i < tn.LDD.Nodes; i++ {
			extraData += accounts[i].HexAddress()[2:]
		}
		extraData += "000000000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000"
		genesis["extraData"] = extraData
	}

	genesis["alloc"] = alloc
	genesis["consensusParams"] = consensusParams
	dat, err := helpers.GetBlockchainConfig("pantheon", 0, "genesis.json", tn.LDD)
	if err != nil {
		return util.LogError(err)
	}

	data, err := mustache.Render(string(dat), util.ConvertToStringMap(genesis))
	if err != nil {
		return util.LogError(err)
	}
	fmt.Println("Writing Genesis File Locally")
	return tn.BuildState.Write("genesis.json", data)

}

func createStaticNodesFile(list string, buildState *state.BuildState) error {
	return buildState.Write("static-nodes.json", list)
}

func getExtraConfigurationFlags(tn *testnet.TestNet, node ssh.Node, pconf *panConf) (string, error) {
	out := ""
	if pconf.Orion {
		orionNode, err := tn.GetNodesSideCar(node, "orion")
		if err != nil {
			return out, util.LogError(err)
		}
		out += fmt.Sprintf(` --privacy-url="http://%s:8888"`, orionNode.GetIP())
	}

	return out, nil
}

// Add handles adding a node to the pantheon testnet
// TODO
func add(tn *testnet.TestNet) error {
	return nil
}
