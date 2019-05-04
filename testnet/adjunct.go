package testnet

import (
	"../db"
	"../ssh"
	"../state"
)

// AdjunctTestNet represents a part of the network which contains
// a class of sidecars.
type Adjunct struct {
	// Testnet is a pointer to the master testnet
	Main       *TestNet
	Index      int
	Nodes      []ssh.Node
	BuildState *state.BuildState //ptr to the main one
	LDD        *db.DeploymentDetails
}
