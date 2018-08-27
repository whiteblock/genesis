package main

import (
	"strings"
)
type Server struct{
	addr 		string //IP to access the server
	iaddr		Iface //Internal IP of the server for NIC attached to the vyos
	nodes 		int
	max 		int
	id 			int
	iface		string
	ips 		[]string
	switches 	[]Switch
}

type Iface struct {
	ip 			string
	gateway 	string
	subnet 		int
}

/**
 * Returns the availible servers and their capacity 
 */
func getAllServers() map[string]Server {
	allServers := make(map[string]Server)
	
	allServers["alpha"] =
		Server{	
			addr:"172.16.1.5",
			iaddr:Iface{ip:"10.254.1.100",gateway:"10.254.1.1",subnet:24},
			nodes:0,
			max:100,
			id:1,
			iface:"eno4",
			ips:[]string{},
			switches:[]Switch{ Switch{addr:"172.16.1.1",iface:"eth1",brand:VYOS} }}

	allServers["bravo"] =
		Server{	
			addr:"172.16.2.5",
			iaddr:Iface{ip:"10.254.2.100",gateway:"10.254.2.1",subnet:24},
			nodes:0,
			max:100,
			id:2,
			iface:"eno1",
			ips:[]string{},
			switches:[]Switch{ Switch{addr:"172.16.2.1",iface:"eth0",brand:HP} }}

	allServers["charlie"] = 
		Server{	
			addr:"172.16.3.5",
			iaddr:Iface{ip:"10.254.3.100",gateway:"10.254.3.1",subnet:24},
			nodes:0,
			max:30,
			id:3,
			iface:"eno3",
			ips:[]string{},
			switches:[]Switch{ Switch{addr:"172.16.1.1",iface:"eth3",brand:VYOS} }}

	allServers["delta"] = 
		Server{	
			addr:"172.16.4.5",
			iaddr:Iface{ip:"10.254.4.100",gateway:"10.254.4.1",subnet:24},
			nodes:0,
			max:32,
			id:4,
			iface:"eno3",
			ips:[]string{},
			switches:[]Switch{ Switch{addr:"172.16.1.1",iface:"eth4",brand:VYOS} }}

	allServers["delta"] = 
		Server{	
			addr:"172.16.4.5",
			iaddr:Iface{ip:"10.254.4.100",gateway:"10.254.4.1",subnet:24},
			nodes:0,
			max:32,
			id:4,
			iface:"eno3",
			ips:[]string{},
			switches:[]Switch{ Switch{addr:"172.16.1.1",iface:"eth4",brand:VYOS} }}

	allServers["ns2"] = 
		Server{
			addr:"172.16.8.8",
			iaddr:Iface{ip:"10.254.5.100",gateway:"10.254.5.1",subnet:24},
			nodes:0,
			max:10,
			id:5,
			iface:"eth0",
			ips:[]string{},
			switches:[]Switch{ Switch{addr:"172.16.5.1",iface:"eth0",brand:VYOS} }}

		
	
	return allServers
}

/**
 * Gets a list of servers based on the availible servers
 * and the command line arguments passed to it
 * @param  string	args	The servers separated by commas
 * @return []Server			The requested servers
 */
func getServers(args string) []Server {
	requestedServers := strings.Split(args,",")
	allServers := getAllServers()
	servers := []Server{}
	for _, server := range requestedServers {
		servers = append(servers,allServers[server])
	}
	return servers
}

/**
 * Get information on a node based on its absolute index number
 * @param  []Server servers       The servers
 * @param  int		index  		  The node index          
 * @return string				  The host server's IP address
 * @return string				  The node's IP address
 * @return int					  The node's relative number on the server
 */
func getInfo(servers []Server, index int) (string,string,int){
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
