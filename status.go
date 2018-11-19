package main

import (
	"strings"
	util "./util"
	db "./db"
	state "./state"
)

type TestNetStatus struct {
	Name		string	`json:"name"`
	Server		int		`json:"server"`
}

type BuildStatus struct {
	Error		error	`json:"error"`
	Progress	float64	`json:"progress"`
}


func CheckTestNetStatus() ([]TestNetStatus, error) {
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


func CheckBuildStatus() BuildStatus {
	return BuildStatus{ Progress:state.BuildingProgress, Error:state.BuildError }
}