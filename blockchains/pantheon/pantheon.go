package pantheon

import (
	db "../../db"
	ssh "../../ssh"
	state "../../state"
	util "../../util"
	helpers "../helpers"
	"fmt"
	"github.com/Whiteblock/mustache"
	"log"
	"sync"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

/*
Build builds out a fresh new ethereum test network using pantheon
*/
func Build(details *db.DeploymentDetails, servers []db.Server, clients []*ssh.Client,
	buildState *state.BuildState) ([]string, error) {

	wg := sync.WaitGroup{}
	mux := sync.Mutex{}

	panconf, err := NewConf(details.Params)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	buildState.SetBuildSteps(6*details.Nodes + 2)
	buildState.IncrementBuildProgress()

	addresses := make([]string, details.Nodes)
	pubKeys := make([]string, details.Nodes)
	privKeys := make([]string, details.Nodes)
	rlpEncodedData := make([]string, details.Nodes)

	buildState.SetBuildStage("Setting Up Accounts")

	err = helpers.AllNodeExecCon(servers, buildState, func(serverNum int, localNodeNum int, absoluteNodeNum int) error {
		_, err := clients[serverNum].DockerExec(localNodeNum,
			"pantheon --data-path=/pantheon/data public-key export-address --to=/pantheon/data/nodeAddress")
		if err != nil {
			log.Println(err)
			return err
		}
		buildState.IncrementBuildProgress()
		_, err = clients[serverNum].DockerExec(localNodeNum,
			"pantheon --data-path=/pantheon/data public-key export --to=/pantheon/data/publicKey")
		if err != nil {
			log.Println(err)
			return err
		}

		addr, err := clients[serverNum].DockerExec(localNodeNum, "cat /pantheon/data/nodeAddress")
		if err != nil {
			log.Println(err)
			return err
		}

		addrs := string(addr[2:])

		mux.Lock()
		addresses[absoluteNodeNum] = addrs
		mux.Unlock()

		key, err := clients[serverNum].DockerExec(localNodeNum, "cat /pantheon/data/publicKey")
		if err != nil {
			log.Println(err)
			return err
		}
		buildState.IncrementBuildProgress()

		mux.Lock()
		pubKeys[absoluteNodeNum] = key[2:]
		mux.Unlock()

		privKey, err := clients[serverNum].DockerExec(localNodeNum, "cat /pantheon/data/key")
		if err != nil {
			log.Println(err)
			return err
		}
		mux.Lock()
		privKeys[absoluteNodeNum] = privKey[2:]
		mux.Unlock()

		_, err = clients[serverNum].DockerExec(localNodeNum, "bash -c 'echo \"[\\\""+addrs+"\\\"]\" >> /pantheon/data/toEncode.json'")
		if err != nil {
			log.Println(err)
			return err
		}

		_, err = clients[serverNum].DockerExec(localNodeNum, "mkdir /pantheon/genesis")
		if err != nil {
			log.Println(err)
			return err
		}

		// used for IBFT2 extraData
		_, err = clients[serverNum].DockerExec(localNodeNum,
			"pantheon rlp encode --from=/pantheon/data/toEncode.json --to=/pantheon/rlpEncodedExtraData")
		if err != nil {
			log.Println(err)
			return err
		}

		rlpEncoded, err := clients[serverNum].DockerExec(localNodeNum, "bash -c 'cat /pantheon/rlpEncodedExtraData'")
		if err != nil {
			log.Println(err)
			return err
		}

		mux.Lock()
		rlpEncodedData[absoluteNodeNum] = rlpEncoded
		mux.Unlock()

		fmt.Println(rlpEncodedData)

		buildState.IncrementBuildProgress()
		return err

	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	wg.Add(1)
	defer wg.Wait()
	go func() {
		defer wg.Done()
		/*
		   Set up a geth node, which is not part of the blockchain network,
		   to sign the transactions in place of the pantheon client. The pantheon
		   client does not support wallet management, so this acts as an easy work around.
		*/
		err := startGeth(clients[0], panconf, addresses, privKeys, buildState)
		if err != nil {
			log.Println(err)
			buildState.ReportError(err)
			return
		}
	}()

	/* Create Genesis File */
	buildState.SetBuildStage("Generating Genesis File")
	err = createGenesisfile(panconf, details, addresses, buildState, rlpEncodedData[0])
	if err != nil {
		log.Println(err)
		return nil, err
	}

	p2pPort := 30303
	enodes := "["
	var enodeAddress string
	for _, server := range servers {
		for i, ip := range server.Ips {
			enodeAddress = fmt.Sprintf("enode://%s@%s:%d",
				pubKeys[i],
				ip,
				p2pPort)
			if i < len(pubKeys)-1 {
				enodes = enodes + "\"" + enodeAddress + "\"" + ","
			} else {
				enodes = enodes + "\"" + enodeAddress + "\""
			}
			buildState.IncrementBuildProgress()
		}
	}
	enodes = enodes + "]"

	/* Create Static Nodes File */
	buildState.SetBuildStage("Setting Up Static Peers")
	buildState.IncrementBuildProgress()
	err = helpers.CopyBytesToAllNodes(servers, clients, buildState,
		enodes, "/pantheon/data/static-nodes.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = helpers.CreateConfigs(servers, clients, buildState, "/pantheon/config.toml",
		func(serverNum int, localNodeNum int, absoluteNodeNum int) ([]byte, error) {
			return helpers.GetBlockchainConfig("pantheon", absoluteNodeNum, "config.toml", details)
		})

	if err != nil {
		log.Println(err)
		return nil, err
	}
	/* Copy static-nodes & genesis files to each node */
	buildState.SetBuildStage("Distributing Files")
	err = helpers.CopyToAllNodes(servers, clients, buildState,
		"genesis.json", "/pantheon/genesis/genesis.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	/* Start the nodes */
	buildState.SetBuildStage("Starting Pantheon")
	err = helpers.AllNodeExecCon(servers, buildState, func(serverNum int, localNodeNum int, absoluteNodeNum int) error {
		err := clients[serverNum].DockerExecdLog(localNodeNum, fmt.Sprintf(
			`pantheon --config-file=/pantheon/config.toml --data-path=/pantheon/data --genesis-file=/pantheon/genesis/genesis.json  `+
				`--rpc-http-enabled --rpc-http-api="ADMIN,CLIQUE,DEBUG,EEA,ETH,IBFT,MINER,NET,TXPOOL,WEB3" `+
				` --p2p-port=%d --rpc-http-port=8545 --rpc-http-host="0.0.0.0" --host-whitelist=all --rpc-http-cors-origins="*"`,
			p2pPort))
		buildState.IncrementBuildProgress()
		return err
	})
	return privKeys, err
}

func createGenesisfile(panconf *PanConf, details *db.DeploymentDetails, address []string, buildState *state.BuildState, ibftExtraData string) error {
	alloc := map[string]map[string]string{}
	for _, addr := range address {
		alloc[addr] = map[string]string{
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
		consensusParams["fixeddifficulty"] = panconf.EthashDifficulty
	}

	genesis := map[string]interface{}{
		"chainId":    panconf.NetworkId,
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
		for _, addr := range address {
			extraData += addr
		}
		extraData += "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
		genesis["extraData"] = extraData
	}

	genesis["alloc"] = alloc
	genesis["consensusParams"] = consensusParams
	dat, err := helpers.GetBlockchainConfig("pantheon", 0, "genesis.json", details)
	if err != nil {
		log.Println(err)
		return err
	}

	data, err := mustache.Render(string(dat), util.ConvertToStringMap(genesis))
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Println("Writing Genesis File Locally")
	return buildState.Write("genesis.json", data)

}

func createStaticNodesFile(list string, buildState *state.BuildState) error {
	return buildState.Write("static-nodes.json", list)
}

func startGeth(client *ssh.Client, panconf *PanConf, addresses []string, privKeys []string, buildState *state.BuildState) error {
	serviceIps, err := util.GetServiceIps(GetServices())
	if err != nil {
		log.Println(err)
		return err
	}

	err = buildState.SetExt("signer_ip", serviceIps["geth"])
	if err != nil {
		log.Println(err)
		return err
	}

	err = buildState.SetExt("accounts", addresses)
	if err != nil {
		log.Println(err)
		return err
	}

	//Set up a geth node as a service to sign transactions
	client.Run(`docker exec wb_service0 mkdir /geth/`)

	unlock := ""
	for i, privKey := range privKeys {

		_, err = client.Run(`docker exec wb_service0 bash -c 'echo "second" >> /geth/passwd'`)
		if err != nil {
			log.Println(err)
			return err
		}

		_, err = client.Run(fmt.Sprintf(`docker exec wb_service0 bash -c 'echo -n "%s" > /geth/pk%d' `, privKey, i))
		if err != nil {
			log.Println(err)
			return err
		}

		_, err = client.Run(fmt.Sprintf(`docker exec wb_service0 geth --datadir /geth/ account import --password /geth/passwd /geth/pk%d`, i))
		if err != nil {
			log.Println(err)
			return err
		}

		if i != 0 {
			unlock += ","
		}
		unlock += "0x" + addresses[i]

	}
	_, err = client.Run(fmt.Sprintf(`docker exec -d wb_service0 geth --datadir /geth/ --rpc --rpcaddr 0.0.0.0`+
		` --rpcapi "admin,web3,db,eth,net,personal" --rpccorsdomain "0.0.0.0" --nodiscover --unlock="%s"`+
		` --password /geth/passwd --networkid %d`, unlock, panconf.NetworkId))
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
