/*
	Copyright 2019 whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Genesis is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package generic

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"reflect"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/registrar"
	"github.com/whiteblock/genesis/protocols/services"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
)

const (
	blockchain = "generic"
	p2pPort    = 9000
)

type topology string

const (
	all       = topology("all")
	sequence  = topology("sequence")
	randomTwo = topology("randomTwo")
)

func init() {
	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterServices(blockchain, func() []services.Service { return []services.Service{services.RegisterPrometheus()} })
	registrar.RegisterDefaults(blockchain, helpers.DefaultGetDefaultsFn(blockchain))
	registrar.RegisterParams(blockchain, helpers.DefaultGetParamsFn(blockchain))
}

// build builds out a fresh new ethereum test network using geth
func build(tn *testnet.TestNet) error {
	tn.BuildState.SetBuildSteps(3 + tn.LDD.Nodes)
	tn.BuildState.SetBuildStage("Copying files inside the Docker containers")
	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Generating key pairs")

	nodeKeyPairs := map[string]crypto.PrivKey{}
	for _, node := range tn.Nodes {
		prvKey, _, _ := crypto.GenerateKeyPairWithReader(crypto.Secp256k1, 2048, rand.Reader)
		nodeKeyPairs[node.ID] = prvKey
	}

	tn.BuildState.IncrementBuildProgress()

	topology, err := getTopology(tn)
	if err != nil {
		return util.LogError(err)
	}

	return buildNetwork(tn, nodeKeyPairs, topology)
}

func getTopology(tn *testnet.TestNet) (topology, error) {
	networkTopology := fmt.Sprintf("%v", tn.LDD.Params["network-topology"])

	switch networkTopology {
	case "all":
		return all, nil
	case "sequence":
		return sequence, nil
	case "randomTwo":
		return randomTwo, nil
	default:
		return all, util.LogError(fmt.Errorf("unsupported network topology, %v", networkTopology))
	}
}

func buildNetwork(tn *testnet.TestNet, nodeKeyPairs map[string]crypto.PrivKey, networkTopology topology) error {
	return helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		err := copyFiles(tn, client, node)
		if err != nil {
			return util.LogError(err)
		}

		params, err := createDefaultParams(tn, node)

		libp2p := tn.LDD.Params["libp2p"] == "true"

		peerIds := map[int]string{}
		for _, peerNode := range tn.Nodes {
			if libp2p {
				id, err := publicKeyToBase58(nodeKeyPairs[peerNode.GetID()])
				if err != nil {
					return util.LogError(err)
				}
				peerIds[peerNode.GetRelativeNumber()] = fmt.Sprintf(" --peers=/ip4/%s/tcp/%d/p2p/%s", peerNode.IP, p2pPort, id)
			} else {
				peerIds[peerNode.GetRelativeNumber()] = fmt.Sprintf(" --peers=%s", peerNode.IP)
			}

		}

		peers, err := createPeers(node.GetRelativeNumber(), peerIds, networkTopology)

		if err != nil {
			return util.LogError(err)
		}

		params += peers

		if libp2p {
			id, err := privateKeyToHexString(nodeKeyPairs[node.GetID()])
			if err != nil {
				return util.LogError(err)
			}
			params += fmt.Sprintf(" --identity=%s", id)
		}

		log.WithField("args", params).Infof("Starting node %s", node.GetID())

		launchScript, ok := tn.LDD.Params["launch-script"].([]interface{})
		if !ok {
			err := fmt.Errorf("tn.LDD.Params['launch-script'] is not a []interface{}, it is a %v", reflect.TypeOf(tn.LDD.Params["launch-script"]))
			return util.LogError(err)
		}

		script := launchScript[node.GetRelativeNumber()]
		buildParams := fmt.Sprintf("%v %s", script, params)

		log.Infof("%s", buildParams)

		_, err = client.DockerExecd(node, buildParams)
		if err != nil {
			return util.LogError(err)
		}

		tn.BuildState.IncrementBuildProgress()
		return nil
	})
}

func createPeers(currentNodeIndex int, peerIds map[int]string, networkTopology topology) (string, error) {
	var out string

	switch networkTopology {
	case all:
		for i := 0; i < len(peerIds); i++ {
			if i == currentNodeIndex {
				continue
			}

			out += peerIds[i]
		}

		return out, nil
	case sequence:
		if currentNodeIndex+1 >= len(peerIds) {
			return "", nil
		}
		out += peerIds[currentNodeIndex+1]

		return out, nil
	case randomTwo:
		if currentNodeIndex < 2 {
			return "", nil
		}

		for i := currentNodeIndex - 1; i >= currentNodeIndex-2; i-- {
			out += peerIds[i]
		}

		return out, nil
	default:
		return "", fmt.Errorf("peers could not be created")
	}
}

func privateKeyToHexString(k crypto.PrivKey) (string, error) {
	bytes, err := k.Raw()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func publicKeyToBase58(k crypto.PrivKey) (string, error) {
	pid, err := peer.IDFromPrivateKey(k)
	if err != nil {
		return "", err
	}
	return pid.Pretty(), nil
}

func copyFiles(tn *testnet.TestNet, client ssh.Client, node ssh.Node) error {
	files, ok := tn.LDD.Params["files"].([]interface{})
	if !ok {
		err := fmt.Errorf("tn.LDD.Params['files'] is not a map[string]string, it is a %v", reflect.TypeOf(tn.LDD.Params["files"]).String())
		return util.LogError(err)
	}

	filesToCopy := files[node.GetRelativeNumber()]

	fileMap, ok := filesToCopy.(map[string]interface{})
	if !ok {
		err := fmt.Errorf("filesToCopy is a %v", reflect.TypeOf(filesToCopy).String())
		return util.LogError(err)
	}

	for src, target := range fileMap {
		err := client.DockerCp(node, src, fmt.Sprintf("%v", target))
		if err != nil {
			return util.LogError(err)
		}
	}

	return nil
}

func createDefaultParams(tn *testnet.TestNet, node ssh.Node) (string, error) {
	var params string

	args, ok := tn.LDD.Params["args"].([]interface{})
	if !ok {
		err := fmt.Errorf("tn.LDD.Params['args'] is not a map[string]string, it is a %v", reflect.TypeOf(tn.LDD.Params["args"]).String())
		return "", util.LogError(err)
	}

	startArguments := args[node.GetRelativeNumber()]

	argMap, ok := startArguments.(map[string]interface{})
	if !ok {
		return "", util.LogErrorf("startArguments is a %v", reflect.TypeOf(startArguments).String())
	}

	for key, param := range argMap {
		params += fmt.Sprintf(" --%s=%v", key, param)
	}

	params += fmt.Sprintf(" --port=%d", p2pPort)

	return params, nil
}

func add(tn *testnet.TestNet) error {
	err := build(tn)

	return err
}
