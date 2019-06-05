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

//Package polkadot handles polkadot specific functionality
package polkadot

import (
	"fmt"
	"strings"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/blockchains/helpers"
	"github.com/whiteblock/genesis/blockchains/registrar"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"regexp"
)

var conf *util.Config

const blockchain = "polkadot"

func init() {
	conf = util.GetConfig()
	alias := "polkadot"

	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterBuild(alias, build) 

	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterAddNodes(alias, add)

	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterServices(alias, GetServices)

	registrar.RegisterDefaults(blockchain, helpers.DefaultGetDefaultsFn(blockchain))
	registrar.RegisterDefaults(alias, helpers.DefaultGetDefaultsFn(blockchain))

	registrar.RegisterParams(blockchain, helpers.DefaultGetParamsFn(blockchain))
	registrar.RegisterParams(alias, helpers.DefaultGetParamsFn(blockchain))
}

// build builds out a fresh new polkadot test network
func build(tn *testnet.TestNet) error {
	// mux := sync.Mutex{}
	dotconf, err := newConf(tn.LDD.Params)
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.SetBuildSteps(8 + (5 * tn.LDD.Nodes))

	tn.BuildState.IncrementBuildProgress()

	tn.BuildState.SetBuildStage("Distributing secrets")

	helpers.MkdirAllNodes(tn, "/polkadot")

	var nodeIDList []string

	tn.BuildState.IncrementBuildProgress()

	tn.BuildState.SetBuildStage("Initializing polkadot")

	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		client.DockerExecd(node, fmt.Sprintf("bash -c 'polkadot --chain=local 2>&1 | tee %s'", conf.DockerOutputFile))
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}

	
	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		output, err := client.DockerRead(node, fmt.Sprintf("%s", conf.DockerOutputFile), -1)
		if err != nil {
				return util.LogError(err)
		}
		loop := true
		for loop {
			reNodeID := regexp.MustCompile(`(?m)Local node identity is: (.{46})`)
			fmt.Println(reNodeID)
			regNodeID := reNodeID.FindAllString(output,1)[0]
			splitNodeID := strings.Split(regNodeID, ":")
			nodeID := strings.Replace(splitNodeID[1], " ", "", -1)
			fmt.Println(nodeID)
			if len(reNodeID.FindAllString(output,1)) != 0 {
				loop = false
			}
			url := fmt.Sprintf("/ip4/%s/tcp/30333/p2p/%s", node.GetIP(), nodeID)
			nodeIDList = append(nodeIDList, url)
		}
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}

	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		client.DockerExec(node, fmt.Sprintf("pkill -f \"^polkadot\""))
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}

	//should delete output.log so there is no overlapping data (?)

	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Starting polkadot")

	nid := strings.Join(nodeIDList," ")

	fmt.Println(nid)
	
	var vmode string

	if (dotconf.ValidatorMode) {
		vmode = " --validator"
	}

	err = helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		client.DockerExecd(node, fmt.Sprintf("bash -c 'polkadot --chain=local %s --reserved-nodes %s 2>&1 | tee %s'", vmode, nid, conf.DockerOutputFile))
		if err != nil {
			return util.LogError(err)
		}
		log.WithFields(log.Fields{"node": node.GetAbsoluteNumber()}).Trace("creating block directory")
		tn.BuildState.IncrementBuildProgress()
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}
	return nil
}

/***************************************************************************************************************************/

// Add handles adding a node to the polkadot testnet
// TODO
func add(tn *testnet.TestNet) error {
	return nil
}

