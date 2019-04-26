package status

import (
	"../db"
	"log"
)

// GetLatestServers gets the servers used in the latest testnet, populated with the
// ips of all the nodes
func GetLatestServers(testnetID string) ([]db.Server, error) {
	nodes, err := db.GetAllNodesByTestNet(testnetID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	serverIDs := db.GetUniqueServerIDs(nodes)

	servers, err := db.GetServers(serverIDs)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	for _, node := range nodes {
		for i := range servers {
			if servers[i].Ips == nil {
				servers[i].Ips = []string{}
			}
			if node.Server == servers[i].ID {
				servers[i].Ips = append(servers[i].Ips, node.IP)
			}
			servers[i].Nodes++
		}
	}
	return servers, nil
}
