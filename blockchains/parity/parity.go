package parity

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	db "../../db"
	state "../../state"
	util "../../util"
	"golang.org/x/sync/semaphore"
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
func Build(data map[string]interface{}, nodes int, servers []db.Server, clients []*util.SshClient,
		   buildState *state.BuildState) ([]string, error) {
	//var mutex = &sync.Mutex{}
	var sem = semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()
	ethconf, err := NewConf(data)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = util.Rm("tmp")
	if err != nil {
		log.Println(err)
	}

	state.SetBuildSteps(8 + (5 * nodes))
	defer func() {
		fmt.Printf("Cleaning up...")
		util.Rm("tmp")
		fmt.Printf("done\n")
	}()

	for i := 1; i <= nodes; i++ {
		err = util.Mkdir(fmt.Sprintf("./tmp/node%d", i))
		if err != nil {
			log.Println(err)
			return nil, err
		}
		//fmt.Printf("---------------------  CREATING pre-allocated accounts for NODE-%d  ---------------------\n",i)

	}
	state.IncrementBuildProgress()

	/**Create the Password files**/
	{
		var data string
		for i := 1; i <= nodes; i++ {
			data += "second\n"
		}

		for i := 1; i <= nodes; i++ {
			err = util.Write(fmt.Sprintf("tmp/node%d/passwd.file", i), data)
			if err != nil {
				log.Println(err)
				return nil, err
			}
		}
	}
	state.IncrementBuildProgress()

	/**Create the wallets**/
	wallets := []string{}
	state.SetBuildStage("Creating the wallets")
	for i := 1; i <= nodes; i++ {

		node := i
		//sem.Acquire(ctx,1)
		parityResults, err := util.BashExec(
			fmt.Sprintf("parity --base-path=tmp/node%d/ --password=tmp/node%d/passwd.file account new",
				node, node))
		if err != nil {
			log.Println(err)
			return nil, err
		}
		//fmt.Printf("RAW:%s\n",parityResults)
		addressPattern := regexp.MustCompile(`\{[A-z|0-9]+\}`)
		addresses := addressPattern.FindAllString(parityResults, -1)
		if len(addresses) < 1 {
			return nil, errors.New("Unable to get addresses")
		}
		address := addresses[0]
		address = address[1 : len(address)-1]
		//sem.Release(1)
		//fmt.Printf("Created wallet with address: %s\n",address)
		//mutex.Lock()
		wallets = append(wallets, address)
		//mutex.Unlock()
		state.IncrementBuildProgress()

	}
	state.IncrementBuildProgress()
	unlock := ""

	for i, wallet := range wallets {
		if i != 0 {
			unlock += ","
		}
		unlock += wallet
	}
	fmt.Printf("unlock = %s\n%+v\n\n", wallets, unlock)

	state.IncrementBuildProgress()
	state.SetBuildStage("Creating the genesis block")
	err = createGenesisfile(ethconf, wallets)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	state.IncrementBuildProgress()
	state.SetBuildStage("Bootstrapping network")
	err = initNodeDirectories(nodes, ethconf.NetworkId, servers)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	state.IncrementBuildProgress()
	err = util.Mkdir("tmp/keystore")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	state.SetBuildStage("Distributing keys")
	err = distributeUTCKeystore(nodes)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	state.IncrementBuildProgress()
	state.SetBuildStage("Starting Parity")
	node := 0
	for i, server := range servers {
		clients[i].Scp("tmp/config.toml", "/home/appo/config.toml")
		defer clients[i].Run("rm -f /home/appo/config.toml")
		for j, ip := range server.Ips {
			sem.Acquire(ctx, 1)
			fmt.Printf("-----------------------------  Starting NODE-%d  -----------------------------\n", node)

			go func(networkId int64, node int, server string, num int, unlock string, nodeIP string, i int) {
				defer sem.Release(1)
				name := fmt.Sprintf("whiteblock-node%d", num)
				_, err := clients[i].Run(fmt.Sprintf("rm -rf tmp/node%d", node))
				if err != nil {
					log.Println(err)
					state.ReportError(err)
					return
				}
				err = clients[i].Scpr(fmt.Sprintf("tmp/node%d", node))
				if err != nil {
					log.Println(err)
					state.ReportError(err)
					return
				}
				state.IncrementBuildProgress()
				parityCmd := fmt.Sprintf(
					`parity --base-path=/whiteblock/node%d --network-id=%d`+
						`--jsonrpc-apis="all" --jsonrpc-cors="all" --unlock="%s"`+
						` --password /whiteblock/node%d/passwd.file --author=%s
                        --reserved-peers`,
					node,
					networkId,
					unlock,
					node,
					wallets[node-1])

				clients[i].Run(fmt.Sprintf("docker cp /home/appo/CustomGenesis.json whiteblock-node%d:/", node))

				clients[i].Run(fmt.Sprintf("docker exec %s mkdir -p /whiteblock/node%d/", name, node))
				clients[i].Run(fmt.Sprintf("docker cp ~/tmp/node%d %s:/whiteblock", node, name))
				clients[i].Run(fmt.Sprintf("docker exec -d %s tmux new -s whiteblock -d", name))
				clients[i].Run(fmt.Sprintf("docker exec -d %s tmux send-keys -t whiteblock '%s' C-m", name, parityCmd))

				if err != nil {
					log.Println(err)
					state.ReportError(err)
					return
				}

				state.IncrementBuildProgress()
			}(ethconf.NetworkId, node+1, server.Addr, j, unlock, ip, i)
			node++
		}
	}
	err = sem.Acquire(ctx, conf.ThreadLimit)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	state.IncrementBuildProgress()
	sem.Release(conf.ThreadLimit)
	if !state.ErrorFree() {
		return nil, state.GetError()
	}
	return nil, nil
}

