package parity

import (
	db "../../db"
	state "../../state"
	util "../../util"
	helpers "../helpers"
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
}

/**
 * Build the Ethereum Test Network
 * @param  map[string]interface{}   data    Configuration Data for the network
 * @param  int      nodes       The number of nodes in the network
 * @param  []Server servers     The list of servers passed from build
 */
func Build(details db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
	buildState *state.BuildState) ([]string, error) {

	mux := sync.Mutex{}
	pconf, err := NewConf(details.Params)
	fmt.Printf("%#v\n", *pconf)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	buildState.SetBuildSteps(8 + (10 * details.Nodes))
	//Make the data directories
	err = helpers.AllNodeExecCon(servers, buildState, func(serverNum int, localNodeNum int, absoluteNodeNum int) error {
		_, err := clients[serverNum].DockerExec(localNodeNum, "mkdir -p /parity")
		return err
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	buildState.IncrementBuildProgress()

	/**Create the Password file**/
	{
		var data string
		for i := 1; i <= details.Nodes; i++ {
			data += "second\n"
		}
		err = buildState.Write("passwd", data)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	buildState.IncrementBuildProgress()
	/**Copy over the password file**/
	err = helpers.CopyToServers(servers, clients, buildState, "passwd", "/home/appo/passwd")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = helpers.AllNodeExecCon(servers, buildState, func(serverNum int, localNodeNum int, absoluteNodeNum int) error {
		buildState.IncrementBuildProgress()
		return clients[serverNum].DockerCp(localNodeNum, "/home/appo/passwd", "/parity/")
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	/**Create the wallets**/
	wallets := make([]string, details.Nodes)
	rawWallets := make([]string, details.Nodes)
	err = helpers.AllNodeExecCon(servers, buildState, func(serverNum int, localNodeNum int, absoluteNodeNum int) error {
		res, err := clients[serverNum].DockerExec(localNodeNum,
			fmt.Sprintf("parity --base-path=/parity/ --password=/parity/passwd account new"))
		if err != nil {
			log.Println(err)
			return err
		}

		if len(res) == 0 {
			return fmt.Errorf("account new returned an empty response")
		}

		address := res[:len(res)-1]

		mux.Lock()
		wallets[absoluteNodeNum] = address
		mux.Unlock()

		res, err = clients[serverNum].DockerExec(localNodeNum, "bash -c 'cat /parity/keys/ethereum/*'")
		if err != nil {
			log.Println(err)
			return err
		}
		buildState.IncrementBuildProgress()

		mux.Lock()
		rawWallets[absoluteNodeNum] = strings.Replace(res, "\"", "\\\"", -1)
		mux.Unlock()
		return nil
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//Start up the geth node
	{
		gethConf, err := GethSpec(pconf, wallets)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		err = buildState.Write("genesis.json", gethConf)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		err = clients[0].Scp("genesis.json", "/home/appo/genesis.json")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer clients[0].Run("rm /home/appo/genesis.json")

		buildState.IncrementBuildProgress()

		_, err = clients[0].Run("docker exec wb_service0 mkdir -p /geth")
		if err != nil {
			log.Println(err)
			return nil, err
		}

		_, err = clients[0].Run("docker cp /home/appo/genesis.json wb_service0:/geth/")
		if err != nil {
			log.Println(err)
			return nil, err
		}

		_, err = clients[0].Run("docker exec wb_service0 bash -c 'echo second >> /geth/passwd'")
		if err != nil {
			log.Println(err)
			return nil, err
		}

		buildState.IncrementBuildProgress()

		res, err := clients[0].Run("docker exec wb_service0 geth --datadir /geth/ --password /geth/passwd account new")
		if err != nil {
			log.Println(err)
			return nil, err
		}

		addressPattern := regexp.MustCompile(`\{[A-z|0-9]+\}`)
		addresses := addressPattern.FindAllString(res, -1)
		if len(addresses) < 1 {
			return nil, fmt.Errorf("Unable to get addresses")
		}
		address := addresses[0]
		address = address[1 : len(address)-1]

		_, err = clients[0].Run(
			fmt.Sprintf("docker exec wb_service0 geth --datadir /geth/ --networkid %d init /geth/genesis.json", pconf.NetworkId))
		if err != nil {
			log.Println(err)
			return nil, err
		}

		buildState.IncrementBuildProgress()

		_, err = clients[0].Run(fmt.Sprintf(`docker exec -d wb_service0 geth --datadir /geth/ --networkid %d --rpc  --rpcaddr 0.0.0.0`+
			` --rpcapi "admin,web3,db,eth,net,personal,miner,txpool" --rpccorsdomain "0.0.0.0" --mine --unlock="%s"`+
			` --password /geth/passwd --etherbase %s --nodiscover`, pconf.NetworkId, address, address))
		if err != nil {
			log.Println(err)
			return nil, err
		}
		_, err = clients[0].KeepTryRun(
			`curl -sS -X POST http://172.30.0.2:8545 -H "Content-Type: application/json" ` +
				` -d '{ "method": "miner_start", "params": [8], "id": 3, "jsonrpc": "2.0" }'`)

		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	buildState.IncrementBuildProgress()

	//Create the chain spec files
	spec, err := BuildSpec(pconf, details.Files, wallets)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = buildState.Write("spec.json", spec)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//create config file
	configToml, err := BuildConfig(pconf, details.Files, wallets, "/parity/passwd")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = buildState.Write("config.toml", configToml)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//Copy over the config file, spec file, and the accounts
	err = helpers.CopyAllToServers(servers, clients, buildState,
		"config.toml", "/home/appo/config.toml",
		"spec.json", "/home/appo/spec.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = helpers.AllNodeExecCon(servers, buildState, func(serverNum int, localNodeNum int, absoluteNodeNum int) error {
		err := clients[serverNum].DockerCp(localNodeNum, "/home/appo/spec.json", "/parity/")
		if err != nil {
			log.Println(err)
			return err
		}

		buildState.IncrementBuildProgress()

		err = clients[serverNum].DockerCp(localNodeNum, "/home/appo/config.toml", "/parity/")
		if err != nil {
			log.Println(err)
			return err
		}

		buildState.IncrementBuildProgress()

		for i, rawWallet := range rawWallets {

			_, err = clients[serverNum].DockerExec(localNodeNum, fmt.Sprintf("bash -c 'echo \"%s\">>/parity/account%d'", rawWallet, i))
			if err != nil {
				log.Println(err)
				return err
			}
			defer clients[serverNum].DockerExec(localNodeNum, fmt.Sprintf("rm /parity/account%d", i))

			_, err = clients[serverNum].DockerExec(localNodeNum,
				fmt.Sprintf("parity --base-path=/parity/ --chain /parity/spec.json --password=/parity/passwd account import /parity/account%d", i))
			if err != nil {
				log.Println(err)
				return err
			}
		}
		buildState.IncrementBuildProgress()
		return nil
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//util.Write("tmp/config.toml",configToml)
	err = helpers.AllNodeExecCon(servers, buildState, func(serverNum int, localNodeNum int, absoluteNodeNum int) error {
		parityCmd := fmt.Sprintf(`parity --author=%s -c /parity/config.toml --chain=/parity/spec.json`, wallets[absoluteNodeNum])
		defer buildState.IncrementBuildProgress()
		return clients[serverNum].DockerExecdLog(localNodeNum, parityCmd)
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//Start peering via curl
	time.Sleep(time.Duration(5 * time.Second))
	//Get the enode addresses
	enodes := make([]string, details.Nodes)
	err = helpers.AllNodeExecCon(servers, buildState, func(serverNum int, localNodeNum int, absoluteNodeNum int) error {
		enode := ""
		for len(enode) == 0 {
			ip := servers[serverNum].Ips[localNodeNum]
			res, err := clients[serverNum].KeepTryRun(
				fmt.Sprintf(
					`curl -sS -X POST http://%s:8545 -H "Content-Type: application/json" `+
						` -d '{ "method": "parity_enode", "params": [], "id": 1, "jsonrpc": "2.0" }'`,
					ip))

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
		buildState.IncrementBuildProgress()
		mux.Lock()
		enodes[absoluteNodeNum] = enode
		mux.Unlock()
		return nil
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = helpers.AllNodeExecCon(servers, buildState, func(serverNum int, localNodeNum int, absoluteNodeNum int) error {
		ip := servers[serverNum].Ips[localNodeNum]
		for i, enode := range enodes {
			if i == absoluteNodeNum {
				continue
			}
			_, err := clients[serverNum].KeepTryRun(
				fmt.Sprintf(
					`curl -sS -X POST http://%s:8545 -H "Content-Type: application/json"  -d `+
						`'{ "method": "parity_addReservedPeer", "params": ["%s"], "id": 1, "jsonrpc": "2.0" }'`,
					ip, enode))
			buildState.IncrementBuildProgress()
			if err != nil {
				log.Println(err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	buildState.IncrementBuildProgress()

	for _, enode := range enodes {
		_, err := clients[0].KeepTryRun(
			fmt.Sprintf(
				`curl -sS -X POST http://172.30.0.2:8545 -H "Content-Type: application/json" `+
					` -d '{ "method": "admin_addPeer", "params": ["%s"], "id": 1, "jsonrpc": "2.0" }'`,
				enode))
		buildState.IncrementBuildProgress()
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	_, err = clients[0].KeepTryRun(
		`curl -sS -X POST http://172.30.0.2:8545 -H "Content-Type: application/json" ` +
			` -d '{ "method": "miner_start", "params": [8], "id": 4, "jsonrpc": "2.0" }'`)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	buildState.IncrementBuildProgress()

	return nil, nil
}

/***************************************************************************************************************************/

func Add(details db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
	newNodes map[int][]string, buildState *state.BuildState) ([]string, error) {
	return nil, nil
}
