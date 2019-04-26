package manager

import (
	"../db"
	"../deploy"
	"../status"
	"fmt"
	"log"
)

/*
   DelNodes simply attempts to remove the given number of nodes from the
   network.
*/
func DelNodes(num int, testnetId string) error {
	//buildState := state.GetBuildStateByServerId(details.Servers[0])
	//defer buildState.DoneBuilding()

	nodes, err := db.GetAllNodesByTestNet(testnetId)
	if err != nil {
		log.Println(err)
		//buildState.ReportError(err)
		return err
	}

	if num >= len(nodes) {
		err = fmt.Errorf("can't remove more than all the nodes in the network")
		//buildState.ReportError(err)
		return err
	}

	servers, err := status.GetLatestServers(testnetId)
	if err != nil {
		log.Println(err)
		//buildState.ReportError(err)
		return err
	}

	toRemove := num
	for _, server := range servers {
		client, err := status.GetClient(server.Id)
		if err != nil {
			log.Println(err.Error())
			//buildState.ReportError(err)
			return err
		}
		for i := len(server.Ips); i > 0; i++ {
			err = deploy.DockerKill(client, i)
			if err != nil {
				log.Println(err.Error())
				//buildState.ReportError(err)
				return err
			}

			err = deploy.DockerNetworkDestroy(client, i)
			if err != nil {
				log.Println(err.Error())
				//buildState.ReportError(err)
				return err
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
