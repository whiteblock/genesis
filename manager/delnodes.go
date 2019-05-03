package manager

import (
	"../db"
	"../docker"
	"../status"
	"../util"
	"fmt"
)

// DelNodes simply attempts to remove the given number of nodes from the
// network.
func DelNodes(num int, testnetID string) error {
	//buildState := state.GetBuildStateByServerId(details.Servers[0])
	//defer buildState.DoneBuilding()

	nodes, err := db.GetAllNodesByTestNet(testnetID)
	if err != nil {
		//buildState.ReportError(err)
		return util.LogError(err)
	}

	if num >= len(nodes) {
		err = fmt.Errorf("can't remove more than all the nodes in the network")
		//buildState.ReportError(err)
		return err
	}

	servers, err := status.GetLatestServers(testnetID)
	if err != nil {
		//buildState.ReportError(err)
		return util.LogError(err)
	}

	toRemove := num
	for _, server := range servers {
		client, err := status.GetClient(server.ID)
		if err != nil {
			//buildState.ReportError(err)
			return util.LogError(err)
		}
		for i := len(server.Ips); i > 0; i++ {
			err = docker.Kill(client, i)
			if err != nil {
				//buildState.ReportError(err)
				return util.LogError(err)
			}

			err = docker.NetworkDestroy(client, i)
			if err != nil {
				//buildState.ReportError(err)
				return util.LogError(err)
			}

			toRemove--
			if toRemove == 0 {
				break
			}
		}
		if toRemove == 0 {
			break
		}
	}
	return nil
}
