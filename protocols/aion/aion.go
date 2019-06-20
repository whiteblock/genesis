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

// Package aion handles artemis specific functionality
package aion

import (
	"fmt"
	"regexp"
	// log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/registrar"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"github.com/whiteblock/mustache"
	// "reflect"
	"strings"
	"sync"
	"encoding/json"
)

var conf *util.Config
const (
	blockchain     = "aion"
	password       = "password"
)

func init() {
	conf = util.GetConfig()
	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, helpers.DefaultGetDefaultsFn(blockchain))
	registrar.RegisterParams(blockchain, helpers.DefaultGetParamsFn(blockchain))
}

// build builds out a fresh new artemis test network
func build(tn *testnet.TestNet) error {
	mux := sync.Mutex{}
	aionconf, err := newConf(tn.LDD.Params)
	if err != nil {
		return util.LogError(err)
	}

	{
		/**Create the Password files**/
		var data string
		data += password + "\\n" + password
		/**Copy over the password file**/
		err = helpers.CopyBytesToAllNodes(tn, data, "/aion/passwd")
		if err != nil {
			return util.LogError(err)
		}
	}

	var addresses = make([]string, tn.LDD.Nodes)
	var nodeIDs []string

	err = helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		output, err := client.DockerExec(node, fmt.Sprintf("bash -c 'echo -e $(cat /aion/passwd) | /aion/./aion.sh ac -n custom'"))
		if err != nil {
			return util.LogError(err)
		}
		reAddr := regexp.MustCompile(`(?m)A new account has been created:(.{67})`)
		regAddr := reAddr.FindAllString(output, 1)[0]
		splitAddr := strings.Split(regAddr, "A new account has been created:")
		addr := strings.Replace(splitAddr[1], " ", "", -1)
		// fmt.Println(addr)
		mux.Lock()
		addresses[node.GetAbsoluteNumber()] = addr
			mux.Unlock()
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}

	//get permanent node id from auto-generated config.xml
	err = helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		output, err := client.DockerRead(node, fmt.Sprintf("/aion/custom/config/config.xml"), -1)
		if err != nil {
			return util.LogError(err)
		}
		reNodeID := regexp.MustCompile(`(?m)<id>(.{36})`)
		regNodeID := reNodeID.FindAllString(output, 1)[0]
		splitNodeID := strings.Split(regNodeID, "<id>")
		nodeID := strings.Replace(splitNodeID[1], " ", "", -1)
		fmt.Println(nodeID)
		mux.Lock()
		nodeIDs = append(nodeIDs, nodeID)
		mux.Unlock()
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}
	fmt.Println(nodeIDs)

	// delete auto generated genesis file and create custom genesis file
	err = helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		_, err := client.DockerExec(node, fmt.Sprintf("rm /aion/custom/config/genesis.json"))
		if err != nil {
			return util.LogError(err)
		}		
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}
	err = createGenesisfile(aionconf, tn, addresses)
	if err != nil {
		return util.LogError(err)
	}

	// delete auto generated config gile and add custom config file 
	err = helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		mux.Lock()
		_, err := client.DockerExec(node, fmt.Sprintf("rm /aion/custom/config/config.xml"))
		if err != nil {
			return util.LogError(err)
		}
		mux.Unlock()

		for _, wallet := range addresses {
			conf, err := buildConfig(aionconf, tn.LDD, wallet, nodeIDs, node.GetIP(), node.GetAbsoluteNumber())
			if err != nil {
				return util.LogError(err)
			}
			err = helpers.SingleCp(client, tn.BuildState, node, []byte(conf), "/aion/custom/config/config.xml")
			if err != nil {
				return util.LogError(err)
			}
		}
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}

	err = helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		_, err := client.DockerExecdit(node, fmt.Sprintf("bash -ic '/aion/aion.sh -n custom 2>&1 | tee %s'", conf.DockerOutputFile))
		if err != nil {
			return util.LogError(err)
		}
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}

	return nil
}

// Add handles adding a node to the Aion testnet
// TODO
func add(tn *testnet.TestNet) error {
	return nil
}

// create network configuration files
// ************************************************************************************

