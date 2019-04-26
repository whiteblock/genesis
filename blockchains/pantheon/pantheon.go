//Package pantheon handles artemis specific functionality
package pantheon

import (
	"../../db"
	"../../ssh"
	"../../state"
	"../../testnet"
	"../../util"
	"../helpers"
	"fmt"
	"github.com/Whiteblock/mustache"
	"log"
	"regexp"
	"sync"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

// Build builds out a fresh new ethereum test network using pantheon
func Build(tn *testnet.TestNet) ([]string, error) {
	buildState := tn.BuildState
	wg := sync.WaitGroup{}
	mux := sync.Mutex{}

	panconf, err := newConf(tn.LDD.Params)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	buildState.SetBuildSteps(6*tn.LDD.Nodes + 2)
	buildState.IncrementBuildProgress()

	addresses := make([]string, tn.LDD.Nodes)
	pubKeys := make([]string, tn.LDD.Nodes)
	privKeys := make([]string, tn.LDD.Nodes)
	rlpEncodedData := make([]string, tn.LDD.Nodes)

	extraAccChan := make(chan []string)

	go func() {
		accs, err := prepareGeth(tn.GetFlatClients()[0], panconf, tn.LDD.Nodes, buildState)
		if err != nil {
			log.Println(err)
			buildState.ReportError(err)
			extraAccChan <- nil
			return
		}
		extraAccChan <- accs
	}()

	buildState.SetBuildStage("Setting Up Accounts")

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, localNodeNum int, absoluteNodeNum int) error {
		_, err := client.DockerExec(localNodeNum,
			"pantheon --data-path=/pantheon/data public-key export-address --to=/pantheon/data/nodeAddress")
		if err != nil {
			log.Println(err)
			return err
		}
		buildState.IncrementBuildProgress()
		_, err = client.DockerExec(localNodeNum,
			"pantheon --data-path=/pantheon/data public-key export --to=/pantheon/data/publicKey")
		if err != nil {
			log.Println(err)
			return err
		}

		addr, err := client.DockerExec(localNodeNum, "cat /pantheon/data/nodeAddress")
		if err != nil {
			log.Println(err)
			return err
		}

		addrs := string(addr[2:])

		mux.Lock()
		addresses[absoluteNodeNum] = addrs
		mux.Unlock()

		key, err := client.DockerExec(localNodeNum, "cat /pantheon/data/publicKey")
		if err != nil {
			log.Println(err)
			return err
		}
		buildState.IncrementBuildProgress()

		mux.Lock()
		pubKeys[absoluteNodeNum] = key[2:]
		mux.Unlock()

		privKey, err := client.DockerExec(localNodeNum, "cat /pantheon/data/key")
		if err != nil {
			log.Println(err)
			return err
		}
		mux.Lock()
		privKeys[absoluteNodeNum] = privKey[2:]
		mux.Unlock()

		_, err = client.DockerExec(localNodeNum, "bash -c 'echo \"[\\\""+addrs+"\\\"]\" >> /pantheon/data/toEncode.json'")
		if err != nil {
			log.Println(err)
			return err
		}

		_, err = client.DockerExec(localNodeNum, "mkdir /pantheon/genesis")
		if err != nil {
			log.Println(err)
			return err
		}

		// used for IBFT2 extraData
		_, err = client.DockerExec(localNodeNum,
			"pantheon rlp encode --from=/pantheon/data/toEncode.json --to=/pantheon/rlpEncodedExtraData")
		if err != nil {
			log.Println(err)
			return err
		}

		rlpEncoded, err := client.DockerExec(localNodeNum, "bash -c 'cat /pantheon/rlpEncodedExtraData'")
		if err != nil {
			log.Println(err)
			return err
		}

		mux.Lock()
		rlpEncodedData[absoluteNodeNum] = rlpEncoded
		mux.Unlock()

		//fmt.Println(rlpEncodedData)

		buildState.IncrementBuildProgress()
		return err

	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//<- extraAccChan
	extraAccs := <-extraAccChan
	addresses = append(addresses, extraAccs...)

	wg.Add(1)
	defer wg.Wait()
	go func() {
		defer wg.Done()
		/*
		   Set up a geth node, which is not part of the blockchain network,
		   to sign the transactions in place of the pantheon client. The pantheon
		   client does not support wallet management, so this acts as an easy work around.
		*/
		clients := tn.GetFlatClients()
		err := startGeth(clients[0], panconf, addresses, privKeys, buildState)
		if err != nil {
			log.Println(err)
			buildState.ReportError(err)
			return
		}
	}()

	/* Create Genesis File */
	buildState.SetBuildStage("Generating Genesis File")
	err = createGenesisfile(panconf, tn.LDD, addresses, buildState, rlpEncodedData[0])
	if err != nil {
		log.Println(err)
		return nil, err
	}

	p2pPort := 30303
	enodes := "["
	var enodeAddress string
	for i, node := range tn.Nodes {
		enodeAddress = fmt.Sprintf("enode://%s@%s:%d",
			pubKeys[i],
			node.IP,
			p2pPort)
		if i < len(pubKeys)-1 {
			enodes = enodes + "\"" + enodeAddress + "\"" + ","
		} else {
			enodes = enodes + "\"" + enodeAddress + "\""
		}
		buildState.IncrementBuildProgress()
	}

	enodes = enodes + "]"

	/* Create Static Nodes File */
	buildState.SetBuildStage("Setting Up Static Peers")
	buildState.IncrementBuildProgress()
	err = helpers.CopyBytesToAllNodes(tn, enodes, "/pantheon/data/static-nodes.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = helpers.CreateConfigs(tn, "/pantheon/config.toml",
		func(serverNum int, localNodeNum int, absoluteNodeNum int) ([]byte, error) {
			return helpers.GetBlockchainConfig("pantheon", absoluteNodeNum, "config.toml", tn.LDD)
		})

	if err != nil {
		log.Println(err)
		return nil, err
	}
	/* Copy static-nodes & genesis files to each node */
	buildState.SetBuildStage("Distributing Files")
	err = helpers.CopyToAllNodes(tn, "genesis.json", "/pantheon/genesis/genesis.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	/* Start the nodes */
	buildState.SetBuildStage("Starting Pantheon")
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, localNodeNum int, absoluteNodeNum int) error {
		err := client.DockerExecdLog(localNodeNum, fmt.Sprintf(
			`pantheon --config-file=/pantheon/config.toml --data-path=/pantheon/data --genesis-file=/pantheon/genesis/genesis.json  `+
				`--rpc-http-enabled --rpc-http-api="ADMIN,CLIQUE,DEBUG,EEA,ETH,IBFT,MINER,NET,TXPOOL,WEB3" `+
				` --p2p-port=%d --rpc-http-port=8545 --rpc-http-host="0.0.0.0" --host-whitelist=all --rpc-http-cors-origins="*"`,
			p2pPort))
		buildState.IncrementBuildProgress()
		return err
	})
	return privKeys, err
}

