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

const blockchain = "pantheon"

func init() {
	conf = util.GetConfig()
	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, GetDefaults)
	registrar.RegisterParams(blockchain, GetParams)
	registrar.RegisterBlockchainSideCars(blockchain, []string{"geth"})

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

	extraAccChan := make(chan []*ethereum.Account)

	go func() {
		accs, err := prepareGeth(tn.GetFlatClients()[0], panconf, tn)
		if err != nil {
			tn.BuildState.ReportError(err)
			extraAccChan <- nil
			return
		}
		extraAccChan <- accs
	}()

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
			tn.BuildState.ReportError(err)
			return
		}
	})

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
		return client.DockerExecdLog(node, fmt.Sprintf(
			`pantheon --config-file=/pantheon/config.toml --data-path=/pantheon/data --genesis-file=/pantheon/genesis/genesis.json  `+
				`--rpc-http-enabled --rpc-http-api="ADMIN,CLIQUE,DEBUG,EEA,ETH,IBFT,MINER,NET,TXPOOL,WEB3" `+
				` --p2p-port=%d --rpc-http-port=8545 --rpc-http-host="0.0.0.0" --host-whitelist=all --rpc-http-cors-origins="*"`,
			p2pPort))
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

func prepareGeth(client *ssh.Client, panconf *panConf, tn *testnet.TestNet) ([]*ethereum.Account, error) {
	passwd := ""
	for i := 0; int64(i) < panconf.Accounts+int64(tn.LDD.Nodes); i++ {
		passwd += "second\n"
	}
	err := tn.BuildState.Write("passwd2", passwd)
	if err != nil {
		return nil, util.LogError(err)
	}
	//Set up a geth node as a service to sign transactions
	client.Run(`docker exec wb_service0 mkdir /geth/`)

	err = client.Scp("passwd2", "/home/appo/passwd2")
	tn.BuildState.Defer(func() { client.Run("rm /home/appo/passwd2") })
	if err != nil {
		return nil, util.LogError(err)
	}
	_, err = client.Run("docker cp /home/appo/passwd2 wb_service0:/geth/passwd")
	if err != nil {
		return nil, util.LogError(err)
	}

	toCreate := panconf.Accounts

	accounts, err := ethereum.GenerateAccounts(int(toCreate))
	if err != nil {
		return nil, util.LogError(err)
	}

	return accounts, nil
}

func startGeth(client *ssh.Client, panconf *panConf, accounts []*ethereum.Account, buildState *state.BuildState) error {
	serviceIps, err := util.GetServiceIps(GetServices())
	if err != nil {
		return util.LogError(err)
	}
	err = buildState.SetExt("signer_ip", serviceIps["geth"])
	if err != nil {
		return util.LogError(err)
	}
	err = buildState.SetExt("accounts", ethereum.ExtractAddresses(accounts))
	if err != nil {
		return util.LogError(err)
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
		return util.LogError(err)
	}
	return nil
}

// Add handles adding a node to the pantheon testnet
// TODO
func add(tn *testnet.TestNet) error {
	return nil
}
