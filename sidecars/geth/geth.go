//Package geth handles the creation of the geth sidecar
package geth

import (
	"../../testnet"
	"../../blockchains/registrar"
	"../../blockchains/helpers"
	"../../blockchains/ethereum"
	"../../util"
	"../../ssh"
	"../../db"
	"fmt"
	"log"
	"encoding/json"
	"sync"
	"github.com/Whiteblock/mustache"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()

	sidecar := "geth"
	registrar.RegisterSideCar(sidecar,registrar.SideCar{
		Image:"gcr.io/whiteblock/geth:dev",
	})
	registrar.RegisterBuildSideCar(sidecar, Build)
	registrar.RegisterAddSideCar(sidecar, Add)
}

// Build builds out a fresh new ethereum test network using geth
func Build(tn *testnet.TestNet) (error) {
	var networkID int64
	var accounts []*ethereum.Account
	var mine bool
	var peers []string
	var conf map[string]string 
	tn.BuildState.GetP("networkID",&networkID)
	tn.BuildState.GetP("accounts",&accounts)
	tn.BuildState.GetP("mine",&mine)
	tn.BuildState.GetP("peers",&peers)
	tn.BuildState.GetP("gethConf",&conf)


	err := helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server,  node ssh.Node) error {
		_,err := client.DockerExec(node,"mkdir -p /geth")
		return err
	})
	if err != nil {
		log.Println(err)
		return err
	}

	err = helpers.CreateConfigsSC(tn, "/geth/genesis.json",func(node ssh.Node) ([]byte, error) {
		gethConf, err := gethSpec(conf, ethereum.ExtractAddresses(accounts))
		return []byte(gethConf),err
	})
	if err != nil {
		log.Println(err)
		return err
	}

	/**Create the Password files**/
	{
		var data string
		for range accounts {
			data += "password\n"
		}
		err = helpers.CopyBytesToAllNodes(tn, data, "/geth/passwd")
		if err != nil {
			log.Println(err)
			return err
		}
	}

	
	out, err := json.Marshal(peers)
	if err != nil {
		log.Println(err)
		return err
	}

	//Copy static-nodes to every server
	err = helpers.CopyBytesToAllNodes(tn, string(out), "/geth/static-nodes.json")
	if err != nil {
		log.Println(err)
		return err
	}

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server,  node ssh.Node) error{
		_, err = client.DockerExec(node,
			fmt.Sprintf("geth --datadir /geth/ --networkid %d init /geth/genesis.json", networkID))
		if err != nil {
			log.Println(err)
			return err
		}
		wg := &sync.WaitGroup{}

		for i, account := range accounts {
			wg.Add(1)
			go func(account *ethereum.Account, i int) {
				defer wg.Done()

				_, err := client.DockerExec(node,fmt.Sprintf("bash -c 'echo \"%s\" >> /geth/pk%d'", account.HexPrivateKey(), i))
				if err != nil {
					tn.BuildState.ReportError(err)
					log.Println(err)
					return
				}
				_, err = client.DockerExec(node,
					fmt.Sprintf("geth --datadir /geth/ --password /geth/passwd account import --password /geth/passwd /geth/pk%d", i))
				if err != nil {
					log.Println(err)
					tn.BuildState.ReportError(err)
					return
				}

			}(account, i)
		}
		wg.Wait()

		unlock := ""
		for i, account := range accounts {
			if i != 0 {
				unlock += ","
			}
			unlock += account.HexAddress()
		}
		flags := ""

		if mine {
			flags = " --mine"
		}

		_, err := client.DockerExecdit(node,fmt.Sprintf(` bash -ic 'geth --datadir /geth/ --rpc --rpcaddr 0.0.0.0`+
			` --rpcapi "admin,web3,db,eth,net,personal" --rpccorsdomain "0.0.0.0"%s --nodiscover --unlock="%s"`+
			` --password /geth/passwd --networkid %d console 2>&1 >> /output.log'`, flags,unlock, networkID))
		return err
	})
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

/***************************************************************************************************************************/

// Add handles adding a node to the geth testnet
// TODO
func Add(tn *testnet.TestNet) (error) {
	return nil
}


func gethSpec(conf map[string]string, wallets []string) (string, error) {
	accounts := make(map[string]interface{})
	for _, wallet := range wallets {
		accounts[wallet] = map[string]interface{}{
			"balance": conf["initBalance"],
		}
	}

	tmp := map[string]interface{}{
		"chainId":        conf["networkID"],
		"difficulty":     conf["difficulty"],
		"gasLimit":       conf["gasLimit"],
		"homesteadBlock": 0,
		"eip155Block":    10,
		"eip158Block":    10,
		"alloc":          accounts,
	}
	filler := util.ConvertToStringMap(tmp)
	dat, err := helpers.GetStaticBlockchainConfig("geth", "genesis.json")
	if err != nil {
		return "", err
	}
	data, err := mustache.Render(string(dat), filler)
	return data, err
}