func createGenesisfile(aionconf *AionConf, tn *testnet.TestNet, accounts []string) error {
	alloc := map[string]map[string]string{}
	for _, account := range accounts {
		alloc[account] = map[string]string{
			"balance": aionconf.InitBalance,
		}
	}

	genesis := map[string]interface{}{
		"initBalance": aionconf.InitBalance,
		"energyLimit": aionconf.EnergyLimit,
		"nonce": aionconf.Nonce,
		"difficulty": aionconf.Difficulty,
		"timeStamp": aionconf.TimeStamp,
		"chainId": aionconf.ChainID,
	}

	genesis["alloc"] = alloc
	tn.BuildState.Set("alloc", alloc)
	tn.BuildState.Set("aionconf", aionconf)

	return helpers.CreateConfigsNewNodes(tn, "/aion/custom/config/genesis.json", func(node ssh.Node) ([]byte, error) {
		template, err := helpers.GetBlockchainConfig(blockchain, node.GetAbsoluteNumber(), "genesis.json.mustache", tn.LDD)
		if err != nil {
			return nil, util.LogError(err)
		}

		data, err := mustache.Render(string(template), util.ConvertToStringMap(genesis))
		if err != nil {
			return nil, util.LogError(err)
		}
		return []byte(data), nil
	})
}

func buildConfig(aionconf *AionConf, details *db.DeploymentDetails, wallet string, nodeIDs []string, nodeIP string, node int) (string, error) {

	dat, err := helpers.GetBlockchainConfig("aion", node, "config.xml.mustache", details)
	if err != nil {
		return "", util.LogError(err)
	}
	var tmp map[string]interface{}

	raw, err := json.Marshal(*aionconf)
	if err != nil {
		return "", util.LogError(err)
	}

	err = json.Unmarshal(raw, &tmp)
	if err != nil {
		return "", util.LogError(err)
	}

	mp := util.ConvertToStringMap(tmp)

	var p2pNodes string
	for i := range nodeIDs {
		p2pNodes += fmt.Sprintf("<node>p2p://%s@%s:30303</node>\n",nodeIDs[i],nodeIP)
	}

	mp["peerID"] = nodeIDs[node]
	mp["corsEnabled"] = fmt.Sprintf("%v",aionconf.CorsEnabled)
	mp["secureConnect"] = fmt.Sprintf("%v",aionconf.SecureConnect)
	mp["nrgDefault"] = fmt.Sprintf("%d",aionconf.NRGDefault)
	mp["nrgMax"] = fmt.Sprintf("%d",aionconf.NRGMax)
	mp["oracleEnabled"] = fmt.Sprintf("%v",aionconf.OracleEnabled)
	mp["nodes"] = p2pNodes
	mp["blocksQueueMax"] = fmt.Sprintf("%v",aionconf.BlocksQueueMax)
	mp["showStatus"] = fmt.Sprintf("%v",aionconf.ShowStatus)
	mp["showStatistics"] = fmt.Sprintf("%v",aionconf.ShowStatistics)
	mp["compactEnabled"] = fmt.Sprintf("%v",aionconf.CompactEnabled)
	mp["slowImport"] = fmt.Sprintf("%d",aionconf.SlowImport)
	mp["frequency"] = fmt.Sprintf("%d",aionconf.Frequency)
	mp["mining"] = fmt.Sprintf("%v",aionconf.Mining)
	mp["minerAddress"] = wallet
	mp["mineThreads"] = fmt.Sprintf("%d",aionconf.MineThreads)
	mp["extraData"] = aionconf.ExtraData
	mp["clampedDecayUB"] = fmt.Sprintf("%d",aionconf.ClampedDecayUB)
	mp["clampedDecayLB"] = fmt.Sprintf("%d",aionconf.ClampedDecayLB)
	mp["database"] = aionconf.Database
	mp["checkIntegrity"] = fmt.Sprintf("%v",aionconf.CheckIntegrity)
	mp["stateStorage"] = aionconf.StateStorage
	mp["vendor"] = aionconf.Vendor
	mp["dbCompression"] = fmt.Sprintf("%v",aionconf.DBCompression)
	mp["logFile"] = fmt.Sprintf("%v",aionconf.LogFile)
	mp["logPath"] = aionconf.LogPath
	mp["genLogs"] = aionconf.GenLogs
	mp["vmLogs"] = aionconf.VMLogs
	mp["apiLogs"] = aionconf.APILogs
	mp["syncLogs"] = aionconf.SyncLogs
	mp["dbLogs"] = aionconf.DBLogs
	mp["consLogs"] = aionconf.ConsLogs
	mp["p2pLogs"] = aionconf.P2PLogs
	mp["cacheMax"] = fmt.Sprintf("%d",aionconf.CacheMax)

	return mustache.Render(string(dat), mp)
}