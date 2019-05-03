//Package parity handles parity specific functionality
package parity

import (
	"../../db"
	"../../ssh"
	"../../state"
	"../../testnet"
	"../../util"
	"../helpers"
	"../registrar"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
	blockchain := "parity"
	registrar.RegisterBuild(blockchain, Build)
	registrar.RegisterAddNodes(blockchain, Add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, GetDefaults)
	registrar.RegisterParams(blockchain, GetParams)
}

// Build builds out a fresh new ethereum test network using parity
func Build(tn *testnet.TestNet) ([]string, error) {
	mux := sync.Mutex{}
	pconf, err := newConf(tn.LDD.Params)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tn.BuildState.SetBuildSteps(9 + (7 * tn.LDD.Nodes))
	//Make the data directories
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		_, err := client.DockerExec(node, "mkdir -p /parity")
		return err
	})
	if err != nil {
		log.Println(err)
		return nil, err
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
			log.Println(err)
			return nil, err
		}
		tn.BuildState.IncrementBuildProgress()
	}

	/**Create the wallets**/
	wallets := make([]string, tn.LDD.Nodes)
	rawWallets := make([]string, tn.LDD.Nodes)
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		res, err := client.DockerExec(node, "parity --base-path=/parity/ --password=/parity/passwd account new")
		if err != nil {
			log.Println(err)
			return err
		}

		if len(res) == 0 {
			return fmt.Errorf("account new returned an empty response")
		}

		mux.Lock()
		wallets[node.GetAbsoluteNumber()] = res[:len(res)-1]
		mux.Unlock()

		res, err = client.DockerExec(node, "bash -c 'cat /parity/keys/ethereum/*'")
		if err != nil {
			log.Println(err)
			return err
		}
		tn.BuildState.IncrementBuildProgress()

		mux.Lock()
		rawWallets[node.GetAbsoluteNumber()] = strings.Replace(res, "\"", "\\\"", -1)
		mux.Unlock()
		return nil
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	/***********************************************************SPLIT************************************************************/
	switch pconf.Consensus {
	case "ethash":
		err = setupPOW(tn, pconf, wallets)
	case "poa":
		err = setupPOA(tn, pconf, wallets)
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}

	/***********************************************************SPLIT************************************************************/

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		for i, rawWallet := range rawWallets {
			_, err := client.DockerExec(node, fmt.Sprintf("bash -c 'echo \"%s\">/parity/account%d'", rawWallet, i))
			if err != nil {
				log.Println(err)
				return err
			}

			_, err = client.DockerExec(node,
				fmt.Sprintf("parity --base-path=/parity/ --chain /parity/spec.json --password=/parity/passwd account import /parity/account%d", i))
			if err != nil {
				log.Println(err)
				return err
			}
		}
		tn.BuildState.IncrementBuildProgress()
		return nil
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//util.Write("tmp/config.toml",configToml)
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		return client.DockerExecdLog(node,
			fmt.Sprintf(`parity --author=%s -c /parity/config.toml --chain=/parity/spec.json`, wallets[node.GetAbsoluteNumber()]))
	})
	if err != nil {
		log.Println(err)
		return nil, err
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
				log.Println(err)
				return err
			}
			var result map[string]interface{}

			err = json.Unmarshal([]byte(res), &result)
			if err != nil {
				log.Println(err)
				return err
			}
			fmt.Println(result)

			err = util.GetJSONString(result, "result", &enode)
			if err != nil {
				log.Println(err)
				return err
			}
		}
		tn.BuildState.IncrementBuildProgress()
		mux.Lock()
		enodes[node.GetAbsoluteNumber()] = enode
		mux.Unlock()
		return nil
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = peerAllNodes(tn, enodes)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tn.BuildState.IncrementBuildProgress()
	if pconf.Consensus == "ethash" {
		return nil, peerWithGeth(tn.GetFlatClients()[0], tn.BuildState, enodes)
	}
	return nil, nil
}

/***************************************************************************************************************************/

// Add handles adding a node to the geth testnet
// TODO
func Add(tn *testnet.TestNet) ([]string, error) {
	return nil, nil
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
				log.Println(err)
				return err
			}
		}
		return nil
	})
}

func setupPOA(tn *testnet.TestNet, pconf *parityConf, wallets []string) error {
	//Create the chain spec files
	spec, err := buildPoaSpec(pconf, tn.LDD, wallets)
	if err != nil {
		log.Println(err)
		return err
	}

	err = helpers.CopyBytesToAllNodes(tn, spec, "/parity/spec.json")
	if err != nil {
		log.Println(err)
		return err
	}

	//handle configuration file
	return helpers.CreateConfigs(tn, "/parity/config.toml",
		func(node ssh.Node) ([]byte, error) {
			configToml, err := buildPoaConfig(pconf, tn.LDD, wallets, "/parity/passwd", node.GetAbsoluteNumber())
			if err != nil {
				log.Println(err)
				return nil, err
			}
			return []byte(configToml), nil
		})
}