/***************************************************************************************************************************/

func Add(data map[string]interface{},nodes int,servers []db.Server,clients []*util.SshClient,
         newNodes map[int][]string,buildState *state.BuildState) ([]string,error) {
    return nil,nil
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

func createGenesisfile(ethconf *EthConf, wallets []string) error {
	file, err := os.Create("tmp/CustomGenesis.json")
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	genesis := fmt.Sprintf(
		`{
    "config": {
        "chainId": %d,
        "homesteadBlock": %d,
        "eip155Block": %d,
        "eip158Block": %d
    },
    "difficulty": "0x0%X",
    "gasLimit": "0x0%X",
    "alloc": {`,
		ethconf.ChainId,
		ethconf.HomesteadBlock,
		ethconf.Eip155Block,
		ethconf.Eip158Block,
		ethconf.Difficulty,
		ethconf.GasLimit)

	_, err = file.Write([]byte(genesis))
	if err != nil {
		log.Println(err)
		return err
	}

	//Fund the accounts
	_, err = file.Write([]byte("\n"))
	if err != nil {
		log.Println(err)
		return err
	}
	for i, wallet := range wallets {
		_, err = file.Write([]byte(fmt.Sprintf("\t\t\"%s\":{\"balance\":\"%s\"}", wallet, ethconf.InitBalance)))
		if err != nil {
			log.Println(err)
			return err
		}
		if len(wallets)-1 != i {
			_, err = file.Write([]byte(","))
			if err != nil {
				log.Println(err)
				return err
			}
		}
		_, err = file.Write([]byte("\n"))
		if err != nil {
			log.Println(err)
			return err
		}
	}
	if ethconf.ExtraAccounts > 0 {
		_, err = file.Write([]byte(","))
		if err != nil {
			log.Println(err)
			return err
		}
	}
	accs := MakeFakeAccounts(int(ethconf.ExtraAccounts))
	log.Println("Finished making fake accounts")
	lenAccs := len(accs)
	for i, wallet := range accs {
		_, err = file.Write([]byte(fmt.Sprintf("\t\t\"%s\":{\"balance\":\"%s\"}", wallet, ethconf.InitBalance)))
		if err != nil {
			log.Println(err)
			return err
		}

		if lenAccs-1 != i {
			_, err = file.Write([]byte(",\n"))
			if err != nil {
				log.Println(err)
				return err
			}
		} else {
			_, err = file.Write([]byte("\n"))
			if err != nil {
				log.Println(err)
				return err
			}
		}

	}

	_, err = file.Write([]byte("\n\t}\n}"))
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

/**
 * Creates the datadir for a node and returns the enode address
 * @param  int      node        The nodes number
 * @param  int64    networkId   The test net network id
 * @param  string   ip          The node's IP address
 * @return string               The node's enode address
 */
func initNode(node int, networkId int64, ip string) (string, error) {
	fmt.Printf("---------------------  CREATING block directory for NODE-%d ---------------------\n", node)
	parityResults, err := util.BashExec(fmt.Sprintf("echo -e \"admin.nodeInfo.enode\\nexit\\n\" |  geth --rpc --datadir tmp/node%d/ --networkid %d console", node, networkId))
	if err != nil {
		log.Println(err)
		return "", nil
	}
	enodePattern := regexp.MustCompile(`enode:\/\/[A-z|0-9]+@(\[\:\:\]|([0-9]|\.)+)\:[0-9]+`)
	enode := enodePattern.FindAllString(parityResults, 1)[0]
	fmt.Printf("ENODE fetched is: %s\n", enode)
	enodeAddressPattern := regexp.MustCompile(`\[\:\:\]|([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})`)
	enode = enodeAddressPattern.ReplaceAllString(enode, ip)

	err = util.Write(fmt.Sprintf("./tmp/node%d/enode", node), fmt.Sprintf("%s\n", enode))
	return enode, err
}

/**
 * Initialize the chain from the configuration file
 * @param  int      nodes       The number of nodes
 * @param  int64    networkId   The test net network id
 * @param  []Server servers     The list of servers
 */
func initNodeDirectories(nodes int, networkId int64, servers []db.Server) error {
	static_nodes := []string{}
	node := 1
	for _, server := range servers {
		for _, ip := range server.Ips {
			res, err := util.BashExec(
				fmt.Sprintf("parity --config tmp/node%d", node))
			if err != nil {
				log.Println(res)
				log.Println(err)
				return err
			}
			static_node, err := initNode(node, networkId, ip)
			if err != nil {
				log.Println(err)
				return err
			}
			static_nodes = append(static_nodes, static_node)
			node++
		}
	}

	snodes := strings.Join(static_nodes, ",")

	for i := 1; i <= nodes; i++ {
		err := util.Write(fmt.Sprintf("tmp/node%d/peers.txt", i), snodes)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

/**
 * Distribute the UTC keystore files amongst the nodes
 * @param  int  nodes   The number of nodes
 */
func distributeUTCKeystore(nodes int) error {
	//Copy all UTC keystore files to every Node directory
	for i := 1; i <= nodes; i++ {
		err := util.Cpr(fmt.Sprintf("tmp/node%d/keystore/", i), "tmp/")
		if err != nil {
			log.Println(err)
			return err
		}
	}
	for i := 1; i <= nodes; i++ {
		err := util.Cpr("tmp/keystore/", fmt.Sprintf("tmp/node%d/", i))
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}
