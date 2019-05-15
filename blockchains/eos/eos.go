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

//Package eos handles eos specific functionality
package eos

import (
	"fmt"
	"github.com/whiteblock/genesis/blockchains/helpers"
	"github.com/whiteblock/genesis/blockchains/registrar"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"math/rand"
	"strings"
	"sync"
)

var conf *util.Config

const blockchain = "eos"

func init() {
	conf = util.GetConfig()
	registrar.RegisterBuild(blockchain, build)
	registrar.RegisterAddNodes(blockchain, Add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, GetDefaults)
	registrar.RegisterParams(blockchain, GetParams)
}

// build builds out a fresh new eos test network using geth
func build(tn *testnet.TestNet) error {
	clients := tn.GetFlatClients()
	if tn.LDD.Nodes < 2 {
		return fmt.Errorf("cannot build network with less than 2 nodes")
	}

	eosconf, err := newConf(tn.LDD.Params)
	if err != nil {
		return util.LogError(err)
	}
	if eosconf.BlockProducers < 2 {
		return fmt.Errorf("cannot build eos network with only one block producer")
	}
	eosconf.BlockProducers++

	wg := sync.WaitGroup{}
	mux := sync.Mutex{}

	masterIP := tn.Nodes[0].IP

	masterNode := tn.Nodes[0]
	masterClient := tn.Clients[masterNode.Server]

	clientPasswords := make(map[string]string)

	tn.BuildState.SetBuildSteps(17 + (tn.LDD.Nodes * (3)) + (int(eosconf.UserAccounts) * (2)) + ((int(eosconf.UserAccounts) / 50) * tn.LDD.Nodes))

	contractAccounts := []string{
		"eosio.bpay",
		"eosio.msig",
		"eosio.names",
		"eosio.ram",
		"eosio.ramfee",
		"eosio.saving",
		"eosio.stake",
		"eosio.token",
		"eosio.vpay",
	}
	km, err := helpers.NewKeyMaster(tn.LDD, blockchain)
	if err != nil {
		return util.LogError(err)
	}
	km.AddGenerator(func(client *ssh.Client) (util.KeyPair, error) {
		data, err := client.DockerExec(masterNode, "cleos create key --to-console | awk '{print $3}'")
		if err != nil {
			return util.KeyPair{}, err
		}
		keyPair := strings.Split(data, "\n")
		if len(data) < 10 {
			return util.KeyPair{}, fmt.Errorf("unexpected create key output %s", keyPair)
		}
		return util.KeyPair{PrivateKey: keyPair[0], PublicKey: keyPair[1]}, nil
	})

	keyPairs, err := km.GetServerKeyPairs(tn.Servers, clients)
	if err != nil {
		return util.LogError(err)
	}

	contractKeyPairs, err := km.GetMappedKeyPairs(contractAccounts, masterClient)
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.IncrementBuildProgress()

	masterKeyPair := keyPairs[masterIP]

	var accountNames []string
	for i := 0; i < int(eosconf.UserAccounts); i++ {
		accountNames = append(accountNames, eosGetregularname(i))
	}
	accountKeyPairs, err := km.GetMappedKeyPairs(accountNames, masterClient)
	if err != nil {
		return util.LogError(err)
	}
	//tn.BuildState.SetBuildStage("Starting up keos")

	/**Start keos and add all the key pairs for all the nodes**/
	tn.BuildState.SetBuildStage("Generating key pairs")

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, node ssh.Node) error {
		/**Start keosd**/
		_, err = client.DockerExecd(node, "keosd --http-server-address 0.0.0.0:8900")
		if err != nil {
			return util.LogError(err)
		}
		mux.Lock()
		clientPasswords[node.GetIP()], err = eosCreatewallet(client, node)
		mux.Unlock()
		if err != nil {
			return util.LogError(err)
		}

		cmds := []string{}
		for _, name := range accountNames {
			if len(cmds) > 50 {
				_, err := client.KTDockerMultiExec(node, cmds)
				if err != nil {
					return util.LogError(err)
				}
				tn.BuildState.IncrementBuildProgress()
				cmds = []string{}
			}

			cmds = append(cmds, fmt.Sprintf("cleos wallet import --private-key %s", accountKeyPairs[name].PrivateKey))
		}

		if len(cmds) > 0 {
			_, err := client.KTDockerMultiExec(node, cmds)
			if err != nil {
				return util.LogError(err)
			}
		}
		tn.BuildState.IncrementBuildProgress()
		return nil
	})

	password := clientPasswords[masterIP]
	passwordNormal := clientPasswords[tn.Nodes[1].IP]
	tn.BuildState.IncrementBuildProgress()

	tn.BuildState.SetBuildStage("Building genesis block")
	genesis, err := eosconf.GenerateGenesis(keyPairs[masterIP].PublicKey, tn.LDD)
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.IncrementBuildProgress()
	err = helpers.MkdirAllNodes(tn, "/datadir/blocks")
	if err != nil {
		return util.LogError(err)
	}

	err = helpers.CopyBytesToAllNodes(tn,
		genesis, "/datadir/genesis.json",
		eosconf.GenerateConfig(), "/datadir/config.ini")
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.IncrementBuildProgress()
	/**Step 2d**/
	tn.BuildState.SetBuildStage("Starting EOS BIOS boot sequence")

	_, err = masterClient.KeepTryDockerExec(masterNode, fmt.Sprintf("cleos wallet import --private-key %s",
		keyPairs[masterIP].PrivateKey))
	if err != nil {
		return util.LogError(err)
	}

	err = masterClient.DockerExecdLog(masterNode,
		fmt.Sprintf(`nodeos -e -p eosio --genesis-json /datadir/genesis.json --config-dir /datadir --data-dir /datadir %s %s`,
			eosGetkeypairflag(keyPairs[masterIP]),
			eosGetptpflags(tn.Nodes, 0)))
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.IncrementBuildProgress()
	/**Step 3**/
	{
		masterClient.DockerExec(masterNode, fmt.Sprintf("cleos -u http://%s:8889 wallet unlock --password %s",
			masterIP, password)) //Can fail

		for _, account := range contractAccounts {
			wg.Add(1)
			go func(masterIP string, account string, masterKeyPair util.KeyPair, contractKeyPair util.KeyPair) {
				defer wg.Done()

				_, err := masterClient.KeepTryDockerExec(masterNode, fmt.Sprintf("cleos wallet import --private-key %s",
					contractKeyPair.PrivateKey))
				if err != nil {
					tn.BuildState.ReportError(err)
					return
				}
				_, err = masterClient.KeepTryDockerExec(masterNode, fmt.Sprintf("cleos -u http://%s:8889 create account eosio %s %s %s",
					masterIP, account, masterKeyPair.PublicKey, contractKeyPair.PublicKey))
				if err != nil {
					tn.BuildState.ReportError(err)
					return
				}

				//log.Println("Finished creating account for "+account)
			}(masterIP, account, masterKeyPair, contractKeyPairs[account])

		}
		wg.Wait()
		if !tn.BuildState.ErrorFree() {
			return tn.BuildState.GetError()
		}

	}
	tn.BuildState.IncrementBuildProgress()
	/**Steps 4 and 5**/
	{
		contracts := []string{"eosio.token", "eosio.msig"}
		masterClient.KeepTryDockerExec(masterNode, fmt.Sprintf("cleos -u http://%s:8889 wallet unlock --password %s", masterIP, password)) //ign

		for _, contract := range contracts {

			_, err = masterClient.KeepTryDockerExec(masterNode, fmt.Sprintf("cleos -u http://%s:8889 set contract %s /opt/eosio/contracts/%s",
				masterIP, contract, contract))
			if err != nil {
				return util.LogError(err)
			}
		}
	}
	tn.BuildState.SetBuildStage("Creating the tokens")
	tn.BuildState.IncrementBuildProgress()
	/**Step 6**/
	_, err = masterClient.KeepTryDockerExecAll(masterNode,
		fmt.Sprintf("cleos -u http://%s:8889 push action eosio.token create '[ \"eosio\", \"10000000000.0000 SYS\" ]' -p eosio.token@active",
			masterIP),
		fmt.Sprintf("cleos -u http://%s:8889 push action eosio.token issue '[ \"eosio\", \"1000000000.0000 SYS\", \"memo\" ]' -p eosio@active",
			masterIP))
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.SetBuildStage("Setting up the system contract")
	masterClient.DockerExec(masterNode, fmt.Sprintf("cleos -u http://%s:8889 wallet unlock --password %s", masterIP, password)) //Ignore fail

	tn.BuildState.IncrementBuildProgress()
	/**Step 7**/

	_, err = masterClient.KeepTryDockerExec(masterNode,
		fmt.Sprintf("cleos -u http://%s:8889 set contract -x 1000 eosio /opt/eosio/contracts/eosio.system", masterIP))

	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.IncrementBuildProgress()
	/**Step 8**/
	_, err = masterClient.KeepTryDockerExecAll(masterNode,
		fmt.Sprintf(`cleos -u http://%s:8889 push action eosio setpriv '["eosio.msig", 1]' -p eosio@active`,
			masterIP),
		fmt.Sprintf(`cleos -u http://%s:8889 push action eosio init '["0", "4,SYS"]' -p eosio@active`, masterIP))
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.SetBuildStage("Creating the block producers")
	tn.BuildState.IncrementBuildProgress()

	/**Step 10a**/

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		if node.GetAbsoluteNumber() == 0 || node.GetAbsoluteNumber() > int(eosconf.BlockProducers) {
			return nil
		}
		keyPair := keyPairs[node.GetIP()]

		_, err := masterClient.DockerExec(masterNode, fmt.Sprintf("cleos wallet import --private-key %s", keyPair.PrivateKey)) //ignore return
		if err != nil {
			return util.LogError(err)
		}
		_, err = masterClient.KeepTryDockerExecAll(masterNode,
			fmt.Sprintf(`cleos -u http://%s:8889 system newaccount eosio --transfer %s %s %s --stake-net "%d SYS" --stake-cpu "%d SYS" --buy-ram-kbytes %d`,
				masterIP,
				eosGetproducername(node.GetAbsoluteNumber()),
				masterKeyPair.PublicKey,
				keyPair.PublicKey,
				eosconf.BpNetStake,
				eosconf.BpCPUStake,
				eosconf.BpRAM),
			fmt.Sprintf(`cleos -u http://%s:8889 transfer eosio %s "%d SYS"`,
				masterIP,
				eosGetproducername(node.GetAbsoluteNumber()),
				eosconf.BpFunds))
		return err
	})
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Starting up the candidate block producers")
	/**Step 11c**/

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, node ssh.Node) error {
		if node.GetAbsoluteNumber() == 0 {
			return nil
		}
		prodFlags := ""

		if node.GetAbsoluteNumber() <= int(eosconf.BlockProducers) {
			prodFlags = " -p " + eosGetproducername(node.GetAbsoluteNumber()) + " "
		}

		return client.DockerExecdLog(node,
			fmt.Sprintf(`nodeos --genesis-json /datadir/genesis.json --config-dir /datadir --data-dir /datadir %s %s %s`,
				prodFlags,
				eosGetkeypairflag(keyPairs[node.GetIP()]),
				eosGetptpflags(tn.Nodes, node.GetAbsoluteNumber())))
	})
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.IncrementBuildProgress()
	/**Step 11a**/

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, node ssh.Node) error {
		if node.GetAbsoluteNumber() == 0 || node.GetAbsoluteNumber() > int(eosconf.BlockProducers) {
			return nil
		}
		if node.GetAbsoluteNumber()%5 == 0 {
			masterClient.DockerExec(masterNode, fmt.Sprintf("cleos -u http://%s:8889 wallet unlock --password %s",
				masterIP, password)) //ignore
		}

		_, err := masterClient.KeepTryDockerExec(masterNode,
			fmt.Sprintf("cleos --wallet-url http://%s:8900 -u http://%s:8889 system regproducer %s %s https://whiteblock.io/%s",
				masterIP,
				masterIP,
				eosGetproducername(node.GetAbsoluteNumber()),
				keyPairs[node.GetIP()].PublicKey,
				keyPairs[node.GetIP()].PublicKey))
		return err
	})
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.IncrementBuildProgress()
	/**Step 11b**/
	_, err = masterClient.DockerExec(masterNode, fmt.Sprintf("cleos -u http://%s:8889 system listproducers", masterIP))
	if err != nil {
		return util.LogError(err)
	}
	/**Create normal user accounts**/
	tn.BuildState.SetBuildStage("Creating funded accounts")
	for _, name := range accountNames {
		wg.Add(1)
		go func(name string, masterKeyPair util.KeyPair, accountKeyPair util.KeyPair) {
			defer wg.Done()
			_, err := masterClient.KeepTryDockerExec(masterNode,
				fmt.Sprintf(`cleos -u http://%s:8889 system newaccount eosio --transfer %s %s %s --stake-net "%d SYS" --stake-cpu "%d SYS" --buy-ram-kbytes %d`,
					masterIP,
					name,
					masterKeyPair.PublicKey,
					accountKeyPair.PublicKey,
					eosconf.AccountNetStake,
					eosconf.AccountCPUStake,
					eosconf.AccountRAM))
			if err != nil {
				tn.BuildState.ReportError(err)
				return
			}

			_, err = masterClient.KeepTryDockerExec(masterNode,
				fmt.Sprintf(`cleos -u http://%s:8889 transfer eosio %s "%d SYS"`,
					masterIP,
					name,
					eosconf.AccountFunds))
			if err != nil {
				tn.BuildState.ReportError(err)
				return
			}
			tn.BuildState.IncrementBuildProgress()

		}(name, masterKeyPair, accountKeyPairs[name])
	}
	wg.Wait()
	if !tn.BuildState.ErrorFree() {
		return tn.BuildState.GetError()
	}

	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Voting in block producers")
	/**Vote in block producers**/
	{
		node := len(tn.Nodes)

		if node > int(eosconf.BlockProducers) {
			node = int(eosconf.BlockProducers)
		}
		masterClient.DockerExec(tn.Nodes[1], fmt.Sprintf("cleos -u http://%s:8889 wallet unlock --password %s", //BUG: bad assumption
			masterIP, passwordNormal))
		n := 0
		for _, name := range accountNames {
			prod := 0
			fmt.Printf("name=%sn=%d\n", name, n)
			if n > 0 {
				prod = rand.Intn(100) % n
			}

			prod = (prod % (node - 1)) + 1
			wg.Add(1)
			go func(masterIP string, name string, prod int) {
				defer wg.Done()

				_, err := masterClient.KeepTryDockerExec(tn.Nodes[1], //BUG
					fmt.Sprintf("cleos -u http://%s:8889 system voteproducer prods %s %s",
						masterIP,
						name,
						eosGetproducername(prod)))
				if err != nil {
					tn.BuildState.ReportError(err)
					return
				}

				tn.BuildState.IncrementBuildProgress()
			}(masterIP, name, prod)
			n++
		}
		wg.Wait()
		if !tn.BuildState.ErrorFree() {
			return util.LogError(tn.BuildState.GetError())
		}
	}
	tn.BuildState.IncrementBuildProgress()
	tn.BuildState.SetBuildStage("Initializing EOSIO")
	/**Step 12**/
	masterClient.DockerExec(masterNode, fmt.Sprintf("cleos -u http://%s:8889 wallet unlock --password %s",
		masterIP, password))
	_, err = masterClient.KeepTryDockerExecAll(masterNode,
		fmt.Sprintf(
			`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio.prods", "permission": "active"}}]}}' -p eosio@owner`,
			masterIP),
		fmt.Sprintf(
			`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio.prods", "permission": "active"}}]}}' -p eosio@active`,
			masterIP),
		fmt.Sprintf(
			`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.bpay", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.bpay@owner`,
			masterIP),
		fmt.Sprintf(
			`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.msig", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.msig@owner`,
			masterIP),
		fmt.Sprintf(
			`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.msig", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.msig@active`,
			masterIP),
		fmt.Sprintf(
			`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.names", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.names@owner`,
			masterIP),
		fmt.Sprintf(
			`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.names", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.names@active`,
			masterIP),
		fmt.Sprintf(
			`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.ram", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.ram@owner`,
			masterIP),
		fmt.Sprintf(
			`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.ram", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.ram@active`,
			masterIP),
		fmt.Sprintf(
			`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.ramfee", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.ramfee@owner`,
			masterIP),
		fmt.Sprintf(
			`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.ramfee", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.ramfee@active`,
			masterIP),
	)
	if err != nil {
		return util.LogError(err)
	}

	passwords := []string{}
	for _, node := range tn.Nodes {
		passwords = append(passwords, clientPasswords[node.IP])
	}
	tn.BuildState.SetExt("passwords", passwords)
	tn.BuildState.SetExt("accounts", accountNames)
	tn.BuildState.SetExt("number_of_accounts", eosconf.UserAccounts)
	tn.BuildState.IncrementBuildProgress()
	return nil
}

