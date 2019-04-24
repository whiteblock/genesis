package deploy

import (
	helpers "../blockchains/helpers"
	db "../db"
	netem "../net"
	ssh "../ssh"
	testnet "../testnet"
)

/*
PurgeTestNetwork goes into each given ssh client and removes all the nodes and the networks.
Increments the build state len(clients) * 2 times and sets it stag to tearing down network,
if buildState is non nil.
*/
func PurgeTestNetwork(tn *testnet.TestNet) {
	if tn.BuildState != nil {
		tn.BuildState.SetBuildStage("Tearing down the previous testnet")
	}
	DockerStopServices(tn)
	helpers.AllServerExecCon(tn, func(client *ssh.Client, server *db.Server) error {
		DockerKillAll(client)
		if tn.BuildState != nil {
			tn.BuildState.IncrementDeployProgress()
		}
		DockerNetworkDestroyAll(client)
		if tn.BuildState != nil {
			tn.BuildState.IncrementDeployProgress()
		}
		netem.RemoveAllOnServer(client, server.Nodes)

		return nil
	})
}

func Destroy(buildConf *db.DeploymentDetails, tn *testnet.TestNet) error {
	DockerStopServices(tn)
	return helpers.AllServerExecCon(tn, func(client *ssh.Client, _ *db.Server) error {
		DockerKillAll(client)
		DockerNetworkDestroyAll(client)
		return nil
	})
}
