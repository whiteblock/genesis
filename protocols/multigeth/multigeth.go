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

//Package multigeth handles multigeth specific functionality
package multigeth

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/protocols/ethereum"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/registrar"
	"github.com/whiteblock/genesis/protocols/services"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"github.com/whiteblock/mustache"
	"regexp"
	//"strings"
	"sync"
	"time"
)

var conf = util.GetConfig()

const (
	blockchain     = "multigeth"
	peeringRetries = 10
	password       = "password"
	passwordFile   = "/geth/passwd"
)

func init() {
	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterServices(blockchain, func() []services.Service { return nil })
	registrar.RegisterDefaults(blockchain, helpers.DefaultGetDefaultsFn(blockchain))
	registrar.RegisterParams(blockchain, helpers.DefaultGetParamsFn(blockchain))
}

func build(tn *testnet.TestNet) error {
	mux := sync.Mutex{}
	etcconf, err := newConf(tn)
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.SetBuildSteps(8 + (5 * tn.LDD.Nodes) + (tn.LDD.Nodes * (tn.LDD.Nodes - 1)))

	tn.BuildState.IncrementBuildProgress()

	tn.BuildState.SetBuildStage("Distributing secrets")

	helpers.MkdirAllNodes(tn, "/geth")

	tn.BuildState.IncrementBuildProgress()

	/**Create the wallets**/
	tn.BuildState.SetBuildStage("Creating the wallets")

	accounts, err := ethereum.GenerateAccounts(tn.LDD.Nodes + int(etcconf.ExtraAccounts))
	if err != nil {
		return util.LogError(err)
	}
	err = ethereum.CreateNPasswordFile(tn, len(accounts), password, passwordFile)
	if err != nil {
		return util.LogError(err)
	}

	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		for i, account := range accounts {
			_, err := client.DockerExec(node, fmt.Sprintf("bash -c 'echo \"%s\" >> /geth/pk%d'", account.HexPrivateKey(), i))
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
	tn.BuildState.Set("generatedAccs", accounts)

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
	err = createGenesisfile(etcconf, tn, accounts)
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Bootstrapping network")

	staticNodes := make([]string, tn.LDD.Nodes)

	tn.BuildState.SetBuildStage("Initializing geth")

	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		//Load the CustomGenesis file
		// _, err := client.DockerExec(node,
		// 	fmt.Sprintf("geth --datadir=/geth/ --network-id=%d --chain=/geth/chain.json", etcconf.NetworkID))

		// log.WithFields(log.Fields{"node": node.GetAbsoluteNumber()}).Trace("creating block directory")

		gethResults, err := client.DockerExec(node,
			fmt.Sprintf("bash -c 'echo -e \"admin.nodeInfo.enode\\nexit\\n\" | "+
				"geth --rpc --datadir=/geth/ --network-id=%d --chain=/geth/chain.json console'", etcconf.NetworkID))
		if err != nil {
			return util.LogError(err)
		}
		log.WithFields(log.Fields{"raw": gethResults}).Trace("grabbed raw enode info")
		enodePattern := regexp.MustCompile(`enode:\/\/[A-z|0-9]+@(\[\:\:\]|([0-9]|\.)+)\:[0-9]+`)
		enode := enodePattern.FindAllString(gethResults, 1)[0]
		log.WithFields(log.Fields{"enode": enode, "node": node.GetIP()}).Trace("parsed the enode")
		enodeAddressPattern := regexp.MustCompile(`\[\:\:\]|([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})`)
		enode = enodeAddressPattern.ReplaceAllString(enode, node.GetIP())

		mux.Lock()
		staticNodes[node.GetAbsoluteNumber()] = enode
		mux.Unlock()

		tn.BuildState.IncrementBuildProgress()
		return nil
	})

	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Starting geth")

	/*err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		log.WithFields(log.Fields{"node": node.GetAbsoluteNumber()}).Trace("adding accounts to right directory")

		cont, err := client.DockerExec(node,
			fmt.Sprintf("bash -c 'cd /geth/%s/keystore && cat $(ls | sed -n %dp)'", etcconf.Identity, node.GetAbsoluteNumber()+1))
		if err != nil {
			return util.LogError(err)
		}
		cont = strings.Replace(cont, "\"", "\\\"", -1)
		tn.BuildState.Set(fmt.Sprintf("node%dKey", node.GetAbsoluteNumber()), cont)

		tn.BuildState.IncrementBuildProgress()
		return nil
	})*/

	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		tn.BuildState.IncrementBuildProgress()

		gethCmd := fmt.Sprintf(
			`geth --datadir=/geth/ --maxpeers=%d --chain=/geth/chain.json --rpc --nodiscover --rpcaddr=%s`+
				` --rpcapi="admin,web3,db,eth,net,personal,miner,txpool" --rpccorsdomain="0.0.0.0" --mine --unlock="%s"`+
				` --password=/geth/passwd --etherbase=%s console  2>&1 | tee %s`,
			etcconf.MaxPeers,
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

	tn.BuildState.SetExt("networkID", etcconf.NetworkID)
	ethereum.ExposeAccounts(tn, accounts)
	tn.BuildState.SetExt("port", ethereum.RPCPort)
	tn.BuildState.SetExt("namespace", "eth")
	tn.BuildState.SetExt("password", password)
	tn.BuildState.Set("staticNodes", staticNodes)
	tn.BuildState.SetBuildStage("peering the nodes")
	time.Sleep(3 * time.Second)
	log.WithFields(log.Fields{"staticNodes": staticNodes}).Debug("peering")
	err = peerAllNodes(tn, staticNodes)
	if err != nil {
		return util.LogError(err)
	}
	return ethereum.UnlockAllAccounts(tn, accounts, password)

}

