//Package pantheon handles pantheon specific functionality
package pantheon

import (
	"../../db"
	"../../ssh"
	"../../state"
	"../../testnet"
	"../../util"
	"../ethereum"
	"../helpers"
	"../registrar"
	"fmt"
	"github.com/Whiteblock/mustache"
	"log"
	"sync"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
	blockchain := "pantheon"
	registrar.RegisterBuild(blockchain, Build)
	registrar.RegisterAddNodes(blockchain, Add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, GetDefaults)
	registrar.RegisterParams(blockchain, GetParams)
}

// Build builds out a fresh new ethereum test network using pantheon
func Build(tn *testnet.TestNet) ([]string, error) {
	mux := sync.Mutex{}

	panconf, err := newConf(tn.LDD.Params)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tn.BuildState.SetBuildSteps(6*tn.LDD.Nodes + 2)
	tn.BuildState.IncrementBuildProgress()

	rlpEncodedData := make([]string, tn.LDD.Nodes)
	accounts := make([]*ethereum.Account, tn.LDD.Nodes)

	extraAccChan := make(chan []*ethereum.Account)

	go func() {
		accs, err := prepareGeth(tn.GetFlatClients()[0], panconf, tn)
		if err != nil {
			log.Println(err)
			tn.BuildState.ReportError(err)
			extraAccChan <- nil
			return
		}
		extraAccChan <- accs
	}()

	tn.BuildState.SetBuildStage("Setting Up Accounts")

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, localNodeNum int, absoluteNodeNum int) error {

		tn.BuildState.IncrementBuildProgress()
		_, err = client.DockerExec(localNodeNum,
			"pantheon --data-path=/pantheon/data public-key export --to=/pantheon/data/publicKey")
		if err != nil {
			log.Println(err)
			return err
		}

		privKey, err := client.DockerRead(localNodeNum, "/pantheon/data/key", -1)
		if err != nil {
			log.Println(err)
			return err
		}
		acc, err := ethereum.CreateAccountFromHex(privKey[2:])
		if err != nil {
			log.Println(err)
			return err
		}

		mux.Lock()
		accounts[absoluteNodeNum] = acc
		mux.Unlock()
		tn.BuildState.IncrementBuildProgress()
		addr := acc.HexAddress()

		_, err = client.DockerExec(localNodeNum, "bash -c 'echo \"[\\\""+addr[2:]+"\\\"]\" >> /pantheon/data/toEncode.json'")
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

		rlpEncoded, err := client.DockerRead(localNodeNum, "/pantheon/rlpEncodedExtraData", -1)
		if err != nil {
			log.Println(err)
			return err
		}

		mux.Lock()
		rlpEncodedData[absoluteNodeNum] = rlpEncoded
		mux.Unlock()

		tn.BuildState.IncrementBuildProgress()
		return nil

	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//<- extraAccChan
	extraAccs := <-extraAccChan
	accounts = append(accounts, extraAccs...)

	tn.BuildState.Async(func() {

		/*
		   Set up a geth node, which is not part of the blockchain network,
		   to sign the transactions in place of the pantheon client. The pantheon
		   client does not support wallet management, so this acts as an easy work around.
		*/
		clients := tn.GetFlatClients()
		err := startGeth(clients[0], panconf, accounts, tn.BuildState)
		if err != nil {
			log.Println(err)
			tn.BuildState.ReportError(err)
			return
		}
	})

	/* Create Genesis File */
	tn.BuildState.SetBuildStage("Generating Genesis File")
	err = createGenesisfile(panconf, tn, accounts, rlpEncodedData[0])
	if err != nil {
		log.Println(err)
		return nil, err
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
		return nil, err
	}

	err = helpers.CreateConfigs(tn, "/pantheon/config.toml", func(_ int, _ int, absoluteNodeNum int) ([]byte, error) {
		return helpers.GetBlockchainConfig("pantheon", absoluteNodeNum, "config.toml", tn.LDD)
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	/* Copy static-nodes & genesis files to each node */
	tn.BuildState.SetBuildStage("Distributing Files")
	err = helpers.CopyToAllNodes(tn, "genesis.json", "/pantheon/genesis/genesis.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	/* Start the nodes */
	tn.BuildState.SetBuildStage("Starting Pantheon")
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, localNodeNum int, _ int) error {
		err := client.DockerExecdLog(localNodeNum, fmt.Sprintf(
			`pantheon --config-file=/pantheon/config.toml --data-path=/pantheon/data --genesis-file=/pantheon/genesis/genesis.json  `+
				`--rpc-http-enabled --rpc-http-api="ADMIN,CLIQUE,DEBUG,EEA,ETH,IBFT,MINER,NET,TXPOOL,WEB3" `+
				` --p2p-port=%d --rpc-http-port=8545 --rpc-http-host="0.0.0.0" --host-whitelist=all --rpc-http-cors-origins="*"`,
			p2pPort))
		tn.BuildState.IncrementBuildProgress()
		return err
	})
	for _, account := range accounts {
		tn.BuildState.SetExt(account.HexAddress(), map[string]string{
			"privateKey": account.HexPrivateKey(),
			"publicKey":  account.HexPublicKey(),
		})
	}
	return nil, err
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
		log.Println(err)
		return err
	}

	data, err := mustache.Render(string(dat), util.ConvertToStringMap(genesis))
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Println("Writing Genesis File Locally")
	return tn.BuildState.Write("genesis.json", data)

}

func createStaticNodesFile(list string, buildState *state.BuildState) error {
	return buildState.Write("static-nodes.json", list)
}

func prepareGeth(client *ssh.Client, panconf *panConf, tn *testnet.TestNet) ([]*ethereum.Account, error) {
	passwd := ""
	for i := 0; int64(i) < panconf.Accounts+int64(tn.LDD.Nodes); i++ {
		passwd += "second\n"
	}
	err := tn.BuildState.Write("passwd2", passwd)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//Set up a geth node as a service to sign transactions
	client.Run(`docker exec wb_service0 mkdir /geth/`)

	err = client.Scp("passwd2", "/home/appo/passwd2")
	tn.BuildState.Defer(func() { client.Run("rm /home/appo/passwd2") })
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

	accounts, err := ethereum.GenerateAccounts(int(toCreate))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return accounts, nil
}

func startGeth(client *ssh.Client, panconf *panConf, accounts []*ethereum.Account, buildState *state.BuildState) error {
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
	err = buildState.SetExt("accounts", ethereum.ExtractAddresses(accounts))
	if err != nil {
		log.Println(err)
		return err
	}

	wg := &sync.WaitGroup{}

	for i, account := range accounts {
		wg.Add(1)
		go func(account *ethereum.Account, i int) {
			defer wg.Done()

			_, err := client.Run(fmt.Sprintf("docker exec wb_service0 bash -c 'echo \"%s\" >> /geth/pk%d'", account.HexPrivateKey(), i))
			if err != nil {
				log.Println(err)
				return
			}
			_, err = client.Run(
				fmt.Sprintf("docker exec wb_service0 geth "+
					"--datadir /geth/ --password /geth/passwd account import --password /geth/passwd /geth/pk%d", i))
			if err != nil {
				log.Println(err)
				return
			}

		}(account, i)
	}
	wg.Wait()

	unlock := ""
	for i, account := range accounts {
		if i != 0 {
			unlock += ","
		}
		unlock += account.HexAddress()
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

// Add handles adding a node to the pantheon testnet
// TODO
func Add(tn *testnet.TestNet) ([]string, error) {
	return nil, nil
}
