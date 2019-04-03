package pantheon

import (
	"fmt"
	"log"
	// "sync"
	// "context"
	//"regexp"
	//"strings"
	// "io/ioutil"
	db "../../db"
	util "../../util"
	state "../../state"
	// "golang.org/x/sync/semaphore"
	"github.com/Whiteblock/mustache"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

func Build(details db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
	buildState *state.BuildState) ([]string, error) {
	// var mutex = &sync.Mutex{}
	// var sem = semaphore.NewWeighted(conf.ThreadLimit)
	// ctx := context.TODO()
	// mux := sync.Mutex{}
	panconf, err := NewConf(details.Params)
	if err != nil {
	     log.Println(err)
	     return nil, err
	}

	buildState.SetBuildSteps(3 * details.Nodes)

	buildState.IncrementBuildProgress()

	addresses := []string{}
	pubKeys := []string{}

	buildState.SetBuildStage("Setting Up Accounts")
	for i, server := range servers {
		for localId, _ := range server.Ips {
				_, err := clients[i].DockerExec(localId, "pantheon --data-path=/pantheon/data public-key export-address --to=/pantheon/data/nodeAddress")
				if err != nil {
						log.Println(err)
						return nil, err
				}

				_, err = clients[i].DockerExec(localId, "pantheon --data-path=/pantheon/data public-key export --to=/pantheon/data/publicKey")
				if err != nil {
						log.Println(err)
						return nil, err
				}

				_, err = clients[i].DockerExecd(localId, "touch /pantheon/data/toEncode.json")
				if err != nil {
						log.Println(err)
						return nil, err
				}

				addr, err := clients[i].DockerExec(localId, "cat /pantheon/data/nodeAddress")
				if err != nil {
						log.Println(err)
						return nil, err
				}

				addrs := string(addr[2:])
				addresses = append(addresses, addrs)

				key, err := clients[i].DockerExec(localId, "cat /pantheon/data/publicKey")
				if err != nil {
						log.Println(err)
						return nil, err
				}

				keys := string(key[2:])
				pubKeys = append(pubKeys, keys)

				_, err = clients[i].DockerExec(localId, "bash -c 'echo \"[\\\"" + addrs + "\\\"]\" >> /pantheon/data/toEncode.json'")
				if err != nil {
						log.Println(err)
						return nil, err
				}

				_, err = clients[i].DockerExecd(localId, "mkdir /pantheon/genesis")
				if err != nil {
						log.Println(err)
						return nil, err
				}

				// used for IBFT2 extraData
				_, err = clients[i].DockerExec(localId, "pantheon rlp encode --from=/pantheon/data/toEncode.json --to=/pantheon/rlpEncodedExtraData")
				if err != nil {
						log.Println(err)
						return nil, err
				}
				
				buildState.IncrementBuildProgress()

		}

		//used for IBFT2 extraData
		/*
		validator, err := clients[0].DockerExec(0, "cat /pantheon/rlpEncodedExtraData")
		if err != nil {
				log.Println(err)
				return nil, err
		}
		*/

		/* Create Genesis File */
		buildState.SetBuildStage("Generating Genesis File")
		err = createGenesisfile(panconf,details,addresses)
		if err != nil{
			log.Println(err)
			return nil,err
		}
		defer util.Rm("./genesis.json")

		port := 30303
		enodes := "["
		var enodeAddress string
		for _, server := range servers {
			for i, ip := range server.Ips {
				enodeAddress = fmt.Sprintf("enode://%s@%s:%d",
				pubKeys[i],
				ip,
				port,
			)
				if i < len(pubKeys)-1 {
					enodes = enodes + "\"" + enodeAddress + "\"" + ","
				} else {
					enodes = enodes + "\"" + enodeAddress + "\""
				}
				port++
				buildState.IncrementBuildProgress()
			}
		}
		enodes = enodes + "]"

		/* Create Static Nodes File */
		buildState.SetBuildStage("Setting Up Static Peers")
		buildState.IncrementBuildProgress()
		err = createStaticNodesFile(enodes)
		if err != nil{
			log.Println(err)
			return nil,err
		}
		defer util.Rm("./static-nodes.json")

		/* Copy static-nodes & genesis files to each node */
		buildState.SetBuildStage("Distributing Files")
		for i, server := range servers {
			err = clients[i].Scp("./static-nodes.json", "/home/appo/static-nodes.json")
			if err != nil {
				log.Println(err)
				return nil, err
			}
			defer clients[i].Run("rm /home/appo/static-nodes.json")

			err = clients[i].Scp("./genesis.json", "/home/appo/genesis.json")
			if err != nil {
				log.Println(err)
				return nil, err
			}
			defer clients[i].Run("rm /home/appo/genesis.json")
	
			for j, _ := range server.Ips {
				err := clients[i].DockerCp(j,"/home/appo/static-nodes.json","/pantheon/data/static-nodes.json")
				if err != nil {
					log.Println(err)
					return nil,err
				}
				err = clients[i].DockerCp(j,"/home/appo/genesis.json","/pantheon/genesis/genesis.json")
				if err != nil {
					log.Println(err)
					return nil,err
				}
			}
			buildState.IncrementBuildProgress()
		}

		/* Start the nodes */
		buildState.SetBuildStage("Starting Pantheon")
		p2pPort := 30303
		httpPort := 8545
		for i, server := range servers {
			for localId, _ := range server.Ips {
				pantheonCmd := fmt.Sprintf(
					`pantheon --data-path /pantheon/data --genesis-file=/pantheon/genesis/genesis.json --rpc-http-enabled --rpc-http-api="ADMIN,CLIQUE,DEBUG,EEA,ETH,IBFT,MINER,NET,WEB3" ` +
						` --p2p-port=%d --rpc-http-port=%d --rpc-http-host="0.0.0.0" --host-whitelist=all --rpc-http-cors-origins="*"`,
					p2pPort,
					httpPort,
					)
				err := clients[i].DockerExecdLog(localId, pantheonCmd)
				if err != nil {
					log.Println(err)
					return nil, err
				}
				buildState.IncrementBuildProgress()
			}
		}
	}
	return nil, nil
}

func createGenesisfile(panconf *PanConf, details db.DeploymentDetails, address []string) error {
	genesis := map[string]interface{}{
		"chainId":        		panconf.NetworkId,
		"difficulty":     		fmt.Sprintf("0x0%X", panconf.Difficulty),
		"gasLimit":       		fmt.Sprintf("0x0%X", panconf.GasLimit),
		"blockPeriodSeconds": 	panconf.BlockPeriodSeconds,
		"epoch":				panconf.Epoch,
		// "extraData": 	  validators, //for IBFT2
	}
	alloc := map[string]map[string]string{}
	for _, addr := range address {
		alloc[addr] = map[string]string{
			"balance": panconf.InitBalance,
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
	fmt.Println("Writing Genesis File Locally")
	return util.Write("genesis.json", data)

}

func createStaticNodesFile(list string) error {
	return util.Write("static-nodes.json", list)
}