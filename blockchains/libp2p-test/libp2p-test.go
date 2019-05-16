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

//Package libp2ptest handles libp2ptest specific functionality
package libp2ptest

import (
	"encoding/json"
	"fmt"
	"github.com/whiteblock/genesis/blockchains/helpers"
	"github.com/whiteblock/genesis/blockchains/registrar"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"regexp"
	"sync"
)

var conf *util.Config

const blockchain = "libp2p-test"

func init() {
	conf = util.GetConfig()

	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterServices(blockchain, func() []util.Service { return []util.Service{} })
	registrar.RegisterDefaults(blockchain, helpers.DefaultGetDefaultsFn(blockchain))
	registrar.RegisterParams(blockchain, helpers.DefaultGetParamsFn(blockchain))
}

type serialPeerInfo struct {
	ID     string   `json:"pid"`
	MAddrs []string `json:"addrs"`
}

// build builds out a fresh new prysm test network
func build(tn *testnet.TestNet) error {
	testConf, err := newConf(tn.LDD.Params)
	if err != nil {
		return util.LogError(err)
	}
	if testConf.Connections <= 0 {
		testConf.Connections = tn.LDD.Nodes - 1
	}
	peers := make([]serialPeerInfo, tn.LDD.Nodes)
	mux := &sync.Mutex{}
	re := regexp.MustCompile(`(?m)(.*)Created a client(.*)`)

	//Get the peer information
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		res, err := client.DockerExec(node,
			fmt.Sprintf("./p2p-tests --generate-only --seed %d --hostAddrs /ip4/%s/tcp/39977", node.GetAbsoluteNumber()+1, node.GetIP()))
		if err != nil {
			return util.LogError(err)
		}
		matches := re.FindAllString(res, 1)
		if len(matches) == 0 {
			return fmt.Errorf("unexpected result: %s", res)
		}
		var peer serialPeerInfo
		err = json.Unmarshal([]byte(matches[0]), &peer)
		if err != nil {
			return util.LogError(err)
		}
		//fmt.Printf("PEER=%#v\n", peer)
		mux.Lock()
		peers[node.GetAbsoluteNumber()] = peer
		mux.Unlock()
		return nil
	})

	if err != nil {
		return util.LogError(err)
	}

	mesh, err := util.GenerateDependentMeshNetwork(tn.LDD.Nodes, testConf.Connections)
	if err != nil {
		return util.LogError(err)
	}

	err = helpers.CreateConfigs(tn, "/p2p-tests/static-peers.json", func(node ssh.Node) ([]byte, error) {
		nodePeers := []serialPeerInfo{}
		for _, peerIndex := range mesh[node.GetAbsoluteNumber()] {
			nodePeers = append(nodePeers, peers[peerIndex])
		}
		return json.Marshal(nodePeers)
	})

	if err != nil {
		return util.LogError(err)
	}

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		cmd := fmt.Sprintf("/p2p-tests/p2p-tests --seed %d --hostAddrs /ip4/%s/tcp/39977 "+
			"--file /p2p-tests/static-peers.json --pubsubRouter %s",
			node.GetAbsoluteNumber()+1, node.GetIP(), testConf.Router)
		if node.GetAbsoluteNumber() < testConf.Senders { //make node 0 the sending node
			cmd += fmt.Sprintf(" --send-interval %d", testConf.Interval)
		}
		if testConf.PayloadSize > 0 {
			cmd += fmt.Sprintf(" --payload-size %d", testConf.PayloadSize)
		}
		_, err := client.DockerExecdit(node, fmt.Sprintf("bash -ic '%s 2>&1 | tee %s'", cmd, conf.DockerOutputFile))
		return err
	})

	return util.LogError(err)
}

// add handles adding nodes to the testnet
func add(tn *testnet.TestNet) error {
	return nil
}
