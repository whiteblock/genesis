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

//Package pantheon handles pantheon specific functionality
package pantheon

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
	"strings"
	"sync"
)

var conf = util.GetConfig()

const (
	blockchain      = "pantheon"
	genesisFile     = "genesis.json"
	genesisFilePath = "/pantheon/genesis/"
	p2pPort         = 30303
)

func init() {
	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, helpers.DefaultGetDefaultsFn(blockchain))
	registrar.RegisterParams(blockchain, helpers.DefaultGetParamsFn(blockchain))
	registrar.RegisterBlockchainSideCars(blockchain, func(tn *testnet.TestNet) []string {
		return []string{"orion"}
	})
}

// build builds out a fresh new ethereum test network using pantheon
func build(tn *testnet.TestNet) error {
	genesisFileLoc := genesisFilePath + genesisFile

	mux := sync.Mutex{}

	panconf, err := newConf(tn.LDD.Params)
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.SetBuildSteps(5*tn.LDD.Nodes + 2)
	tn.BuildState.IncrementBuildProgress()

	accounts := make([]*ethereum.Account, tn.LDD.Nodes)

	extraAccounts, err := ethereum.GenerateAccounts(int(panconf.Accounts))
	if err != nil {
		return util.LogError(err)
	}
	accounts = append(accounts, extraAccounts...)

	tn.BuildState.SetBuildStage("Setting up accounts")

	helpers.MkdirAllNodes(tn, genesisFilePath)

	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {

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
		return nil

	})
	if err != nil {
		return util.LogError(err)
	}

	err = createGenesisfile(panconf, tn, accounts)
	if err != nil {
		return util.LogError(err)
	}

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
	err = helpers.CopyToAllNodes(tn, genesisFile, genesisFileLoc)
	if err != nil {
		return util.LogError(err)
	}

	/* Start the nodes */
	tn.BuildState.SetBuildStage("Starting Pantheon")
	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		flags, err := getExtraConfigurationFlags(tn, node, panconf, accounts)
		if err != nil {
			return util.LogError(err)
		}
		return client.DockerRunMainDaemon(node, fmt.Sprintf(
			`pantheon --config-file=/pantheon/config.toml --data-path=/pantheon/data --genesis-file=%s  `+
				`--rpc-http-enabled --rpc-http-api="ADMIN,CLIQUE,DEBUG,EEA,ETH,IBFT,MINER,NET,TXPOOL,WEB3" `+
				` --p2p-port=%d --rpc-http-port=8545 --rpc-http-host="0.0.0.0" --host-whitelist=all %s`,
			genesisFileLoc, p2pPort, flags))
	})

	if err != nil {
		return util.LogError(err)
	}

	ethereum.ExposeAccounts(tn, accounts)
	tn.BuildState.SetExt("port", ethereum.RPCPort)
	tn.BuildState.Set("networkID", panconf.NetworkID)
	tn.BuildState.SetExt("networkID", panconf.NetworkID)
	helpers.SetFunctionalityGroup(tn, "eth")
	tn.BuildState.Set("wallets", ethereum.ExtractAddresses(accounts))
	return nil
}

func createGenesisfile(panconf *panConf, tn *testnet.TestNet, accounts []*ethereum.Account) error {
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
		var err error
		genesis["extraData"], err = getIBFTExtraData(tn, panconf, accounts)
		if err != nil {
			return util.LogError(err)
		}
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
	dat, err := helpers.GetGlobalBlockchainConfig(tn, genesisFile)
	if err != nil {
		return util.LogError(err)
	}

	data, err := mustache.Render(string(dat), util.ConvertToStringMap(genesis))
	if err != nil {
		return util.LogError(err)
	}
	log.WithFields(log.Fields{"file": genesisFile}).Trace("writing the genesis file")
	return util.LogError(tn.BuildState.Write(genesisFile, data))

}

func getIBFTExtraData(tn *testnet.TestNet, panconf *panConf, accounts []*ethereum.Account) (string, error) {
	if panconf.Validators > tn.LDD.Nodes {
		return "", util.LogError(fmt.Errorf("invalid number of validators(%d), cannot be greater than number of nodes (%d)",
			panconf.Validators, tn.LDD.Nodes))
	}
	validatorAccounts := []string{}
	/* Create Genesis File */
	tn.BuildState.SetBuildStage("Generating Genesis File")
	for i := 0; i < panconf.Validators; i++ {

		toAdd := accounts[i].HexAddress()

		log.WithFields(log.Fields{"address": toAdd, "index": i}).Trace("adding validator address")
		if strings.HasPrefix(toAdd, "0x") {
			toAdd = toAdd[2:]
		}
		validatorAccounts = append(validatorAccounts, toAdd)
	}
	vaJSON, err := json.Marshal(validatorAccounts)
	if err != nil {
		return "", util.LogError(err)
	}
	//add the extra escapes by marshaling it as a string again
	vaJSON, err = json.Marshal(string(vaJSON))
	if err != nil {
		return "", util.LogError(err)
	}

	_, err = tn.Clients[tn.Nodes[0].Server].DockerExec(tn.Nodes[0],
		fmt.Sprintf("bash -c 'echo %s > /pantheon/rlpValidators.json'", string(vaJSON)))
	if err != nil {
		return "", util.LogError(err)
	}

	ibftExtraData, err := tn.Clients[tn.Nodes[0].Server].DockerExec(
		tn.Nodes[0], "pantheon rlp encode --from=/pantheon/rlpValidators.json")
	if err != nil {
		return "", util.LogError(err)
	}

	log.WithFields(log.Fields{"ibftExtras": ibftExtraData}).Debug("got the validator address list in rlp")
	return strings.Trim(ibftExtraData, "\n\r"), nil
}

func getExtraConfigurationFlags(tn *testnet.TestNet, node ssh.Node, pconf *panConf, accounts []*ethereum.Account) (string, error) {
	out := ""
	if pconf.Orion {
		orionNode, err := tn.GetNodesSideCar(node, "orion")
		if err != nil {
			return out, util.LogError(err)
		}
		out += fmt.Sprintf(` --privacy-url="http://%s:8888"`, orionNode.GetIP())
	}

	switch pconf.Consensus {
	case "ibft2":
	case "clique":
	case "ethash":
		out += fmt.Sprintf(` --miner-coinbase="%s"`, accounts[node.GetAbsoluteNumber()].HexAddress())
	}

	return out, nil
}

// Add handles adding a node to the pantheon testnet
// TODO
func add(tn *testnet.TestNet) error {
	return nil
}