func setupPOW(tn *testnet.TestNet, pconf *parityConf, wallets []string) error {
	//Start up the geth node
	err := setupGeth(tn.GetFlatClients()[0], tn.BuildState, pconf, wallets)
	if err != nil {
		log.Println(err)
		return err
	}
	tn.BuildState.IncrementBuildProgress()

	//Create the chain spec files
	spec, err := buildSpec(pconf, tn.LDD, wallets)
	if err != nil {
		log.Println(err)
		return err
	}
	//create config file
	err = helpers.CreateConfigs(tn, "/parity/config.toml",
		func(node ssh.Node) ([]byte, error) {
			configToml, err := buildConfig(pconf, tn.LDD, wallets, "/parity/passwd", node.GetAbsoluteNumber())
			if err != nil {
				log.Println(err)
				return nil, err
			}
			return []byte(configToml), nil
		})

	//Copy over the config file, spec file, and the accounts
	return helpers.CopyBytesToAllNodes(tn,
		spec, "/parity/spec.json")
}

func setupGeth(client *ssh.Client, buildState *state.BuildState, pconf *parityConf, wallets []string) error {

	gethConf, err := gethSpec(pconf, wallets)
	if err != nil {
		log.Println(err)
		return err
	}

	err = buildState.Write("genesis.json", gethConf)
	if err != nil {
		log.Println(err)
		return err
	}

	err = client.Scp("genesis.json", "/home/appo/genesis.json")
	if err != nil {
		log.Println(err)
		return err
	}
	buildState.Defer(func() { client.Run("rm /home/appo/genesis.json") })

	buildState.IncrementBuildProgress()

	_, err = client.FastMultiRun(
		"docker exec wb_service0 mkdir -p /geth",
		"docker cp /home/appo/genesis.json wb_service0:/geth/",
		"docker exec wb_service0 bash -c 'echo second >> /geth/passwd'")

	res, err := client.Run("docker exec wb_service0 geth --datadir /geth/ --password /geth/passwd account new")
	if err != nil {
		log.Println(err)
		return err
	}
	buildState.IncrementBuildProgress()

	addressPattern := regexp.MustCompile(`\{[A-z|0-9]+\}`)
	addresses := addressPattern.FindAllString(res, -1)
	if len(addresses) < 1 {
		return fmt.Errorf("unable to get addresses")
	}
	address := addresses[0][1 : len(addresses[0])-1]

	_, err = client.Run(
		fmt.Sprintf("docker exec wb_service0 geth --datadir /geth/ --networkid %d init /geth/genesis.json", pconf.NetworkID))
	if err != nil {
		log.Println(err)
		return err
	}

	buildState.IncrementBuildProgress()

	_, err = client.Run(fmt.Sprintf(`docker exec -d wb_service0 geth --datadir /geth/ --networkid %d --rpc  --rpcaddr 0.0.0.0`+
		` --rpcapi "admin,web3,db,eth,net,personal,miner,txpool" --rpccorsdomain "0.0.0.0" --unlock="%s"`+
		` --password /geth/passwd --etherbase %s --nodiscover`, pconf.NetworkID, address, address))
	if err != nil {
		log.Println(err)
		return err
	}
	if !pconf.DontMine {
		_, err = client.KeepTryRun(
			`curl -sS -X POST http://172.30.0.2:8545 -H "Content-Type: application/json" ` +
				` -d '{ "method": "miner_start", "params": [8], "id": 3, "jsonrpc": "2.0" }'`)
	} else {
		_, err = client.KeepTryRun(
			`curl -sS -X POST http://172.30.0.2:8545 -H "Content-Type: application/json" ` +
				` -d '{ "method": "miner_stop", "params": [], "id": 3, "jsonrpc": "2.0" }'`)
	}
	return err
}

func peerWithGeth(client *ssh.Client, buildState *state.BuildState, enodes []string) error {
	for _, enode := range enodes {
		_, err := client.KeepTryRun(
			fmt.Sprintf(
				`curl -sS -X POST http://172.30.0.2:8545 -H "Content-Type: application/json" `+
					` -d '{ "method": "admin_addPeer", "params": ["%s"], "id": 1, "jsonrpc": "2.0" }'`,
				enode))
		buildState.IncrementBuildProgress()
		if err != nil {
			log.Println(err)
			return err
		}
	}

	buildState.IncrementBuildProgress()
	return nil
}
