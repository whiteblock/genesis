//Package geth handles the creation of the geth sidecar
package geth

import (
	"../../testnet"
	"../../blockchains/registrar"
	"../../blockchains/helpers"
	"../../util"
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
	networkid int64
	accounts []*ethereum.Account
func startGeth(client *ssh.Client, panconf *panConf, accounts []*ethereum.Account, buildState *state.BuildState) error {
	serviceIps, err := util.GetServiceIps(GetServices())
	if err != nil {
		log.Println(err)
		return err
	}
	err = buildState.SetExt("signer_ip", serviceIps["geth"])
	if err != nil {
		log.Println(err)
		return err
	}
	err = buildState.SetExt("accounts", ethereum.ExtractAddresses(accounts))
	if err != nil {
		log.Println(err)
		return err
	}

	wg := &sync.WaitGroup{}

	for i, account := range accounts {
		wg.Add(1)
		go func(account *ethereum.Account, i int) {
			defer wg.Done()

			_, err := client.Run(fmt.Sprintf("docker exec wb_service0 bash -c 'echo \"%s\" >> /geth/pk%d'", account.HexPrivateKey(), i))
			if err != nil {
				log.Println(err)
				return
			}
			_, err = client.Run(
				fmt.Sprintf("docker exec wb_service0 geth "+
					"--datadir /geth/ --password /geth/passwd account import --password /geth/passwd /geth/pk%d", i))
			if err != nil {
				log.Println(err)
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
	_, err = client.Run(fmt.Sprintf(`docker exec -itd wb_service0 bash -ic 'geth --datadir /geth/ --rpc --rpcaddr 0.0.0.0`+
		` --rpcapi "admin,web3,db,eth,net,personal" --rpccorsdomain "0.0.0.0" --nodiscover --unlock="%s"`+
		` --password /geth/passwd --networkid %d console 2>&1 >> /output.log'`, unlock, panconf.NetworkID))
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
	return nil
}

/***************************************************************************************************************************/

// Add handles adding a node to the geth testnet
// TODO
func Add(tn *testnet.TestNet) (error) {
	return nil
}
