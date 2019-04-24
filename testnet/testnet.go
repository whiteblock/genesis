package testnet

import(
	db "../db"
	ssh "../ssh"
	state "../state"
)

/*
	Represents a testnet and some state on that testnet. Should also contain the details needed to 
	rebuild this testnet.
 */
type TestNet struct {
	TestNetID		string
	Servers 		[]db.Server
	Clients			[]*ssh.Client
	NewlyBuiltNodes []db.Node
	BuildState  	BuildState
	Details 		[]db.DeploymentDetails
}

func NewTestNet(details db.DeploymentDetails, buildId string) (*TestNet,error) {
	
}