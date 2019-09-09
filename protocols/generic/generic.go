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

	p2pPort := 9000

	tn.BuildState.SetBuildStage("Generating key pairs")

	nodeKeyPairs := map[string]crypto.PrivKey{}
	for _, node := range tn.Nodes {
		prvKey, _, _ := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
		nodeKeyPairs[node.ID] = prvKey
	}

	tn.BuildState.IncrementBuildProgress()

	libp2p := tn.LDD.Params["libp2p"] == "true"

	err := helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		if files, ok := tn.LDD.Params["files"].([]interface{}); ok {
			filesToCopy := files[node.GetRelativeNumber()]

			if fileMap, ok := filesToCopy.(map[string]string); ok {
				for src, target := range fileMap {
					err := client.Scp(src, target)
					if err != nil {
						return util.LogError(err)
					}
				}
			} else {
				err := errors.New(fmt.Sprintf("filesToCopy is a %v", reflect.TypeOf(filesToCopy).String()))
				return util.LogError(err)
			}
		} else {
			err := errors.New(fmt.Sprintf("tn.LDD.Params['files'] is not a map[string]string, it is a %v", reflect.TypeOf(tn.LDD.Params["files"]).String()))
			return util.LogError(err)
		}

		var params string

		if args, ok := tn.LDD.Params["args"].([]interface{}); ok {
			startArguments := args[node.GetRelativeNumber()]

			if argMap, ok := startArguments.(map[string]string); ok {
				for key, param := range argMap {
					params += fmt.Sprintf(" --%s %v", key, param)
				}
			} else {
				err := errors.New(fmt.Sprintf("startArguments is a %v", reflect.TypeOf(startArguments).String()))
				return util.LogError(err)
			}
		} else {
			err := errors.New(fmt.Sprintf("tn.LDD.Params['args'] is not a map[string]string, it is a %v", reflect.TypeOf(tn.LDD.Params["args"]).String()))
			return util.LogError(err)
		}

		params += fmt.Sprintf(" --port %d", p2pPort)

		params += " --peers "

		for _, peerNode := range tn.Nodes {
			if node.GetID() == peerNode.GetID() {
				continue
			}

			if libp2p {
				params += fmt.Sprintf(" /ip4/%s/tcp/%d/p2p/%s:%d", peerNode.IP, p2pPort, idString(nodeKeyPairs[peerNode.GetID()]), p2pPort)
			} else {
				params += fmt.Sprintf(" %s", peerNode.IP)
			}

			tn.BuildState.IncrementBuildProgress()
		}

		if libp2p {
			params += fmt.Sprintf(" --identity %s", nodeKeyPairs[node.GetID()])
		}

		log.WithField("args", params).Infof("Starting node %s", node.GetID())

		_, err := client.DockerExec(node, fmt.Sprintf("sh /launch/start.sh %s", params))
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
