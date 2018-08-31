package main

import(
	db "./db"

)

type DeploymentDetails struct{
	Servers		[]int
	Blockchain	string
	Nodes		int
	Image		string
}

//Currently only stores to the database
func AddTestNet(dd DeploymentDetails) error {
	servers,err := db.GetServers(dd.Servers)
	if err != nil {
		return err
	}
	config := Config{nodes:dd.Nodes, image:dd.Image, servers:dd.Servers}

	newServerData := build(&config,servers)//TODO: Restructure distribution of nodes over servers

	testNetId := db.InsertTestNet( db.TestNet{Id:-1, Blockchain:dd.Blockchain, Nodes:dd.Nodes, Image:db.Image} )

	for _,server := range newServerData {
		for i,ip := range server.Ips {
			node := db.Node(Id:-1, TestNetId:testNetId, Server:server.Id, LocalId: i,Ip:ip )//TODO: Correct LocalId obtaining method
			db.InsertNode(node)
		}
	}

	

	return nil
}


func AddTestNet(testnet db.TestNet){

}

func RebuildTestNet(){

}

func RemoveTestNet(...){

}

func ClearServers(servers []Servers){

}
