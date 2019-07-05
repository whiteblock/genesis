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

// Package artemis handles artemis specific functionality
package artemis

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/registrar"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	// "reflect"
	"strings"
)

var conf *util.Config

const blockchain = "artemis"

func init() {
	conf = util.GetConfig()
	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, helpers.DefaultGetDefaultsFn(blockchain))
	registrar.RegisterParams(blockchain, helpers.DefaultGetParamsFn(blockchain))
	registrar.RegisterAdditionalLogs(blockchain, map[string]string{
		"json": "/artemis/data/log.json"})
}

// build builds out a fresh new artemis test network
func build(tn *testnet.TestNet) error {
	aconf, err := newConf(tn.LDD.Params)
	if err != nil {
		return util.LogError(err)
	}
	fetchedConfChan := make(chan string)

	go func(aconf artemisConf) {
		res, err := util.HTTPRequest("GET", aconf["constantsSource"].(string), "")
		if err != nil {
			tn.BuildState.ReportError(err)
			return
		}
		fetchedConfChan <- string(res)

	}(aconf)

	tn.BuildState.SetBuildSteps(0 + (tn.LDD.Nodes * 4))

	port := 9000
	peers := "["
	var peer string
	for i, node := range tn.Nodes {
		peer = fmt.Sprintf("%s://whiteblock-node%d@%s:%d",
			aconf["networkMode"],
			node.LocalID,
			node.IP,
			port,
		)
		if i != len(tn.Nodes)-1 {
			peers = peers + "\"" + peer + "\"" + ","
		} else {
			peers = peers + "\"" + peer + "\""
		}
		tn.BuildState.IncrementBuildProgress()
	}

	peers = peers + "]"
	log.WithFields(log.Fields{"peers": peers}).Trace("generated the peers")

	tn.BuildState.SetBuildStage("Creating node configuration files")
	/**Create node config files**/
	fetchedConf := <-fetchedConfChan

	constantsIndex := strings.Index(fetchedConf, "[constants]")
	if constantsIndex == -1 {
		return util.LogError(fmt.Errorf("couldn't find \"[constants]\" in file fetched from given source"))
	}
	rawConstants := fetchedConf[constantsIndex:]
	err = helpers.CreateConfigs(tn, "/artemis/config/config.toml", func(node ssh.Node) ([]byte, error) {
		defer tn.BuildState.IncrementBuildProgress()
		identity := fmt.Sprintf("0x%.8x", node.GetAbsoluteNumber())
		artemisNodeConfig, err := makeNodeConfig(aconf, identity, peers, node.GetAbsoluteNumber(), tn.LDD, rawConstants)
		return []byte(artemisNodeConfig), err
	})
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.SetBuildStage("Starting Artemis")
	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, server *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		/*
		var logFolder string
		obj := tn.CombinedDetails.Params["logFolder"]
		if obj != nil && reflect.TypeOf(obj).Kind() == reflect.String {
			logFolder = obj.(string)
		} else {
			logFolder = ""
		}
		*/
		artemisCmd := fmt.Sprintf("artemis -c /artemis/config/config.toml 2>&1 | tee /output.log")

		_, err := client.DockerExecd(node, "tmux new -s whiteblock -d")
		if err != nil {
			return util.LogError(err)
		}

		_, err = client.DockerExecd(node, fmt.Sprintf("tmux send-keys -t whiteblock '%s' C-m", artemisCmd))
		return util.LogError(err)
	})
	return util.LogError(err)
}

// Add handles adding a node to the artemis testnet
// TODO
func add(tn *testnet.TestNet) error {

	var prysymIPList []string
	tn.BuildState.GetP("IPList", &prysymIPList)

	var prysmP2PPort int64
	tn.BuildState.GetP("prysmP2PPort", &prysmP2PPort)

	aconf, err := newConf(tn.LDD.Params)
	if err != nil {
		return util.LogError(err)
	}
	fetchedConfChan := make(chan string)

	go func(aconf artemisConf) {
		res, err := util.HTTPRequest("GET", aconf["constantsSource"].(string), "")
		if err != nil {
			tn.BuildState.ReportError(err)
			return
		}
		fetchedConfChan <- string(res)

	}(aconf)

	tn.BuildState.SetBuildSteps(0 + (tn.LDD.Nodes * 4))

	artemisPort := 9000
	peers := "["
	var peer string
	for i, node := range tn.Nodes {
		peer = fmt.Sprintf("%s://whiteblock-node%d@%s:%d",
			aconf["networkMode"],
			node.LocalID,
			node.IP,
			artemisPort,
		)
		if i != len(tn.Nodes)-1 {
			peers = peers + "\"" + peer + "\"" + ","
		} else {
			peers = peers + "\"" + peer + "\""
		}
		tn.BuildState.IncrementBuildProgress()
	}

	peers = peers + ","
	for j, nodeIP := range prysymIPList {
		peer = fmt.Sprintf("%s://whiteblock-node%d@%s:%d",
		aconf["networkMode"],
		j,
		nodeIP,
		prysmP2PPort,
		)
		if j != len(prysymIPList)-1 {
			peers = peers + "\"" + peer + "\"" + ","
		} else {
			peers = peers + "\"" + peer + "\""
		}
		tn.BuildState.IncrementBuildProgress()
	}
	peers = peers + "]"
	log.WithFields(log.Fields{"peers": peers}).Trace("generated the peers")

	tn.BuildState.SetBuildStage("Creating node configuration files")
	/**Create node config files**/
	fetchedConf := <-fetchedConfChan

	constantsIndex := strings.Index(fetchedConf, "[constants]")
	if constantsIndex == -1 {
		return util.LogError(fmt.Errorf("couldn't find \"[constants]\" in file fetched from given source"))
	}
	rawConstants := fetchedConf[constantsIndex:]
	err = helpers.CreateConfigsNewNodes(tn, "/artemis/config/config.toml", func(node ssh.Node) ([]byte, error) {
		defer tn.BuildState.IncrementBuildProgress()
		identity := fmt.Sprintf("0x%.8x", node.GetAbsoluteNumber())
		artemisNodeConfig, err := makeNodeConfig(aconf, identity, peers, node.GetAbsoluteNumber(), tn.LDD, rawConstants)
		return []byte(artemisNodeConfig), err
	})
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.SetBuildStage("Starting Artemis")
	err = helpers.AllNewNodeExecCon(tn, func(client ssh.Client, server *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		/*
		var logFolder string
		obj := tn.CombinedDetails.Params["logFolder"]
		if obj != nil && reflect.TypeOf(obj).Kind() == reflect.String {
			logFolder = obj.(string)
		} else {
			logFolder = ""
		}
		*/
		artemisCmd := fmt.Sprintf("artemis -c /artemis/config/config.toml 2>&1 | tee /output.log")

		_, err := client.DockerExecd(node, "tmux new -s whiteblock -d")
		if err != nil {
			return util.LogError(err)
		}

		_, err = client.DockerExecd(node, fmt.Sprintf("tmux send-keys -t whiteblock '%s' C-m", artemisCmd))
		return util.LogError(err)
	})
	return util.LogError(err)
}
