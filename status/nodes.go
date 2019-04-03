package status

import (
    "log"
    "strings"
    "fmt"
    "strconv"
    "sync"
    "context"
    "golang.org/x/sync/semaphore"
    util "../util"
    db "../db"
)

var conf *util.Config

func init() {
    conf = util.GetConfig()
}

/*
    Represents the status of the node
 */
type NodeStatus struct {
    Name        string  `json:"name"`
    Server      int     `json:"server"`
    Ip          string  `json:"ip"`
    Up          bool    `json:"up"`
    Cpu         float64 `json:"cpu"`
}

/*
    Finds the index of a node by name and server id
 */
func FindNodeIndex(status []NodeStatus,name string,serverId int) int {
    for i,stat := range status {
        if stat.Name == name && serverId == stat.Server {
            return i
        }
    }
    return -1
}

/*
    Gets the cpu usage of a node
 */
func SumCpuUsage(c *util.SshClient,name string) (float64,error) {
    res,err := c.Run(fmt.Sprintf("docker exec %s ps aux --no-headers | awk '{print $3}'",name))
    if err != nil {
        log.Println(err)
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
            log.Println(err)
            return -1,err
        }
        out += parsed;
    }
    return out, nil
}

/*
    Checks the status of the nodes in the current testnet
 */
func CheckNodeStatus(nodes []db.Node) ([]NodeStatus, error) {

    serverIds := []int{}
    out := make([]NodeStatus,len(nodes))

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
        out[node.LocalId] = NodeStatus{
                                Name:fmt.Sprintf("%s%d",conf.NodePrefix,node.LocalId),
                                Ip:node.Ip,
                                Server:node.Server,
                                Up:false,
                                Cpu:-1,
                            }//local id to testnet

    }
    servers, err := db.GetServers(serverIds)
    if err != nil {
        log.Println(err)
        return nil, err
    }
    mux := sync.Mutex{}
    sem := semaphore.NewWeighted(conf.ThreadLimit)
    ctx := context.TODO()

    for _, server := range servers {
        client,err := util.NewSshClient(server.Addr,server.Id)
        defer client.Close()
        if err != nil {
            log.Println(err)
            return nil,err
        }
        res, err := client.Run(
            fmt.Sprintf("docker ps | egrep -o '%s[0-9]*' | sort",conf.NodePrefix))
        if err != nil {
            log.Println(err)
            return nil, err
        }
        names := strings.Split(res,"\n")
        for _,name := range names {
            if len(name) == 0 {
                continue
            }
            
            index := FindNodeIndex(out,name,server.Id)
            if index == -1 {
                log.Printf("name=\"%s\",server=%d\n",name,server.Id)
            }
            sem.Acquire(ctx,1)
            go func(client *util.SshClient,name string,index int){
                defer sem.Release(1)
                cpuUsage,err := SumCpuUsage(client,name)
                if err != nil {
                    log.Println(err)
                }
                mux.Lock()
                out[index].Up = true
                out[index].Cpu = cpuUsage
                mux.Unlock()
            }(client,name,index)
        }
    }
    err = sem.Acquire(ctx,conf.ThreadLimit)
    if err != nil {
        log.Println(err)
        return nil,err
    }

    sem.Release(conf.ThreadLimit)
    return out, nil
}

