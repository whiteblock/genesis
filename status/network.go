package status

import(
    "log"
    "github.com/Whiteblock/go.uuid"
    db "../db"
)

/*
    GetNextTestNetId gets the next testnet id. Used for
    getting the id of a testnet that is in progress of being built
 */
func GetNextTestNetId() (string, error) {
    uid,err := uuid.NewV4()
    if err != nil {
        log.Println(err)
        return "",err
    }
    str := uid.String()
    return str,nil
}

/*
    Get the servers used in the latest testnet, populated with the 
    ips of all the nodes
 */
func GetLatestServers(testnetId string) ([]db.Server,error) {
    nodes,err := db.GetAllNodesByTestNet(testnetId)
    if err != nil {
        log.Println(err)
        return nil,err
    }
    serverIds := []int{}
    for _,node := range nodes {
        shouldAdd := true
        for _,id := range serverIds {
            if id == node.Server {
                shouldAdd = false
            }
        }
        if shouldAdd {
            serverIds = append(serverIds,node.Server)
        }
    }
    
    servers,err := db.GetServers(serverIds)
    if err != nil{
        log.Println(err)
        return nil,err
    }
    for _,node := range nodes {
        for i,_ := range servers {
            if servers[i].Ips == nil {
                servers[i].Ips = []string{}
            }
            if node.Server == servers[i].Id {
                servers[i].Ips = append(servers[i].Ips,node.Ip)
            }
            servers[i].Nodes++
        }
    }
    return servers,nil
}
