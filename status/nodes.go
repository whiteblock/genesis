package status

import (
    "log"
    "strings"
    "fmt"
    util "../util"
    db "../db"
)

type TestNetStatus struct {
    Name        string  `json:"name"`
    Server      int     `json:"server"`
    Up          bool    `json:"up"`
    Cpu         float64 `json:"cpu"`
}

func FindNodeIndex(status []TestNetStatus,name string,serverId int) int {
    for i,stat := range status {
        if stat.Name == name && serverId == stat.Server {
            return i
        }
    }
    return -1
}

func GetLastTestNetId() (int,error) {
    testNets,err := db.GetAllTestNets()
    if err != nil{
        log.Println(err)
        return 0,err
    }
    highestId := -1

    for _, testNet := range testNets {
        if testNet.Id > highestId {
            highestId = testNet.Id
        }
    }
    return highestId,nil
}

//func SumCpuUsage()

func CheckTestNetStatus() ([]TestNetStatus, error) {
    testnetId,err := GetLastTestNetId()
    if err != nil {
        return nil,err
    }
    nodes,err := db.GetAllNodesByTestNet(testnetId)

    if err != nil {
        return nil, err
    }

    serverIds := []int{}
    out := []TestNetStatus{}

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
        initStatus := TestNetStatus{
                            Name:fmt.Sprintf("whiteblock-node%d",node.LocalId),
                            Server:node.Server,
                            Up:false,
                            Cpu:-1}
        out = append(out,initStatus)

    }
    servers, err := db.GetServers(serverIds)
    if err != nil {
        return nil, err
    }
    
    for _, server := range servers {
        res, err := util.SshExec(server.Addr,"docker ps | egrep -o 'whiteblock-node[0-9]*' | sort")
        if err != nil {
            return nil, err
        }
        names := strings.Split(res,"\n")
        for _,name := range names {
            if len(name) == 0 {
                continue
            }
            
            index := FindNodeIndex(out,name,server.Id)

            out[index].Up = true
        }
    }
    return out, nil
}

