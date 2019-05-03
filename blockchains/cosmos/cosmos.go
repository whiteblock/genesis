//Package cosmos handles cosmos specific functionality
package cosmos

import (
	"../../db"
	"../../ssh"
	"../../testnet"
	"../../util"
	"../helpers"
	"../registrar"
	"fmt"
	"log"
	"strings"
	"sync"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
	blockchain := "cosmos"
	registrar.RegisterBuild(blockchain, Build)
	registrar.RegisterAddNodes(blockchain, Add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, GetDefaults)
	registrar.RegisterParams(blockchain, GetParams)
}

// Build builds out a fresh new cosmos test network
func Build(tn *testnet.TestNet) ([]string, error) {
	tn.BuildState.SetBuildSteps(4 + (tn.LDD.Nodes * 2))

	tn.BuildState.SetBuildStage("Setting up the first node")

	masterNode := tn.Nodes[0]
	masterClient := tn.Clients[masterNode.Server]
	/**
	 * Set up first node
	 */
	_, err := masterClient.DockerExec(tn.Nodes[0], "gaiad init --chain-id=whiteblock whiteblock")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	tn.BuildState.IncrementBuildProgress()
	_, err = masterClient.DockerExec(tn.Nodes[0], "bash -c 'echo \"password\\n\" | gaiacli keys add validator -ojson'")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	res, err := masterClient.DockerExec(tn.Nodes[0], "gaiacli keys show validator -a")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	tn.BuildState.IncrementBuildProgress()
	_, err = masterClient.DockerExec(tn.Nodes[0], fmt.Sprintf("gaiad add-genesis-account %s 100000000stake,100000000validatortoken", 
		   							 res[:len(res)-1]))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	_, err = masterClient.DockerExec(tn.Nodes[0], "bash -c 'echo \"password\\n\" | gaiad gentx --name validator'")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	tn.BuildState.IncrementBuildProgress()
	_, err = masterClient.DockerExec(tn.Nodes[0], "gaiad collect-gentxs")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	genesisFile, err := masterClient.DockerExec(tn.Nodes[0], "cat /root/.gaiad/config/genesis.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Initializing the rest of the nodes")
	peers := make([]string, tn.LDD.Nodes)
	mux := sync.Mutex{}

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, node ssh.Node) error {
		ip := tn.Nodes[node.GetAbsoluteNumber()].IP
		if node.GetAbsoluteNumber() != 0 {
			//init everything
			_, err := client.DockerExec(node, "gaiad init --chain-id=whiteblock whiteblock")
			if err != nil {
				log.Println(res)
				return err
			}
		}

		//Get the node id
		res, err := client.DockerExec(node, "gaiad tendermint show-node-id")
		if err != nil {
			log.Println(err)
			return err
		}
		nodeID := res[:len(res)-1]
		mux.Lock()
		peers[node.GetAbsoluteNumber()] = fmt.Sprintf("%s@%s:26656", nodeID, ip)
		mux.Unlock()
		tn.BuildState.IncrementBuildProgress()
		return nil
	})

	tn.BuildState.SetBuildStage("Copying the genesis file to each node")

	err = helpers.CopyBytesToAllNodes(tn, genesisFile, "/root/.gaiad/config/genesis.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tn.BuildState.SetBuildStage("Starting cosmos")

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server,node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		peersCpy := make([]string, len(peers))
		copy(peersCpy, peers)
		_, err := client.DockerExecd(node, fmt.Sprintf("gaiad start --p2p.persistent_peers=%s",
			strings.Join(append(peersCpy[:node.GetAbsoluteNumber()], peersCpy[node.GetAbsoluteNumber()+1:]...), ",")))
		return err
	})
	return nil, err
}

// Add handles adding a node to the cosmos testnet
// TODO
func Add(tn *testnet.TestNet) ([]string, error) {
	return nil, nil
}
