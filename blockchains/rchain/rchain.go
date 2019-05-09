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

//Package rchain handles rchain specific functionality
package rchain

import (
	"github.com/Whiteblock/genesis/db"
	"github.com/Whiteblock/genesis/ssh"
	"github.com/Whiteblock/genesis/testnet"
	"github.com/Whiteblock/genesis/util"
	"github.com/Whiteblock/genesis/blockchains/helpers"
	"github.com/Whiteblock/genesis/blockchains/registrar"
	"fmt"
	"github.com/Whiteblock/mustache"
	"log"
	"regexp"
	"sync"
	"time"
)

var conf *util.Config

const blockchain = "rchain"

func init() {
	conf = util.GetConfig()

	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, GetDefaults)
	registrar.RegisterParams(blockchain, GetParams)
}

// build builds out a fresh new rchain test network
func build(tn *testnet.TestNet) error {
	buildState := tn.BuildState
	masterNode := tn.Nodes[0]
	masterClient := tn.Clients[masterNode.Server]

	rConf, err := newRChainConf(tn.LDD.Params)
	if err != nil {
		return util.LogError(err)
	}
	buildState.SetBuildSteps(9 + (len(tn.Servers) * 2) + (tn.LDD.Nodes * 2))
	buildState.SetBuildStage("Setting up data collection")

	services, err := util.GetServiceIps(GetServices())
	buildState.IncrementBuildProgress()
	if err != nil {
		return util.LogError(err)
	}

	/**Make the data directories**/

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		buildState.IncrementBuildProgress()
		_, err := client.DockerExec(node, "mkdir /datadir")
		return err
	})
	/**Setup the first node**/
	err = createFirstConfigFile(tn, masterClient, masterNode, rConf, services["wb_influx_proxy"])
	if err != nil {
		return util.LogError(err)
	}
	/**Check to make sure the rnode command is valid**/
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		_, err := client.DockerExec(node, fmt.Sprintf("bash -c '%s --help'", rConf.Command))
		if err != nil {
			//log.Println(err)
			return fmt.Errorf("could not find command \"%s\"", rConf.Command)
		}
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}

	buildState.IncrementBuildProgress()
	km, err := helpers.NewKeyMaster(tn.LDD, blockchain)
	keyPairs := make([]util.KeyPair, tn.LDD.Nodes)
	validatorKeyPairs := make([]util.KeyPair, rConf.Validators)
	for i := range keyPairs {
		keyPairs[i], err = km.GetKeyPair(masterClient)
		if i > 0 && i-1 < len(validatorKeyPairs) {
			validatorKeyPairs[i-1] = keyPairs[i]
		}
		if err != nil {
			return util.LogError(err)
		}
	}
	for i := len(keyPairs) - 1; i < len(validatorKeyPairs); i++ {
		validatorKeyPairs[i], err = km.GetKeyPair(masterClient)
		if err != nil {
			return util.LogError(err)
		}
	}
	//fmt.Printf("Keypairs = %#v\n", keyPairs)
	//fmt.Printf("BalidatorKeyPairs = %#v\n", validatorKeyPairs)
	buildState.Set("keyPairs", keyPairs)
	buildState.Set("validatorKeyPairs", validatorKeyPairs)

	buildState.IncrementBuildProgress()

	buildState.SetBuildStage("Setting up bonds")
	/**Setup bonds**/
	{
		bonds := make([]string, len(validatorKeyPairs))
		for i, keyPair := range validatorKeyPairs {
			bonds[i] = fmt.Sprintf("%s %d", keyPair.PublicKey, rConf.BondsValue)
		}
		buildState.IncrementBuildProgress()
		err = buildState.Write("bonds.txt", util.CombineConfig(bonds))
		if err != nil {
			return util.LogError(err)
		}
		buildState.IncrementBuildProgress()

		err = masterClient.Scp("bonds.txt", "/tmp/bonds.txt")
		if err != nil {
			return util.LogError(err)
		}
		buildState.IncrementBuildProgress()
		buildState.Defer(func() { masterClient.Run("rm -f /tmp/bonds.txt") })

		err = masterClient.DockerCp(masterNode, "/tmp/bonds.txt", "/bonds.txt")
		if err != nil {
			return util.LogError(err)
		}
		buildState.IncrementBuildProgress()

	}

	buildState.SetBuildStage("Starting the boot node")
	var enode string
	{
		err = masterClient.DockerExecdLog(masterNode,
			fmt.Sprintf("%s run --standalone --data-dir \"/datadir\" --host %s --bonds-file /bonds.txt --has-faucet",
				rConf.Command, masterNode.IP))
		buildState.IncrementBuildProgress()
		if err != nil {
			return util.LogError(err)
		}
		//fmt.Println("Attempting to get the enode address")
		buildState.SetBuildStage("Waiting for the boot node's address")
		for i := 0; i < 1000; i++ {
			fmt.Println("Checking if the boot node is ready...")
			time.Sleep(time.Duration(1 * time.Second))
			output, err := masterClient.DockerExec(masterNode, fmt.Sprintf("cat %s", conf.DockerOutputFile))
			if err != nil {
				return util.LogError(err)
			}
			re := regexp.MustCompile(`(?m)rnode:\/\/[a-z|0-9]*\@([0-9]{1,3}\.){3}[0-9]{1,3}\?protocol=[0-9]*\&discovery=[0-9]*`)

			if !re.MatchString(output) {
				fmt.Println("Not ready")
				continue
			}
			enode = re.FindAllString(output, 1)[0]
			fmt.Println("Ready")
			break
		}
		buildState.IncrementBuildProgress()
		/*
		   influxIp
		   validators
		*/
		log.Println("Got the address for the bootnode: " + enode)
	}
	buildState.Set("bootnode", enode)
	buildState.Set("rConf", *rConf)

	err = helpers.CreateConfigs(tn, "/datadir/rnode.conf", func(node ssh.Node) ([]byte, error) {
		if node.GetAbsoluteNumber() == 0 {
			return nil, nil
		}
		return createConfigFile(tn, enode, rConf, services["wb_influx_proxy"], node.GetAbsoluteNumber())
	})
	if err != nil {
		return util.LogError(err)
	}
	buildState.SetBuildStage("Configuring the other rchain nodes")
	/**Copy config files to the rest of the nodes**/
	buildState.IncrementBuildProgress()

	buildState.SetBuildStage("Starting the rest of the nodes")
	/**Start up the rest of the nodes**/
	mux := sync.Mutex{}

	var validators int64

	return helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, node ssh.Node) error {
		defer buildState.IncrementBuildProgress()
		if node.GetAbsoluteNumber() == 0 {
			return nil
		}
		mux.Lock()
		isValidator := validators < rConf.Validators
		validators++
		mux.Unlock()
		if isValidator {
			return client.DockerExecdLog(node,
				fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --validator-private-key %s --host %s",
					rConf.Command, enode, keyPairs[node.GetAbsoluteNumber()].PrivateKey, node.GetIP()))
		}
		return client.DockerExecdLog(node,
			fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --host %s",
				rConf.Command, enode, node.GetIP()))
	})
}