func createGenesisfile(panconf *panConf, details *db.DeploymentDetails, address []string, buildState *state.BuildState, ibftExtraData string) error {
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
		for i := 0; i < len(address) && i < details.Nodes; i++ {
			extraData += address[i]
		}
		extraData += "000000000000000000000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000000000000000000000000000000000000000"
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

func prepareGeth(client *ssh.Client, panconf *panConf, nodes int, buildState *state.BuildState) ([]string, error) {
	addresses := []string{}
	passwd := ""
	for i := 0; int64(i) < panconf.Accounts+int64(nodes); i++ {
		passwd += "second\n"
	}
	err := buildState.Write("passwd2", passwd)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//Set up a geth node as a service to sign transactions
	client.Run(`docker exec wb_service0 mkdir /geth/`)

	err = client.Scp("passwd2", "/home/appo/passwd2")
	buildState.Defer(func() { client.Run("rm /home/appo/passwd2") })
	if err != nil {
		log.Println(err)
		return nil, err
	}
	_, err = client.Run("docker cp /home/appo/passwd2 wb_service0:/geth/passwd")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	toCreate := panconf.Accounts
	wg := &sync.WaitGroup{}
	mux := &sync.Mutex{}
	for i := 0; i < int(toCreate); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gethResults, err := client.Run("docker exec wb_service0 geth --datadir /geth/ --password /geth/passwd account new")
			if err != nil {
				log.Println(err)
				buildState.ReportError(err)
				return
			}

			addressPattern := regexp.MustCompile(`\{[A-z|0-9]+\}`)
			addrRaw := addressPattern.FindAllString(gethResults, -1)
			if len(addrRaw) < 1 {
				buildState.ReportError(fmt.Errorf("unable to get addresses"))
				return
			}
			address := addrRaw[0]
			address = address[1 : len(address)-1]

			mux.Lock()
			addresses = append(addresses, address)
			mux.Unlock()

		}()
	}
	wg.Wait()
	return addresses, nil
}

func startGeth(client *ssh.Client, panconf *panConf, addresses []string, privKeys []string, buildState *state.BuildState) error {
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

	unlock := ""
	for i, privKey := range privKeys {
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
	}

	for i, addr := range addresses {
		if i != 0 {
			unlock += ","
		}
		unlock += "0x" + addr
	}
	_, err = client.Run(fmt.Sprintf(`docker exec -itd wb_service0 bash -ic 'geth --datadir /geth/ --rpc --rpcaddr 0.0.0.0`+
		` --rpcapi "admin,web3,db,eth,net,personal" --rpccorsdomain "0.0.0.0" --nodiscover --unlock="%s"`+
		` --password /geth/passwd --networkid %d console 2>&1 >> /output.log'`, unlock, panconf.NetworkID))
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
