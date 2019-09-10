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
	blockchain		= "generic"
	p2pPort			= 9000
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
	registrar.RegisterServices(blockchain, func() []services.Service { return nil })
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
		prvKey, _, _ := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
		nodeKeyPairs[node.ID] = prvKey
	}

	tn.BuildState.IncrementBuildProgress()

	topology, err := getTopology(tn)
	if err != nil {
		return util.LogError(err)
	}

	return buildNetwork(tn, nodeKeyPairs, topology)
}

func idString(k crypto.PrivKey) string {
	pid, err := peer.IDFromPrivateKey(k)
	if err != nil {
		panic(err)
	}
	return pid.Pretty()
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

		peers := createPeers(tn, node, nodeKeyPairs, networkTopology)
		if peers == "error" {
			return util.LogError(fmt.Errorf("peers could not be created"))
		}

		params += peers

		log.WithField("args", params).Infof("Starting node %s", node.GetID())

		launchScript, ok := tn.LDD.Params["launch-script"].([]interface{})
		if !ok {
			err := fmt.Errorf("tn.LDD.Params['launch-script'] is not a []interface{}, it is a %v", reflect.TypeOf(tn.LDD.Params["launch-script"]))
			return util.LogError(err)
		}

		script := launchScript[node.GetRelativeNumber()]
		buildParams := fmt.Sprintf("%v %s", script, params)

		_, err = client.DockerExec(node, buildParams)
		if err != nil {
			return util.LogError(err)
		}

		tn.BuildState.IncrementBuildProgress()
		return nil
	})
}

func createPeers(tn *testnet.TestNet, node ssh.Node, nodeKeyPairs map[string]crypto.PrivKey, networkTopology topology) string {
	libp2p := tn.LDD.Params["libp2p"] == "true"

	var out string

	switch networkTopology {
	case all:
		for _, peerNode := range tn.Nodes {
			if node.GetID() == peerNode.GetID() {
				continue
			}

			if libp2p {
				out += fmt.Sprintf(" --peers=/ip4/%s/tcp/%d/p2p/%s:%d", peerNode.IP, p2pPort, idString(nodeKeyPairs[peerNode.GetID()]), p2pPort)
			} else {
				out += fmt.Sprintf(" --peers=%s", peerNode.IP)
			}

			tn.BuildState.IncrementBuildProgress()
		}

		if libp2p {
			out += fmt.Sprintf(" --identity=%s", idString(nodeKeyPairs[node.GetID()]))
		}

		return out
	case sequence:
		var peerNode db.Node

		if node.GetRelativeNumber()+1 >= len(tn.Nodes) {
			peerNode = tn.Nodes[0]
		} else {
			peerNode = tn.Nodes[node.GetRelativeNumber()+1]
		}

		if libp2p {
			out += fmt.Sprintf(" --peers=/ip4/%s/tcp/%d/p2p/%s:%d", peerNode.IP, p2pPort, idString(nodeKeyPairs[peerNode.GetID()]), p2pPort)
			out += fmt.Sprintf(" --identity=%s", idString(nodeKeyPairs[node.GetID()]))
		} else {
			out += fmt.Sprintf(" --peers=%s", peerNode.IP)
		}

		return out
	case randomTwo:
		return " " // TODO
	default:
		return "error"
	}
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
}j

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
		err := fmt.Errorf("startArguments is a %v", reflect.TypeOf(startArguments).String())
		return "", util.LogError(err)
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