func add(tn *testnet.TestNet) error {
	return nil
}

/**
 * Create the custom genesis file for Ethereum
 * @param  *etcconf etcconf     The chain configuration
 * @param  []string wallets     The wallets to be allocated a balance
 */

func createGenesisfile(etcconf *ethConf, tn *testnet.TestNet, accounts []*ethereum.Account) error {

	alloc := map[string]map[string]string{}
	for _, account := range accounts {
		alloc[account.HexAddress()[2:]] = map[string]string{
			"balance": etcconf.InitBalance,
		}
	}

	consensusParams := map[string]interface{}{}
	switch etcconf.Consensus {
	case "clique":
		consensusParams["period"] = etcconf.BlockPeriodSeconds
		consensusParams["epoch"] = etcconf.Epoch
	case "ethash":
		consensusParams["difficulty"] = etcconf.Difficulty
	}

	genesis := map[string]interface{}{
		"network":            etcconf.NetworkID,
		"chainId":            etcconf.ChainID,
		"homesteadBlock":     etcconf.HomesteadBlock,
		"eip150Block":        etcconf.EIP150Block,
		"eip155Block":        etcconf.EIP155Block,
		"byzantiumBlock":     etcconf.ByzantiumBlock,
		"disposalBlock":      etcconf.DisposalBlock,
		"ecip1017EraRounds":  etcconf.ECIP1017EraRounds,
		"eip160FBlock":       etcconf.EIP160FBlock,
		"ecip1010PauseBlock": etcconf.ECIP1010PauseBlock,
		"ecip1010Length":     etcconf.ECIP1010Length,
		"gasLimit":           fmt.Sprintf("0x%x", etcconf.GasLimit),
		"difficulty":         fmt.Sprintf("0x0%x", etcconf.Difficulty),
		"mixhash":            etcconf.MixHash,
	}

	switch etcconf.Consensus {
	case "clique":
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
	tn.BuildState.Set("alloc", alloc)
	tn.BuildState.Set("etcconf", etcconf)

	return helpers.CreateConfigs(tn, "/geth/chain.json", func(node ssh.Node) ([]byte, error) {
		template, err := helpers.GetBlockchainConfig(blockchain, node.GetAbsoluteNumber(), "chain.json", tn.LDD)
		if err != nil {
			return nil, util.LogError(err)
		}

		data, err := mustache.Render(string(template), util.ConvertToStringMap(genesis))
		if err != nil {
			return nil, util.LogError(err)
		}
		return []byte(data), nil
	})
}

func peerAllNodes(tn *testnet.TestNet, enodes []string) error {
	return helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		for i, enode := range enodes {
			if i == node.GetAbsoluteNumber() {
				continue
			}
			var err error
			for i := 0; i < peeringRetries; i++ { //give it some extra tries
				_, err = client.KeepTryRun(
					fmt.Sprintf(
						`curl -sS -X POST http://%s:8545 -H "Content-Type: application/json"  -d `+
							`'{ "method": "admin_addPeer", "params": ["%s"], "id": 1, "jsonrpc": "2.0" }'`,
						node.GetIP(), enode))
				if err == nil {
					break
				}
				time.Sleep(1 * time.Second)
			}
			tn.BuildState.IncrementBuildProgress()
			if err != nil {
				return util.LogError(err)
			}
		}
		return nil
	})
}
