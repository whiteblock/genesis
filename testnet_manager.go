package main

import (
	"fmt"
	"log"
	"errors"
	"strings"
	db "./db"
	deploy "./deploy"
	util "./util"
	eos "./blockchains/eos"
	eth "./blockchains/ethereum"
)

type DeploymentDetails struct {
	Servers    []int
	Blockchain string
	Nodes      int
	Image      string
}

type TestNetStatus struct {
	Name		string	`json:"name"`
	Server		int		`json:"server"`
}

func AddTestNet(details DeploymentDetails) error {
	
	servers, err := db.GetServers(details.Servers)
	
	if err != nil {
		log.Println(err.Error())
		return err
	}
	fmt.Println("Got the Servers")
	
	config := deploy.Config{Nodes: details.Nodes, Image: details.Image, Servers: details.Servers}
	fmt.Printf("Created the build configuration : %+v \n",config)

	newServerData := deploy.Build(&config, servers) //TODO: Restructure distribution of nodes over servers
	fmt.Println("Built the docker containers")

	
	switch(details.Blockchain){
		case "eos":
			eos.Eos(details.Nodes,newServerData);
		case "ethereum":
			eth.Ethereum(4000000,15468,15468,details.Nodes,newServerData)
		default:
			return errors.New("Unknown blockchain")
	}

	testNetId := db.InsertTestNet(db.TestNet{Id: -1, Blockchain: details.Blockchain, Nodes: details.Nodes, Image: details.Image})

	for _, server := range newServerData {
		db.UpdateServerNodes(server.Id,0)
		for i, ip := range server.Ips {
			node := db.Node{Id: -1, TestNetId: testNetId, Server: server.Id, LocalId: i, Ip: ip} 
			db.InsertNode(node)
		}
	}
	return nil
}

func GetLastTestNetId() int {
	testNets := db.GetAllTestNets()
	highestId := -1

	for _, testNet := range testNets {
		if testNet.Id > highestId {
			highestId = testNet.Id
		}
	}
	return highestId
}

func GetNextTestNetId() string {
	highestId := GetLastTestNetId()
	return fmt.Sprintf("%d",highestId+1)
}

func RebuildTestNet(id int) {

}

func RemoveTestNet(id int) error {
	nodes, err := db.GetAllNodesByTestNet(id)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		server, _, err := db.GetServer(node.Server)
		if err != nil {
			return err
		}
		util.SshExec(server.Addr, fmt.Sprintf("~/local_deploy/deploy --kill=%d", node.LocalId))
	}
	return nil
}


func CheckTestNetStatus() ([]TestNetStatus, error){
	testnetId := GetLastTestNetId()
	nodes,err := db.GetAllNodesByTestNet(testnetId)

	if err != nil {
		return nil, err
	}

	serverIds := []int{}
	for _, node := range nodes {
		push := true
		for _, id := range serverIds {
			if id == node.Server {
				push = false
			}
		}
		if push {
			serverIds = append(serverIds,node.Server)
		}
	}
	servers, err := db.GetServers(serverIds)
	if err != nil {
		return nil, err
	}
	out := []TestNetStatus{}
	for _, server := range servers {
		res, err := util.SshExecCheck(server.Addr,"docker ps | egrep -o 'whiteblock-node[0-9]*' | sort")
		if err != nil {
			return nil, err
		}
		names := strings.Split(res,"\n")
		for _,name := range names {
			if len(name) == 0 {
				continue
			}
			status := TestNetStatus{Name:name,Server:server.Id}
			out = append(out,status)
		}
	}
	return out, nil
}