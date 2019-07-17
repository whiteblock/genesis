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
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/registrar"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"github.com/whiteblock/mustache"
	"regexp"
	"strings"
	"sync"
)

var conf = util.GetConfig()

const (
	blockchain = "aion"
	password   = ""
)

type aionAcc struct {
	PrivateKey string
	PublicKey  string
	Address    string
}

func init() {
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

	tn.BuildState.SetBuildSteps(8 + (tn.LDD.Nodes) + (tn.LDD.Nodes * (tn.LDD.Nodes - 2)))

	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Distributing secrets")
	{
		/**Create the Password files**/
		var data string
		data += "\\n"
		/**Copy over the password file**/
		err = helpers.CopyBytesToAllNodes(tn, data, "/aion/passwd")
		if err != nil {
			return util.LogError(err)
		}
	}
	tn.BuildState.IncrementBuildProgress()

	var addresses = make([]string, tn.LDD.Nodes)
	var nodeIDs = make([]string, tn.LDD.Nodes)

	tn.BuildState.SetBuildStage("Creating the wallets")

	err = helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		addr, err := createAccount(client, node)
		if err != nil {
			return util.LogError(err)
		}
		mux.Lock()
		addresses[node.GetAbsoluteNumber()] = addr
		mux.Unlock()
		tn.BuildState.IncrementBuildProgress()
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}
	wg := sync.WaitGroup{}
	for i := 0; i < int(aionconf.ExtraAccounts); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			addr, err := createAccount(tn.Clients[tn.Nodes[0].Server], tn.Nodes[0])
			if err != nil {
				log.Error(err)
				return
			}
			log.Debug("created an extra account")
			mux.Lock()
			addresses = append(addresses, addr)
			mux.Unlock()
		}()
	}
	if aionconf.ExtraAccounts > 0 {
		wg.Wait()
	}

	accounts := []aionAcc{}

	err = helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		pubkeyOut, err := client.DockerExec(node, "bash -c '/aion/./aion.sh -a list -n custom'")
		if err != nil {
			return util.LogError(err)
		}
		pubk := regexp.MustCompile(`(?m)0x(.{64})`)
		publicKeys := pubk.FindAllString(pubkeyOut, -1)
		log.WithFields(log.Fields{"regex": pubk, "pubkey": publicKeys}).Trace("Extracted the public key")
		wg := sync.WaitGroup{}
		for _, publicKey := range publicKeys {
			wg.Add(1)
			go func(publicKey string) {
				defer wg.Done()
				privatekeyOut, err := client.DockerExec(node, fmt.Sprintf("bash -c 'echo -e $(cat /aion/passwd) | /aion/./aion.sh -a export %s -n custom'", publicKey))
				if err != nil {
					tn.BuildState.ReportError(err)
					return
				}
				privk := regexp.MustCompile(`(?m)0x(.{128})`)
				privateKey := privk.FindAllString(privatekeyOut, 1)[0]
				log.WithFields(log.Fields{"regex": privk, "privatekey": privateKey}).Trace("Extracted the private key")

				mux.Lock()
				accounts = append(accounts, aionAcc{
					PrivateKey: privateKey,
					PublicKey:  publicKey,
					Address:    publicKey,
				})
				mux.Unlock()
			}(publicKey)
		}
		wg.Wait()
		if !tn.BuildState.ErrorFree() {
			return tn.BuildState.GetError()
		}
		tn.BuildState.IncrementBuildProgress()
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}

	helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		wg := sync.WaitGroup{}
		for _, acc := range accounts {
			wg.Add(1)
			go func(account aionAcc) {
				defer wg.Done()
				client.DockerExec(node,
					fmt.Sprintf("bash -c 'echo -e $(cat /aion/passwd) | "+
						"/aion/./aion.sh -a import %s -n custom'", account.PrivateKey))
			}(acc)
		}
		wg.Wait()
		//Errors are ok
		return nil
	})

	log.WithFields(log.Fields{"accounts": accounts}).Trace("extracted accounts")
	tn.BuildState.Set("generatedAccs", accounts)
	tn.BuildState.IncrementBuildProgress()

	//get permanent node id from auto-generated config.xml
	err = helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		output, err := client.DockerRead(node, "/aion/custom/config/config.xml", -1)
		if err != nil {
			return util.LogError(err)
		}
		reNodeID := regexp.MustCompile(`(?m)<id>(.{36})`)
		regNodeID := reNodeID.FindAllString(output, 1)[0]
		splitNodeID := strings.Split(regNodeID, "<id>")
		nodeID := strings.Replace(splitNodeID[1], " ", "", -1)
		log.WithFields(log.Fields{"nodeID": nodeID}).Trace("extracted node id")
		mux.Lock()
		nodeIDs[node.GetAbsoluteNumber()] = nodeID
		mux.Unlock()
		return nil
	})
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.Set("nodeIDs", nodeIDs)
	tn.BuildState.IncrementBuildProgress()

	tn.BuildState.SetBuildStage("Creating the genesis block")
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
	tn.BuildState.IncrementBuildProgress()

	tn.BuildState.SetBuildStage("Creating the configuration file")
	// delete auto generated config gile and add custom config file
	err = helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		_, err := client.DockerExec(node, "rm /aion/custom/config/config.xml")
		return util.LogError(err)
	})
	if err != nil {
		return util.LogError(err)
	}
	helpers.CreateConfigs(tn, "/aion/custom/config/config.xml", func(node ssh.Node) ([]byte, error) {
		conf, err := buildConfig(aionconf, tn, addresses[node.GetAbsoluteNumber()], node)
		if err != nil {
			return nil, util.LogError(err)
		}
		tn.BuildState.IncrementBuildProgress()
		return []byte(conf), nil
	})
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.IncrementBuildProgress()

	tn.BuildState.SetBuildStage("Starting network")
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
	tn.BuildState.IncrementBuildProgress()

	tn.BuildState.SetExt("networkID", aionconf.ChainID)
	helpers.SetFunctionalityGroup(tn, "eth")
	tn.BuildState.SetExt("accounts", addresses)
	tn.BuildState.SetExt("port", 8545)
	tn.BuildState.SetExt("namespace", "eth")
	tn.BuildState.SetExt("password", password)

	for _, account := range accounts {
		tn.BuildState.SetExt(account.Address, map[string]string{
			"privateKey": account.PrivateKey,
			"publicKey":  account.PublicKey,
		})
	}
	unlockAllAccounts(tn, accounts)

	tn.BuildState.IncrementBuildProgress()

	return nil
}