// Add handles adding a node to the eos testnet
// TODO
func Add(tn *testnet.TestNet) error {
	return nil
}

func eosCreatewallet(client *ssh.Client, node ssh.Node) (string, error) {
	data, err := client.DockerExec(node, "cleos wallet create --to-console | tail -n 1")
	if err != nil {
		return "", err
	}
	//fmt.Printf("CREATE WALLET DATA %s\n",data)
	offset := 0
	for data[len(data)-(offset+1)] != '"' {
		offset++
	}
	offset++
	data = data[1 : len(data)-offset]
	fmt.Printf("CREATE WALLET DATA %s\n", data)
	return data, nil
}

func eosGetkeypairflag(keyPair util.KeyPair) string {
	return fmt.Sprintf("--signature-provider %s=KEY:%s", keyPair.PublicKey, keyPair.PrivateKey)
}

func eosGetproducername(num int) string {
	if num == 0 {
		return "eosio"
	}
	out := ""

	for i := num; i > 0; i = (i - (i % 4)) / 4 {
		place := i % 4
		place++
		out = fmt.Sprintf("%d%s", place, out) //I hate this
	}
	for i := len(out); i < 5; i++ {
		out = "x" + out
	}

	return "prod" + out
}

func eosGetregularname(num int) string {

	out := ""
	//num -= blockProducers

	for i := num; i > 0; i = (i - (i % 4)) / 4 {
		place := i % 4
		place++
		out = fmt.Sprintf("%d%s", place, out) //I hate this
	}
	for i := len(out); i < 8; i++ {
		out = "x" + out
	}

	return "user" + out
}

func eosGetptpflags(nodes []db.Node, exclude int) string {
	flags := ""
	for i, node := range nodes {
		if i == exclude {
			continue
		}
		flags += fmt.Sprintf("--p2p-peer-address %s:8999 ", node.IP)
	}
	return flags
}
