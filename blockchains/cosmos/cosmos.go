package cosmos

import (
	db "../../db"
	state "../../state"
	util "../../util"
	helpers "../helpers"
	"fmt"
	"log"
	"strings"
	"sync"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

func Build(details *db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
	buildState *state.BuildState) ([]string, error) {

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
	peers := make([]string, details.Nodes)
	mux := sync.Mutex{}

	err = helpers.AllNodeExecCon(servers, buildState, func(serverNum int, localNodeNum int, absoluteNodeNum int) error {
		ip := servers[serverNum].Ips[localNodeNum]
		if absoluteNodeNum != 0 {
			//init everything
			_, err = clients[serverNum].DockerExec(localNodeNum, "gaiad init --chain-id=whiteblock whiteblock")
			if err != nil {
				log.Println(res)
				return err
			}
		}

		//Get the node id
		res, err = clients[serverNum].DockerExec(localNodeNum, "gaiad tendermint show-node-id")
		if err != nil {
			log.Println(err)
			return err
		}
		nodeId := res[:len(res)-1]
		mux.Lock()
		peers[absoluteNodeNum] = fmt.Sprintf("%s@%s:26656", nodeId, ip)
		mux.Unlock()
		buildState.IncrementBuildProgress()
		return nil
	})

	buildState.SetBuildStage("Copying the genesis file to each node")

	err = helpers.CopyToAllNodes(servers, clients, buildState, genesisFile, "/root/.gaiad/config/genesis.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	buildState.SetBuildStage("Starting cosmos")

	err = helpers.AllNodeExecCon(servers, buildState, func(serverNum int, localNodeNum int, absoluteNodeNum int) error {
		defer buildState.IncrementBuildProgress()
		_, err := clients[serverNum].DockerExecd(localNodeNum, fmt.Sprintf("gaiad start --p2p.persistent_peers=%s",
			strings.Join(append(peers[:absoluteNodeNum], peers[absoluteNodeNum+1:]...), ",")))
		return err
	})
	return nil, err
}

func Add(details *db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
	newNodes map[int][]string, buildState *state.BuildState) ([]string, error) {
	return nil, nil
}
