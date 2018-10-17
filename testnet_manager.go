package main

import (
	"fmt"
	db "./db"
	deploy "./deploy"
	util "./util"
)

type DeploymentDetails struct {
	Servers    []int
	Blockchain string
	Nodes      int
	Image      string
}

func AddTestNet(details DeploymentDetails) error {
	servers, err := db.GetServers(details.Servers)
	if err != nil {
		return err
	}
	config := deploy.Config{Nodes: details.Nodes, Image: details.Image, Servers: details.Servers}

	newServerData := deploy.Build(&config, servers) //TODO: Restructure distribution of nodes over servers

	testNetId := db.InsertTestNet(db.TestNet{Id: -1, Blockchain: details.Blockchain, Nodes: details.Nodes, Image: details.Image})

	for _, server := range newServerData {
		db.UpdateServerNodes(server.Id,server.Nodes)
		for i, ip := range server.Ips {
			node := db.Node{Id: -1, TestNetId: testNetId, Server: server.Id, LocalId: i, Ip: ip} //TODO: Correct LocalId obtaining method
			db.InsertNode(node)
		}
	}
	return nil
}

func RebuildTestNet(id int) {

}

func RemoveTestNet(id int) {
	nodes := db.GetAllNodesByTestNet(id)
	for _, node := range nodes {
		server, _, _ := db.GetServer(node.Server)
		util.SshExec(server.Addr, fmt.Sprintf("~/local_deploy/deploy --kill=%d", node.LocalId))
	}

}

func ClearServers(servers []db.Server) {

}
