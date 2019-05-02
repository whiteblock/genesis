//Package geth handles the creation of the geth sidecar
package geth

import (
	"../../testnet"
	"../../blockchains/registrar"
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
	
	return nil
}

/***************************************************************************************************************************/

// Add handles adding a node to the geth testnet
// TODO
func Add(tn *testnet.TestNet) (error) {
	return nil
}
