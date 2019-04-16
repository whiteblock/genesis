package cosmos

import (
	db "../../db"
	state "../../state"
	util "../../util"
	helpers "../helpers"
	"fmt"
	"log"
	"strings"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

func Build(details *db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
	buildState *state.BuildState) ([]string, error) {

	peers := []string{}
	buildState.SetBuildSteps(4 + (details.Nodes * 2))

	buildState.SetBuildStage("Setting up the first node")
	/**
	 * Set up first node
	 */
	_, err := clients[0].DockerExec(0, "gaiad init --chain-id=whiteblock whiteblock")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	buildState.IncrementBuildProgress()
	_, err = clients[0].DockerExec(0, "bash -c 'echo \"password\\n\" | gaiacli keys add validator -ojson'")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	res, err := clients[0].DockerExec(0, "gaiacli keys show validator -a")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	buildState.IncrementBuildProgress()
	_, err = clients[0].DockerExec(0, fmt.Sprintf("gaiad add-genesis-account %s 100000000stake,100000000validatortoken", res[:len(res)-1]))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	_, err = clients[0].DockerExec(0, "bash -c 'echo \"password\\n\" | gaiad gentx --name validator'")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	buildState.IncrementBuildProgress()
	_, err = clients[0].DockerExec(0, "gaiad collect-gentxs")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	genesisFile, err := clients[0].DockerExec(0, "cat /root/.gaiad/config/genesis.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	buildState.IncrementBuildProgress()
	buildState.SetBuildStage("Initializing the rest of the nodes")
	node := 0
	for i, server := range servers {
		for j, ip := range server.Ips {
			if node != 0 {
				//init everything
				_, err = clients[i].DockerExec(j, "gaiad init --chain-id=whiteblock whiteblock")
				if err != nil {
					log.Println(res)
					return nil, err
				}
			}

			//Get the node id
			res, err = clients[i].DockerExec(j, "gaiad tendermint show-node-id")
			if err != nil {
				log.Println(err)
				return nil, err
			}
			nodeId := res[:len(res)-1]
			peers = append(peers, fmt.Sprintf("%s@%s:26656", nodeId, ip))

			buildState.IncrementBuildProgress()
			node++
		}
	}

	buildState.SetBuildStage("Copying the genesis file to each node")
	err = buildState.Write("genesis.json", genesisFile)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = helpers.CopyToAllNodes(servers, clients, buildState, "genesis.json", "/root/.gaiad/config/")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Printf("%v", peers)
	buildState.SetBuildStage("Starting cosmos")
	node = 0
	for i, server := range servers {
		for j, _ := range server.Ips {
			cmd := fmt.Sprintf("gaiad start --p2p.persistent_peers=%s",
				strings.Join(append(peers[:node], peers[node+1:]...), ","))
			_, err := clients[i].DockerExecd(j, cmd)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			node++
			buildState.IncrementBuildProgress()
		}
	}
	return nil, nil
}

func Add(details *db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
	newNodes map[int][]string, buildState *state.BuildState) ([]string, error) {
	return nil, nil
}
