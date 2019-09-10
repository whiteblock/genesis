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
	"errors"
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
	blockchain     = "generic"
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

	err := helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		files, ok := tn.LDD.Params["files"].([]interface{})
		if !ok {
			err := fmt.Errorf("tn.LDD.Params['files'] is not a map[string]string, it is a %v", reflect.TypeOf(tn.LDD.Params["files"]).String())
			return util.LogError(err)
		}

		filesToCopy := files[node.GetRelativeNumber()]

		fileMap, ok := filesToCopy.(map[string]interface{})
		if !ok {
			err := errors.New(fmt.Sprintf("filesToCopy is a %v", reflect.TypeOf(filesToCopy).String()))
			return util.LogError(err)
		}

		for src, target := range fileMap {
			err := client.DockerCp(node, src, fmt.Sprintf("%v", target))
			if err != nil {
				return util.LogError(err)
			}
		}

		params, err := createParams(tn, node, nodeKeyPairs)

		launchScript, ok := tn.LDD.Params["launch-script"].([]interface{})
		if !ok {
			err := fmt.Errorf("tn.LDD.Params['launch-script'] is not a []interface{}, it is a %v", reflect.TypeOf(tn.LDD.Params["launch-script"]))
			return util.LogError(err)
		}

		script := launchScript[node.GetRelativeNumber()]

		_, err = client.DockerExec(node, fmt.Sprintf("%v %s", script, params))
		if err != nil {
			return util.LogError(err)
		}

		tn.BuildState.IncrementBuildProgress()
		return nil
	})

	return err
}

func idString(k crypto.PrivKey) string {
	pid, err := peer.IDFromPrivateKey(k)
	if err != nil {
		panic(err)
	}
	return pid.Pretty()
}

func add(tn *testnet.TestNet) error {
	err := build(tn)

	return err
}

func createParams(tn *testnet.TestNet, node ssh.Node, nodeKeyPairs map[string]crypto.PrivKey) (string, error) {
	p2pPort := 9000
	libp2p := tn.LDD.Params["libp2p"] == "true"

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

	for _, peerNode := range tn.Nodes {
		if node.GetID() == peerNode.GetID() {
			continue
		}

		if libp2p {
			params += fmt.Sprintf(" --peers=/ip4/%s/tcp/%d/p2p/%s:%d", peerNode.IP, p2pPort, idString(nodeKeyPairs[peerNode.GetID()]), p2pPort)
		} else {
			params += fmt.Sprintf(" --peers=%s", peerNode.IP)
		}

		tn.BuildState.IncrementBuildProgress()
	}

	if libp2p {
		params += fmt.Sprintf(" --identity=%s", idString(nodeKeyPairs[node.GetID()]))
	}

	log.WithField("args", params).Infof("Starting node %s", node.GetID())

	return params, nil
}
