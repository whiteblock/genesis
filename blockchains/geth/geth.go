package geth

import (
	"../../db"
	"../../ssh"
	"../../state"
	"../../testnet"
	"../../util"
	"../helpers"
	"encoding/json"
	"fmt"
	"github.com/Whiteblock/mustache"
	"log"
	"regexp"
	"strings"
	"sync"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

const ETH_NET_STATS_PORT = 3338

/*
Build builds out a fresh new ethereum test network using geth
*/
func Build(tn *testnet.TestNet) ([]string, error) {
	clients := tn.GetFlatClients()
	mux := sync.Mutex{}
	ethconf, err := NewConf(tn.LDD.Params)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tn.BuildState.SetBuildSteps(8 + (5 * tn.LDD.Nodes))

	tn.BuildState.IncrementBuildProgress()

	tn.BuildState.SetBuildStage("Distributing secrets")
	/**Copy over the password file**/
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, localNodeNum int, _ int) error {
		_, err := client.DockerExec(localNodeNum, "mkdir -p /geth")
		return err
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	/**Create the Password files**/
	{
		var data string
		for i := 1; i <= tn.LDD.Nodes; i++ {
			data += "second\n"
		}
		err = helpers.CopyBytesToAllNodes(tn, data, "/geth/passwd")
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	tn.BuildState.IncrementBuildProgress()

	/**Create the wallets**/
	wallets := make([]string, tn.LDD.Nodes)
	rawWallets := make([]string, tn.LDD.Nodes)
	tn.BuildState.SetBuildStage("Creating the wallets")

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, localNodeNum int, absoluteNodeNum int) error {
		gethResults, err := client.DockerExec(localNodeNum, "geth --datadir /geth/ --password /geth/passwd account new")
		if err != nil {
			log.Println(err)
			return err
		}

		addressPattern := regexp.MustCompile(`\{[A-z|0-9]+\}`)
		addresses := addressPattern.FindAllString(gethResults, -1)
		if len(addresses) < 1 {
			return fmt.Errorf("unable to get addresses")
		}
		address := addresses[0]
		address = address[1 : len(address)-1]

		//fmt.Printf("Created wallet with address: %s\n",address)
		mux.Lock()
		wallets[absoluteNodeNum] = address
		mux.Unlock()

		tn.BuildState.IncrementBuildProgress()

		res, err := client.DockerExec(localNodeNum, "bash -c 'cat /geth/keystore/*'")
		if err != nil {
			log.Println(err)
			return err
		}
		mux.Lock()
		rawWallets[absoluteNodeNum] = strings.Replace(res, "\"", "\\\"", -1)
		mux.Unlock()

		return nil
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	fmt.Printf("%v\n%v\n", wallets, rawWallets)
	tn.BuildState.IncrementBuildProgress()
	unlock := ""

	for i, wallet := range wallets {
		if i != 0 {
			unlock += ","
		}
		unlock += wallet
	}
	fmt.Printf("unlock = %s\n%+v\n\n", wallets, unlock)

	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Creating the genesis block")
	err = createGenesisfile(ethconf, tn.LDD, wallets, tn.BuildState)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Bootstrapping network")

	err = helpers.CopyToAllNodes(tn, "CustomGenesis.json", "/geth/")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, localNodeNum int, absoluteNodeNum int) error {
		for i := range rawWallets {
			if i == absoluteNodeNum {
				continue
			}
			_, err := client.DockerExec(localNodeNum,
				fmt.Sprintf("bash -c 'echo \"%s\">>/geth/keystore/account%d'", rawWallets[i], i))
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

	staticNodes := make([]string, tn.LDD.Nodes)

	tn.BuildState.SetBuildStage("Initializing geth")

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, localNodeNum int, absoluteNodeNum int) error {
		ip := tn.Nodes[absoluteNodeNum].IP
		//Load the CustomGenesis file
		_, err := client.DockerExec(localNodeNum,
			fmt.Sprintf("geth --datadir /geth/ --networkid %d init /geth/CustomGenesis.json", ethconf.NetworkId))
		if err != nil {
			log.Println(err)
			return err
		}
		fmt.Printf("---------------------  CREATING block directory for NODE-%d ---------------------\n", absoluteNodeNum)
		gethResults, err := client.DockerExec(localNodeNum,
			fmt.Sprintf("bash -c 'echo -e \"admin.nodeInfo.enode\\nexit\\n\" | "+
				"geth --rpc --datadir /geth/ --networkid %d console'", ethconf.NetworkId))
		if err != nil {
			log.Println(err)
			return err
		}
		//fmt.Printf("RAWWWWWWWWWWWW%s\n\n\n",gethResults)
		enodePattern := regexp.MustCompile(`enode:\/\/[A-z|0-9]+@(\[\:\:\]|([0-9]|\.)+)\:[0-9]+`)
		enode := enodePattern.FindAllString(gethResults, 1)[0]
		//fmt.Printf("ENODE fetched is: %s\n",enode)
		enodeAddressPattern := regexp.MustCompile(`\[\:\:\]|([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})`)

		enode = enodeAddressPattern.ReplaceAllString(enode, ip)

		mux.Lock()
		staticNodes[absoluteNodeNum] = enode
		mux.Unlock()

		tn.BuildState.IncrementBuildProgress()
		return nil
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	out, err := json.Marshal(staticNodes)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Starting geth")
	//Copy static-nodes to every server
	err = helpers.CopyBytesToAllNodes(tn, string(out), "/geth/static-nodes.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, localNodeNum int, absoluteNodeNum int) error {
		ip := tn.Nodes[absoluteNodeNum].IP
		tn.BuildState.IncrementBuildProgress()

		gethCmd := fmt.Sprintf(
			`geth --datadir /geth/ --maxpeers %d --networkid %d --rpc --nodiscover --rpcaddr %s`+
				` --rpcapi "web3,db,eth,net,personal,miner,txpool" --rpccorsdomain "0.0.0.0" --mine --unlock="%s"`+
				` --password /geth/passwd --etherbase %s console  2>&1 | tee %s`,
			ethconf.MaxPeers,
			ethconf.NetworkId,
			ip,
			unlock,
			wallets[absoluteNodeNum],
			conf.DockerOutputFile)

		_, err := client.DockerExecdit(localNodeNum, fmt.Sprintf("bash -ic '%s'", gethCmd))
		if err != nil {
			log.Println(err)
			return err
		}

		tn.BuildState.IncrementBuildProgress()
		return nil
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	tn.BuildState.IncrementBuildProgress()

	err = setupEthNetStats(clients[0])
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, localNodeNum int, absoluteNodeNum int) error {
		ip := tn.Nodes[absoluteNodeNum].IP
		absName := fmt.Sprintf("%s%d", conf.NodePrefix, absoluteNodeNum)
		sedCmd := fmt.Sprintf(`sed -i -r 's/"INSTANCE_NAME"(\s)*:(\s)*"(\S)*"/"INSTANCE_NAME"\t: "%s"/g' /eth-net-intelligence-api/app.json`, absName)
		sedCmd2 := fmt.Sprintf(`sed -i -r 's/"WS_SERVER"(\s)*:(\s)*"(\S)*"/"WS_SERVER"\t: "http:\/\/%s:%d"/g' /eth-net-intelligence-api/app.json`,
			util.GetGateway(server.SubnetID, absoluteNodeNum), ETH_NET_STATS_PORT)
		sedCmd3 := fmt.Sprintf(`sed -i -r 's/"RPC_HOST"(\s)*:(\s)*"(\S)*"/"RPC_HOST"\t: "%s"/g' /eth-net-intelligence-api/app.json`, ip)

		//sedCmd3 := fmt.Sprintf("docker exec -it %s sed -i 's/\"WS_SECRET\"(\\s)*:(\\s)*\"[A-Z|a-z|0-9| ]*\"/\"WS_SECRET\"\\t: \"second\"/g' /eth-net-intelligence-api/app.json",container)
		_, err := client.DockerMultiExec(localNodeNum, []string{
			sedCmd,
			sedCmd2,
			sedCmd3})

		if err != nil {
			log.Println(err)
			return err
		}
		_, err = client.DockerExecd(localNodeNum, "bash -c 'cd /eth-net-intelligence-api && pm2 start app.json'")
		if err != nil {
			log.Println(err)
			return err
		}
		tn.BuildState.IncrementBuildProgress()
		return nil
	})
	return nil, err
}

/***************************************************************************************************************************/

func Add(tn *testnet.TestNet) ([]string, error) {
	return nil, nil
}

func MakeFakeAccounts(accs int) []string {
	out := make([]string, accs)
	for i := 1; i <= accs; i++ {
		acc := fmt.Sprintf("%X", i)
		for j := len(acc); j < 40; j++ {
			acc = "0" + acc
		}
		acc = "0x" + acc
		out[i-1] = acc
	}
	return out
}

/**
 * Create the custom genesis file for Ethereum
 * @param  *EthConf ethconf     The chain configuration
 * @param  []string wallets     The wallets to be allocated a balance
 */

func createGenesisfile(ethconf *EthConf, details *db.DeploymentDetails, wallets []string, buildState *state.BuildState) error {

	genesis := map[string]interface{}{
		"chainId":        ethconf.NetworkId,
		"homesteadBlock": ethconf.HomesteadBlock,
		"eip155Block":    ethconf.Eip155Block,
		"eip158Block":    ethconf.Eip158Block,
		"difficulty":     fmt.Sprintf("0x0%X", ethconf.Difficulty),
		"gasLimit":       fmt.Sprintf("0x0%X", ethconf.GasLimit),
	}
	alloc := map[string]map[string]string{}
	for _, wallet := range wallets {
		alloc[wallet] = map[string]string{
			"balance": ethconf.InitBalance,
		}
	}

	accs := MakeFakeAccounts(int(ethconf.ExtraAccounts))
	log.Println("Finished making fake accounts")

	for _, wallet := range accs {
		alloc[wallet] = map[string]string{
			"balance": ethconf.InitBalance,
		}
	}
	genesis["alloc"] = alloc
	dat, err := helpers.GetBlockchainConfig("geth", 0, "genesis.json", details)
	if err != nil {
		log.Println(err)
		return err
	}

	data, err := mustache.Render(string(dat), util.ConvertToStringMap(genesis))
	if err != nil {
		log.Println(err)
		return err
	}
	return buildState.Write("CustomGenesis.json", data)

}

/**
 * Setup Eth Net Stats on a server
 * @param  string    ip     The servers config
 */
func setupEthNetStats(client *ssh.Client) error {
	_, err := client.Run(fmt.Sprintf(
		"docker exec -d wb_service0 bash -c 'cd /eth-netstats && WS_SECRET=second PORT=%d npm start'", ETH_NET_STATS_PORT))
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
