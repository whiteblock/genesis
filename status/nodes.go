package status

import (
    "log"
    "strings"
    "fmt"
    "strconv"
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

func SumCpuUsage(c *util.SshClient,name string) (float64,error) {
    res,err := c.Run(fmt.Sprintf("docker exec %s ps aux --no-headers | awk '{print $3}'",name))
    if err != nil {
        return -1,err
    }
    values := strings.Split(res,"\n")
    fmt.Printf("%#v\n",values)
    var out float64
    for _,value := range values {
        if len(value) == 0 {
            continue
        }
        parsed,err := strconv.ParseFloat(value, 64)
        if err != nil {
            return -1,err
        }
        out += parsed;
    }
    return out, nil
}

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
        client,err := util.NewSshClient(server.Addr)
        defer client.Close()
        if err != nil {
            return nil,err
        }
        res, err := client.Run("docker ps | egrep -o 'whiteblock-node[0-9]*' | sort")
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
            out[index].Cpu,err = SumCpuUsage(client,name)
            if err != nil {
                log.Println(err)
            }
        }
    }
    return out, nil
}