// Add handles adding a node to the Aion testnet
// TODO
func add(tn *testnet.TestNet) error {
	return nil
}

// create network configuration files
// ************************************************************************************

func createGenesisfile(aionconf *AConf, tn *testnet.TestNet, accounts []string) error {
	alloc := map[string]map[string]string{}
	for _, account := range accounts {
		alloc[account] = map[string]string{
			"balance": aionconf.InitBalance,
		}
	}

	genesis := map[string]interface{}{
		"initBalance": aionconf.InitBalance,
		"energyLimit": aionconf.EnergyLimit,
		"nonce":       aionconf.Nonce,
		"difficulty":  aionconf.Difficulty,
		"timeStamp":   aionconf.TimeStamp,
		"chainId":     aionconf.ChainID,
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

//tn *testnet.TestNet
func buildConfig(aionconf *AConf, tn *testnet.TestNet, wallet string, node ssh.Node) (string, error) {

	dat, err := helpers.GetBlockchainConfig(blockchain, node.GetAbsoluteNumber(), "config.xml.mustache", tn.LDD)
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
	var nodeIDs []string
	tn.BuildState.GetP("nodeIDs", &nodeIDs)
	var p2pNodes string
	for _, nod := range tn.NewlyBuiltNodes {
		if nod.GetID() == node.GetID() {
			continue
		}
		p2pNodes += fmt.Sprintf("<node>p2p://%s@%s:30303</node>\n", nod.GetID(), nodeIDs[nod.GetAbsoluteNumber()])
	}

	mp["peerID"] = nodeIDs[node.GetAbsoluteNumber()]
	mp["corsEnabled"] = fmt.Sprintf("%v", aionconf.CorsEnabled)
	mp["secureConnect"] = fmt.Sprintf("%v", aionconf.SecureConnect)
	mp["nrgDefault"] = fmt.Sprintf("%d", aionconf.NRGDefault)
	mp["nrgMax"] = fmt.Sprintf("%d", aionconf.NRGMax)
	mp["oracleEnabled"] = fmt.Sprintf("%v", aionconf.OracleEnabled)
	mp["nodes"] = p2pNodes
	mp["ipAddr"] = node.GetIP()
	mp["blocksQueueMax"] = fmt.Sprintf("%v", aionconf.BlocksQueueMax)
	mp["showStatus"] = fmt.Sprintf("%v", aionconf.ShowStatus)
	mp["showStatistics"] = fmt.Sprintf("%v", aionconf.ShowStatistics)
	mp["compactEnabled"] = fmt.Sprintf("%v", aionconf.CompactEnabled)
	mp["slowImport"] = fmt.Sprintf("%d", aionconf.SlowImport)
	mp["frequency"] = fmt.Sprintf("%d", aionconf.Frequency)
	mp["mining"] = fmt.Sprintf("%v", aionconf.Mining)
	mp["minerAddress"] = wallet
	mp["mineThreads"] = fmt.Sprintf("%d", aionconf.MineThreads)
	mp["extraData"] = aionconf.ExtraData
	mp["clampedDecayUB"] = fmt.Sprintf("%d", aionconf.ClampedDecayUB)
	mp["clampedDecayLB"] = fmt.Sprintf("%d", aionconf.ClampedDecayLB)
	mp["database"] = aionconf.Database
	mp["checkIntegrity"] = fmt.Sprintf("%v", aionconf.CheckIntegrity)
	mp["stateStorage"] = aionconf.StateStorage
	mp["vendor"] = aionconf.Vendor
	mp["dbCompression"] = fmt.Sprintf("%v", aionconf.DBCompression)
	mp["logFile"] = fmt.Sprintf("%v", aionconf.LogFile)
	mp["logPath"] = aionconf.LogPath
	mp["genLogs"] = aionconf.GenLogs
	mp["vmLogs"] = aionconf.VMLogs
	mp["apiLogs"] = aionconf.APILogs
	mp["syncLogs"] = aionconf.SyncLogs
	mp["dbLogs"] = aionconf.DBLogs
	mp["consLogs"] = aionconf.ConsLogs
	mp["p2pLogs"] = aionconf.P2PLogs
	mp["cacheMax"] = fmt.Sprintf("%d", aionconf.CacheMax)

	return mustache.Render(string(dat), mp)
}

func createAccount(client ssh.Client, node ssh.Node) (string, error) {
	output, err := client.DockerExec(node, fmt.Sprintf("bash -c 'echo -e $(cat /aion/passwd) | /aion/./aion.sh ac -n custom'"))
	if err != nil {
		return "", util.LogError(err)
	}
	reAddr := regexp.MustCompile(`(?m)A new account has been created:(.{67})`)
	regAddr := reAddr.FindAllString(output, 1)[0]
	splitAddr := strings.Split(regAddr, "A new account has been created:")
	addr := strings.Replace(splitAddr[1], " ", "", -1)
	log.WithFields(log.Fields{"addr": addr}).Trace("A new account has been created:")
	return addr, nil
}

// works but need to wait for some time before it actually works. Need to figure out what the reason for the needed delay is
func unlockAllAccounts(tn *testnet.TestNet, accounts []aionAcc) error {
	return helpers.AllNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		wg := sync.WaitGroup{}
		for _, acc := range accounts {
			wg.Add(1)
			go func(account aionAcc) {
				defer wg.Done()
				for {
					_, err := client.Run(
						fmt.Sprintf(
							`curl -sS -X POST http://%s:8545 -H "Content-Type: application/json"  -d `+
								`'{ "method": "personal_unlockAccount", "params": ["%s","%s",0], "id": 3, "jsonrpc": "2.0" }'`,
							node.GetIP(), account.Address, password))
					//pass = !(strings.Contains(out, ":true"))
					if err == nil {
						break
					}
				}
			}(acc)
		}
		wg.Wait()
		return nil
	})
}
