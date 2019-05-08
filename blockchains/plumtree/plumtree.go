/*
	Copyright 2019 Whiteblock Inc.
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

//Package plumtree handles artplumtreeemis specific functionality
package plumtree

import (
	"fmt"

	"../../db"
	"../../ssh"
	"../../testnet"
	"../../util"
	"../helpers"
	"../registrar"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
	blockchain := "plumtree"
	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, GetDefaults)
	registrar.RegisterParams(blockchain, GetParams)
	registrar.RegisterAdditionalLogs(blockchain, map[string]string{
		"json": "/plumtree/data/log.json"})
}

// build builds out a fresh new plumtree test network
func build(tn *testnet.TestNet) error {

	tn.BuildState.SetBuildSteps(0 + (tn.LDD.Nodes * 4))

	port := 9000
	peers := ""
	var peer string
	for i, node := range tn.Nodes {
		peer = fmt.Sprintf("tcp://whiteblock-node%d@%s:%d",
			node.LocalID,
			node.IP,
			port,
		)
		if i != len(tn.Nodes)-1 {
			peers = peers + " " + peer + " "
		} else {
			peers = peers + " " + peer
		}
		tn.BuildState.IncrementBuildProgress()
	}

	peers = peers
	fmt.Println(peers)

	tn.BuildState.SetBuildStage("Starting plumtree")
	err := helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()

		artemisCmd := "gossip -n 0.0.0.0 -l 9000 -r 9001 -m /plumtree/data/log.json --peer=" + peers + " 2>&1 | tee /output.log"

		_, err := client.DockerExecd(node, "tmux new -s whiteblock -d")
		if err != nil {
			return util.LogError(err)
		}

		_, err = client.DockerExecd(node, fmt.Sprintf("tmux send-keys -t whiteblock '%s' C-m", artemisCmd))
		return err
	})
	if err != nil {
		return util.LogError(err)
	}

	return nil
}

// Add handles adding a node to the artemis testnet
// TODO
func add(tn *testnet.TestNet) error {
	return nil
}
