//Package parity handles parity specific functionality
package parity

import (
	"../../db"
	"../../ssh"
	//"../../state"
	"../../testnet"
	"../../util"
	"../ethereum"
	"../helpers"
	"../registrar"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

var conf *util.Config

const blockchain = "parity"

func init() {
	conf = util.GetConfig()
	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, GetDefaults)
	registrar.RegisterParams(blockchain, GetParams)
	registrar.RegisterBlockchainSideCars(blockchain, []string{"geth"})
}

// build builds out a fresh new ethereum test network using parity
func build(tn *testnet.TestNet) error {
	mux := sync.Mutex{}
	pconf, err := newConf(tn.LDD.Params)
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.SetBuildSteps(9 + (7 * tn.LDD.Nodes))
	//Make the data directories
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		_, err := client.DockerExec(node, "mkdir -p /parity")
		return err
	})
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.IncrementBuildProgress()

	/**Create the Password file and copy it over**/
	{
		var data string
		for i := 1; i <= tn.LDD.Nodes; i++ {
			data += "second\n"
		}
		tn.BuildState.IncrementBuildProgress()
		err = helpers.CopyBytesToAllNodes(tn, data, "/parity/passwd")
		if err != nil {
			return util.LogError(err)
		}
		tn.BuildState.IncrementBuildProgress()
	}

	/**Create the wallets**/
	wallets := make([]string, tn.LDD.Nodes)
	rawWallets := make([]string, tn.LDD.Nodes)
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		res, err := client.DockerExec(node, "parity --base-path=/parity/ --password=/parity/passwd account new")
		if err != nil {
			return util.LogError(err)
		}

		if len(res) == 0 {
			return fmt.Errorf("account new returned an empty response")
		}

		mux.Lock()
		wallets[node.GetAbsoluteNumber()] = res[:len(res)-1]
		mux.Unlock()

		res, err = client.DockerExec(node, "bash -c 'cat /parity/keys/ethereum/*'")
		if err != nil {
			return util.LogError(err)
		}
		tn.BuildState.IncrementBuildProgress()

		mux.Lock()
		rawWallets[node.GetAbsoluteNumber()] = strings.Replace(res, "\"", "\\\"", -1)
		mux.Unlock()
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}
	/***********************************************************SPLIT************************************************************/
	switch pconf.Consensus {
	case "ethash":
		err = setupPOW(tn, pconf, wallets)
	case "poa":
		err = setupPOA(tn, pconf, wallets)
	}
	if err != nil {
		return util.LogError(err)
	}

	/***********************************************************SPLIT************************************************************/

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		for i, rawWallet := range rawWallets {
			_, err := client.DockerExec(node, fmt.Sprintf("bash -c 'echo \"%s\">/parity/account%d'", rawWallet, i))
			if err != nil {
				return util.LogError(err)
			}

			_, err = client.DockerExec(node,
				fmt.Sprintf("parity --base-path=/parity/ --chain /parity/spec.json --password=/parity/passwd account import /parity/account%d", i))
			if err != nil {
				return util.LogError(err)
			}
		}
		tn.BuildState.IncrementBuildProgress()
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}

	//util.Write("tmp/config.toml",configToml)
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		return client.DockerExecdLog(node,
			fmt.Sprintf(`parity --author=%s -c /parity/config.toml --chain=/parity/spec.json`, wallets[node.GetAbsoluteNumber()]))
	})
	if err != nil {
		return util.LogError(err)
	}
	//Start peering via curl
	time.Sleep(time.Duration(5 * time.Second))
	//Get the enode addresses
	enodes := make([]string, tn.LDD.Nodes)
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, node ssh.Node) error {
		enode := ""
		for len(enode) == 0 {
			res, err := client.KeepTryRun(
				fmt.Sprintf(
					`curl -sS -X POST http://%s:8545 -H "Content-Type: application/json" `+
						` -d '{ "method": "parity_enode", "params": [], "id": 1, "jsonrpc": "2.0" }'`,
					node.GetIP()))

			if err != nil {
				return util.LogError(err)
			}
			var result map[string]interface{}

			err = json.Unmarshal([]byte(res), &result)
			if err != nil {
				return util.LogError(err)
			}
			fmt.Println(result)

			err = util.GetJSONString(result, "result", &enode)
			if err != nil {
				return util.LogError(err)
			}
		}
		tn.BuildState.IncrementBuildProgress()
		mux.Lock()
		enodes[node.GetAbsoluteNumber()] = enode
		mux.Unlock()
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}
	storeGethParameters(tn, pconf, wallets, enodes)

	err = peerAllNodes(tn, enodes)
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.IncrementBuildProgress()
	return nil
}