func createFirstConfigFile(tn *testnet.TestNet, client *ssh.Client, node ssh.Node, rConf *rChainConf, influxIP string) error {

	raw := map[string]interface{}{
		"influxIp":       influxIP,
		"validatorCount": rConf.ValidatorCount,
		"standalone":     true,
	}
	raw = util.MergeStringMaps(raw, tn.LDD.Params) //Allow arbitrary custom options for rchain

	filler := util.ConvertToStringMap(raw)
	dat, err := helpers.GetBlockchainConfig("rchain", 0, "rchain.conf.mustache", tn.LDD)
	if err != nil {
		return util.LogError(err)
	}
	data, err := mustache.Render(string(dat), filler)
	if err != nil {
		return util.LogError(err)
	}
	err = tn.BuildState.Write("rnode.conf", data)
	if err != nil {
		return util.LogError(err)
	}
	err = client.Scp("rnode.conf", "/tmp/rnode.conf")
	tn.BuildState.Defer(func() { client.Run("rm -f /tmp/rnode.conf") })
	if err != nil {
		return util.LogError(err)
	}
	return client.DockerCp(node, "/tmp/rnode.conf", "/datadir/rnode.conf")
}

/**********************************************************************ADD********************************************************************/

// add handles the addition of nodes to the rchain testnet
func add(tn *testnet.TestNet) error {

	rConf, err := newRChainConf(tn.CombinedDetails.Params)
	tn.BuildState.SetBuildSteps(1 + 2*len(tn.NewlyBuiltNodes)) //TODO
	if err != nil {
		return util.LogError(err)
	}
	iEnode, ok := tn.BuildState.Get("bootnode")
	if !ok {
		return util.LogError(fmt.Errorf("rebuild: missing bootnode"))
	}
	enode := iEnode.(string)

	services, err := util.GetServiceIps(GetServices())
	if err != nil {
		return util.LogError(err)
	}
	keyPairs := []util.KeyPair{}
	km, err := helpers.NewKeyMaster(&tn.CombinedDetails, "rchain")
	if err != nil {
		return util.LogError(err)
	}
	for range tn.Nodes {
		kp, err := km.GetKeyPair(tn.GetFlatClients()[0])
		if err != nil {
			return util.LogError(err)
		}
		keyPairs = append(keyPairs, kp)
	}

	helpers.AllNewNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		client.DockerExec(node, "mkdir /datadir") //Don't bother checking for errors, ok if dir exists
		return nil
	})

	err = helpers.CreateConfigsNewNodes(tn, "/datadir/rnode.conf", func(node ssh.Node) ([]byte, error) {
		return createConfigFile(tn, enode, rConf, services["wb_influx_proxy"], node.GetAbsoluteNumber())
	})
	if err != nil {
		return util.LogError(err)
	} //VRFY check why this commented out section existed
	/*if !ok {
		return nil, util.LogError(fmt.Errorf("rebuild: missing key pairs"))
	}*/
	tn.BuildState.SetBuildStage("Configuring the other rchain nodes")
	/**Copy config files to the rest of the nodes**/
	tn.BuildState.IncrementBuildProgress()

	tn.BuildState.SetBuildStage("Starting the rest of the nodes")
	/**Start up the rest of the nodes**/
	var validators int64
	mux := sync.Mutex{}
	return helpers.AllNewNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()

		mux.Lock()
		isValidator := validators < rConf.Validators
		validators++
		mux.Unlock()

		if isValidator {
			err = client.DockerExecdLog(node,
				fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --validator-private-key %s --host %s",
					rConf.Command, enode, keyPairs[node.GetAbsoluteNumber()].PrivateKey, node.GetIP()))
			return err
		}
		return client.DockerExecdLog(node,
			fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --host %s",
				rConf.Command, enode, node.GetIP()))
	})
}

func createConfigFile(tn *testnet.TestNet, bootnodeAddr string, rConf *rChainConf,
	influxIP string, node int) ([]byte, error) {

	raw := map[string]interface{}{
		"influxIp":       influxIP,
		"validatorCount": rConf.ValidatorCount,
		"standalone":     false,
	}

	raw = util.MergeStringMaps(raw, tn.CombinedDetails.Params) //Allow arbitrary custom options for rchain

	filler := util.ConvertToStringMap(raw)
	filler["bootstrap"] = fmt.Sprintf("bootstrap = \"%s\"", bootnodeAddr)
	dat, err := helpers.GetBlockchainConfig("rchain", node, "rchain.conf.mustache", &tn.CombinedDetails)
	if err != nil {
		return nil, util.LogError(err)
	}
	data, err := mustache.Render(string(dat), filler)
	if err != nil {
		return nil, util.LogError(err)
	}
	return []byte(data), nil
}
