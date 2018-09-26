package deploy

import (
	"strings"
	db "../db"
)


/**
 * Gets a list of servers based on the availible servers
 * and the command line arguments passed to it
 * @param  string	args	The servers separated by commas
 * @return []Server			The requested servers
 */
func GetServers(args string) []db.Server {
	requestedServers := strings.Split(args,",")
	allServers := db.getAllServers()
	servers := []db.Server{}
	for _, server := range requestedServers {
		servers = append(servers,allServers[server])
	}
	return servers
}

/**
 * Get information on a node based on its absolute index number
 * @param  []Server servers		The servers
 * @param  int		index		The node index          
 * @return string				The host server's IP address
 * @return string				The node's IP address
 * @return int					The node's relative number on the server
 */
func GetInfo(servers []db.Server, index int) (string,string,int){
	k := 0
	for i := 0; i < len(servers); i++ {
		for j := 0; j < len(servers[i].ips); j++ {
			if k == index {
				return servers[i].addr, servers[i].ips[j], j
			}
			k++
		}
	}
	panic("Index out of bounds")
	return "","",-1
}
