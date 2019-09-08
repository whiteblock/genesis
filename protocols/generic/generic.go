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
	"github.com/libp2p/go-libp2p-core/peer"
	log "github.com/sirupsen/logrus"
	"path/filepath"

	"github.com/libp2p/go-libp2p-core/crypto"
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

	var params string

	var startArguments map[string]interface{}

	if args, ok := tn.LDD.Params["args"].(map[string]interface{}); ok {
		startArguments = args
		for key, param := range startArguments {
			params += fmt.Sprintf(" --%s %v", key, param)
		}
	} else {
		return util.LogError(errors.New("tn.LDD.Params['args'] is not a map[string]interface{}"))
	}

	params += fmt.Sprintf(" --port %d", p2pPort)

	libp2p := tn.LDD.Params["libp2p"] == "true"

	err := helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		if files, ok := tn.LDD.Params["files"].([]string); ok {
			for _, fileToCopy := range files {
				err := client.Scp(fileToCopy, fmt.Sprintf("/launch/%s", filepath.Base(fileToCopy)))
				if err != nil {
					return util.LogError(err)
				}
			}
		} else {
			return util.LogError(errors.New("tn.LDD.Params['files'] is not a []string"))
		}

		thisNodeParams := params + " --peers "

		for _, peerNode := range tn.Nodes {
			if node.GetID() == peerNode.GetID() {
				continue
			}

			if libp2p {
				thisNodeParams += fmt.Sprintf(" /ip4/%s/tcp/%d/p2p/%s:%d", peerNode.IP, p2pPort, idString(nodeKeyPairs[peerNode.GetID()]), p2pPort)
			} else {
				thisNodeParams += fmt.Sprintf(" %s", peerNode.IP)
			}

			tn.BuildState.IncrementBuildProgress()
		}

		if libp2p {
			thisNodeParams += fmt.Sprintf(" --identity %s", nodeKeyPairs[node.GetID()])
		}

		log.WithField("args", thisNodeParams).Infof("Starting node %s", node.GetID())

		_, err := client.DockerExec(node, fmt.Sprintf("sh /launch/start.sh %s", thisNodeParams))
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
