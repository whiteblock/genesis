package status

import(
    "log"
    db "../db"
)

func GetLastTestNetId() (int,error) {
    testNets,err := db.GetAllTestNets()
    if err != nil{
        log.Println(err)
        return 0,err
    }
    highestId := 0

    for _, testNet := range testNets {
        if testNet.Id > highestId {
            highestId = testNet.Id
        }
    }
    return highestId,nil
}

func GetLatestTestnet() (db.TestNet,error) {
    testnetId,err := GetLastTestNetId()
    if err != nil {
        log.Println(err)
        return db.TestNet{},err
    }
    return db.GetTestNet(testnetId)
}

func GetLatestTestnetNodes() ([]db.Node,error){
    testnetId,err := GetLastTestNetId()
    if err != nil {
        log.Println(err)
        return nil,err
    }
    return db.GetAllNodesByTestNet(testnetId)
}


func GetLatestServers() ([]db.Server,error) {
    nodes,err := GetLatestTestnetNodes()
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

func GetLatestBuild() (db.DeploymentDetails,error) {
    testnetId,err := GetLastTestNetId()
    if err != nil {
        log.Println(err)
        return db.DeploymentDetails{},err
    }
    return db.GetBuildByTestnet(testnetId)
}