/***************************************************************************************************************************/

// Add handles adding a node to the geth testnet
// TODO
func add(tn *testnet.TestNet) error {
	return nil
}

func peerAllNodes(tn *testnet.TestNet, enodes []string) error {
	return helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		for i, enode := range enodes {
			if i == node.GetAbsoluteNumber() {
				continue
			}
			_, err := client.Run(
				fmt.Sprintf(
					`curl -sS -X POST http://%s:8545 -H "Content-Type: application/json"  -d `+
						`'{ "method": "parity_addReservedPeer", "params": ["%s"], "id": 1, "jsonrpc": "2.0" }'`,
					node.GetIP(), enode))
			tn.BuildState.IncrementBuildProgress()
			if err != nil {
				return util.LogError(err)
			}
		}
		return nil
	})
}

func storeGethParameters(tn *testnet.TestNet, pconf *parityConf, wallets []string, enodes []string) {
	accounts, err := ethereum.GenerateAccounts(tn.LDD.Nodes)
	if err != nil {
		log.Println(err)
	}

	tn.BuildState.Set("networkID", pconf.NetworkID)
	tn.BuildState.Set("accounts", accounts)
	switch pconf.Consensus {
	case "ethash":
		tn.BuildState.Set("mine", true)
	case "poa":
		tn.BuildState.Set("mine", false)
	}

	tn.BuildState.Set("peers", enodes)

	tn.BuildState.Set("gethConf", map[string]interface{}{
		"networkID":   pconf.NetworkID,
		"initBalance": pconf.InitBalance,
		"difficulty":  fmt.Sprintf("0x%x", pconf.Difficulty),
		"gasLimit":    fmt.Sprintf("0x%x", pconf.GasLimit),
	})

	tn.BuildState.Set("wallets", wallets)
}

func setupPOA(tn *testnet.TestNet, pconf *parityConf, wallets []string) error {
	//Create the chain spec files
	spec, err := buildPoaSpec(pconf, tn.LDD, wallets)
	if err != nil {
		return util.LogError(err)
	}

	err = helpers.CopyBytesToAllNodes(tn, spec, "/parity/spec.json")
	if err != nil {
		return util.LogError(err)
	}

	//handle configuration file
	return helpers.CreateConfigs(tn, "/parity/config.toml",
		func(node ssh.Node) ([]byte, error) {
			configToml, err := buildPoaConfig(pconf, tn.LDD, wallets, "/parity/passwd", node.GetAbsoluteNumber())
			if err != nil {
				return nil, util.LogError(err)
			}
			return []byte(configToml), nil
		})
}

func setupPOW(tn *testnet.TestNet, pconf *parityConf, wallets []string) error {
	tn.BuildState.IncrementBuildProgress()

	//Create the chain spec files
	spec, err := buildSpec(pconf, tn.LDD, wallets)
	if err != nil {
		return util.LogError(err)
	}
	//create config file
	err = helpers.CreateConfigs(tn, "/parity/config.toml",
		func(node ssh.Node) ([]byte, error) {
			configToml, err := buildConfig(pconf, tn.LDD, wallets, "/parity/passwd", node.GetAbsoluteNumber())
			if err != nil {
				return nil, util.LogError(err)
			}
			return []byte(configToml), nil
		})

	//Copy over the config file, spec file, and the accounts
	return helpers.CopyBytesToAllNodes(tn,
		spec, "/parity/spec.json")
}
