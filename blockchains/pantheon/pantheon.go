package pantheon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	db "../../db"
	state "../../state"
	util "../../util"
	"github.com/Whiteblock/mustache"
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
func Build(details db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
	buildState *state.BuildState) ([]string, error) {
	//var mutex = &sync.Mutex{}
	var sem = semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()
	mux := sync.Mutex{}
	ethconf, err := NewConf(details.Params)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	buildState.SetBuildSteps(8 + (5 * details.Nodes))

	buildState.IncrementBuildProgress()

	/**Create the Password files**/
	{
		var data string
		for i := 1; i <= details.Nodes; i++ {
			data += "second\n"
		}
		err = util.Write("./passwd", data)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	defer util.Rm("./passwd")
	buildState.SetBuildStage("Distributing secrets")
	/**Copy over the password file**/
	for i, server := range servers {
		err = clients[i].Scp("./passwd", "/home/appo/passwd")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer clients[i].Run("rm /home/appo/passwd")

		for j, _ := range server.Ips {
			res, err := clients[i].DockerExec(j, "mkdir -p /pantheon")
			if err != nil {
				log.Println(res)
				log.Println(err)
				return nil, err
			}

			err = clients[i].DockerCp(j, "/home/appo/passwd", "/pantheon/")
			if err != nil {
				log.Println(err)
				return nil, err
			}
		}
		buildState.IncrementBuildProgress()
	}

	/**Create the wallets**/
	wallets := make([]string, details.Nodes)
	rawWallets := make([]string, details.Nodes)
	buildState.SetBuildStage("Creating the wallets")
	{
		node := 0
		for i, server := range servers {
			for j, _ := range server.Ips {
				sem.Acquire(ctx, 1)
				go func(index int, node int) {
					defer sem.Release(1)
					pantheonResults, err := clients[i].DockerExec(node, "pantheon --data-path /pantheon/ --password /pantheon/passwd account new")
					if err != nil {
						buildState.ReportError(err)
						log.Println(pantheonResults)
						log.Println(err)
						return
					}

					addressPattern := regexp.MustCompile(`\{[A-z|0-9]+\}`)
					addresses := addressPattern.FindAllString(pantheonResults, -1)
					if len(addresses) < 1 {
						buildState.ReportError(errors.New("Unable to get addresses"))
					}
					address := addresses[0]
					address = address[1 : len(address)-1]

					mux.Lock()
					wallets[index] = address
					mux.Unlock()

					buildState.IncrementBuildProgress()

					res, err := clients[i].DockerExec(node, "bash -c 'cat /pantheon/keystore/*'")
					if err != nil {
						buildState.ReportError(err)
						log.Println(res)
						log.Println(err)
						return
					}
					mux.Lock()
					rawWallets[index] = strings.Replace(res, "\"", "\\\"", -1)
					mux.Unlock()
				}(node, j)
				node++
			}
		}

		err = sem.Acquire(ctx, conf.ThreadLimit)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		sem.Release(conf.ThreadLimit)
	}
	fmt.Printf("%v\n%v\n", wallets, rawWallets)
	buildState.IncrementBuildProgress()
	unlock := ""

	for i, wallet := range wallets {
		if i != 0 {
			unlock += ","
		}
		unlock += wallet
	}
	fmt.Printf("unlock = %s\n%+v\n\n", wallets, unlock)

	buildState.IncrementBuildProgress()
	buildState.SetBuildStage("Creating the genesis block")
	err = createGenesisfile(ethconf, details, wallets)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer util.Rm("./genesis.json")

	buildState.IncrementBuildProgress()
	buildState.SetBuildStage("Bootstrapping network")
	node := 0
	for i, server := range servers {
		err = clients[i].Scp("./genesis.json", "/home/appo/genesis.json")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer clients[i].Run("rm /home/appo/genesis.json")

		for j, _ := range server.Ips {
			sem.Acquire(ctx, 1)
			go func(i int, j int, node int) {
				defer sem.Release(1)
				err := clients[i].DockerCp(j, "/home/appo/genesis.json", "/pantheon/")
				if err != nil {
					log.Println(err)
					buildState.ReportError(err)
					return
				}
				for k, rawWallet := range rawWallets {
					if k == node {
						continue
					}
					_, err = clients[i].DockerExec(j, fmt.Sprintf("bash -c 'echo \"%s\">>/pantheon/keystore/account%d'", rawWallet, k))
					if err != nil {
						log.Println(err)
						buildState.ReportError(err)
						return
					}
				}
			}(i, j, node)
			node++
		}
	}

	err = sem.Acquire(ctx, conf.ThreadLimit)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	sem.Release(conf.ThreadLimit)

	if !buildState.ErrorFree() {
		return nil, buildState.GetError()
	}

	staticNodes := make([]string, details.Nodes)
	node = 0
	for i, server := range servers {
		for j, ip := range server.Ips {
			sem.Acquire(ctx, 1)
			go func(i int, j int, node int, ip string) {
				defer sem.Release(1)
				res, err := clients[i].DockerExec(j,
					fmt.Sprintf("pantheon --data-path /pantheon/ --networkid %d --genesis /pantheon/genesis.json", ethconf.NetworkId))
				if err != nil {
					log.Println(res)
					log.Println(err)
					buildState.ReportError(err)
					return
				}
				fmt.Printf("---------------------  CREATING block directory for NODE-%d ---------------------\n", node)
				pantheonResults, err := clients[i].DockerExec(j,
					fmt.Sprintf("bash -c 'echo -e \"admin.nodeInfo.enode\\nexit\\n\" |  pantheon --rpc --data-path /pantheon/ --networkid %d console'", ethconf.NetworkId))
				if err != nil {
					log.Println(err)
					buildState.ReportError(err)
					return
				}
				enodePattern := regexp.MustCompile(`enode:\/\/[A-z|0-9]+@(\[\:\:\]|([0-9]|\.)+)\:[0-9]+`)
				enode := enodePattern.FindAllString(pantheonResults, 1)[0]
				enodeAddressPattern := regexp.MustCompile(`\[\:\:\]|([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})`)
				enode = enodeAddressPattern.ReplaceAllString(enode, ip)
				mux.Lock()
				staticNodes[node] = enode
				mux.Unlock()

				buildState.IncrementBuildProgress()
			}(i, j, node, ip)
			node++

		}
	}

	err = sem.Acquire(ctx, conf.ThreadLimit)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	sem.Release(conf.ThreadLimit)

	if !buildState.ErrorFree() {
		return nil, buildState.GetError()
	}

	out, err := json.Marshal(staticNodes)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer util.Rm("static-nodes.json")
	err = util.Write("static-nodes.json", string(out))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	buildState.IncrementBuildProgress()
	buildState.SetBuildStage("Starting pantheon")
	node = 0
	for i, server := range servers {
		err = clients[i].Scp("./static-nodes.json", "/home/appo/static-nodes.json")
		if err != nil {
			log.Println(err)
			return nil, err
		}

		for j, ip := range server.Ips {
			sem.Acquire(ctx, 1)
			fmt.Printf("-----------------------------  Starting NODE-%d  -----------------------------\n", node)

			go func(networkId int64, node int, server string, num int, unlock string, nodeIP string, i int) {
				defer sem.Release(1)

				buildState.IncrementBuildProgress()

				pantheonCmd := fmt.Sprintf(
					`pantheon --data-path /pantheon/ --maxpeers %d --networkid %d --rpc --rpcaddr %s`+
						` --rpcapi "web3,db,eth,net,personal,miner,txpool" --rpccorsdomain "0.0.0.0" --miner-enabled --unlock="%s"`+
						` --password /pantheon/passwd --miner-etherbase %s console  2>&1 | tee output.log`,
					ethconf.MaxPeers,
					networkId,
					nodeIP,
					unlock,
					wallets[node])

				err = clients[i].DockerCp(num, "/home/appo/static-nodes.json", "/pantheon/")
				if err != nil {
					log.Println(err)
					buildState.ReportError(err)
					return
				}
				clients[i].DockerExecd(num, "tmux new -s whiteblock -d")
				clients[i].DockerExecd(num, fmt.Sprintf("tmux send-keys -t whiteblock '%s' C-m", pantheonCmd))

				if err != nil {
					log.Println(err)
					buildState.ReportError(err)
					return
				}

				buildState.IncrementBuildProgress()
			}(ethconf.NetworkId, node, server.Addr, j, unlock, ip, i)
			node++
		}
	}
	err = sem.Acquire(ctx, conf.ThreadLimit)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	sem.Release(conf.ThreadLimit)

	buildState.IncrementBuildProgress()
	if !buildState.ErrorFree() {
		return nil, buildState.GetError()
	}

	err = sem.Acquire(ctx, conf.ThreadLimit)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	sem.Release(conf.ThreadLimit)
	return nil, nil

}

/***************************************************************************************************************************/

func Add(details db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
	newNodes map[int][]string, buildState *state.BuildState) ([]string, error) {
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

func createGenesisfile(ethconf *EthConf, details db.DeploymentDetails, wallets []string) error {

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
	dat, err := util.GetBlockchainConfig("pantheon", "genesis.json", details.Files)
	if err != nil {
		log.Println(err)
		return err
	}

	data, err := mustache.Render(string(dat), util.ConvertToStringMap(genesis))
	if err != nil {
		log.Println(err)
		return err
	}
	return util.Write("genesis.json", data)

}
