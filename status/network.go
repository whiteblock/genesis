package status

import(
    "log"
    "github.com/satori/go.uuid"
    db "../db"
)


/*
    GetNextTestNetId gets the next testnet id. Used for
    getting the id of a testnet that is in progress of being built
 */
func GetNextTestNetId() (string, error) {
    uid,err := uuid.NewV4()
    str := uid.String()//strings.Replace(uid.String(),"-","_",-1)
    return str,err
}


/*
    Get the id of the latest testnet
 */
func GetLastTestNetId() (string,error) {
    testNets,err := db.GetAllTestNets()
    if err != nil{
        log.Println(err)
        return "",err
    }
    var highestTS int64 = 0
    highestId := ""

    for _, testNet := range testNets {
        if testNet.Ts > highestTS {
            highestId = testNet.Id
            highestTS = testNet.Ts
        }
    }
    return highestId,nil
}

/*
    Get the latest testnet
 */
func GetLatestTestnet() (db.TestNet,error) {
    testnetId,err := GetLastTestNetId()
    if err != nil {
        log.Println(err)
        return db.TestNet{},err
    }
    return db.GetTestNet(testnetId)
}

/*
    Get all of the nodes in the latest testnet
 */
func GetLatestTestnetNodes() ([]db.Node,error){
    testnetId,err := GetLastTestNetId()
    if err != nil {
        log.Println(err)
        return nil,err
    }
    return db.GetAllNodesByTestNet(testnetId)
}

/*
    Get the servers used in the latest testnet, populated with the 
    ips of all the nodes
 */
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

/*
    Get the last successful build parameters
 */
func GetLatestBuild() (db.DeploymentDetails,error) {
    testnetId,err := GetLastTestNetId()
    if err != nil {
        log.Println(err)
        return db.DeploymentDetails{},err
    }
    return db.GetBuildByTestnet(testnetId)
}