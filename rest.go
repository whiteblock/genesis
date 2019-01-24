package main

import (
    "encoding/json"
    "github.com/gorilla/mux"
    "net/http"
    "strconv"
    "log"
    "fmt"
    "io/ioutil"
    util "./util"
    db "./db"
    state "./state"
    status "./status"
    netem "./net"
)


func StartServer() {
    router := mux.NewRouter()

    
    router.HandleFunc("/servers", getAllServerInfo).Methods("GET")
    router.HandleFunc("/servers/", getAllServerInfo).Methods("GET")

    router.HandleFunc("/servers/{name}", addNewServer).Methods("PUT") //Private

    router.HandleFunc("/servers/{id}", getServerInfo).Methods("GET") 
    router.HandleFunc("/servers/{id}", deleteServer).Methods("DELETE") //Private
    router.HandleFunc("/servers/{id}", updateServerInfo).Methods("UPDATE") //Private

    router.HandleFunc("/testnets/", getAllTestNets).Methods("GET")
    router.HandleFunc("/testnets", getAllTestNets).Methods("GET")

    router.HandleFunc("/testnets/", createTestNet).Methods("POST") //Create new test net
    router.HandleFunc("/testnets", createTestNet).Methods("POST") //Create new test net

    router.HandleFunc("/switches/", getAllSwitchesInfo).Methods("GET")

    router.HandleFunc("/testnets/{id}", getTestNetInfo).Methods("GET")
    router.HandleFunc("/testnets/{id}/", getTestNetInfo).Methods("GET")
    router.HandleFunc("/testnets/{id}", deleteTestNet).Methods("DELETE")
    router.HandleFunc("/testnets/{id}/", deleteTestNet).Methods("DELETE")

    router.HandleFunc("/testnets/{id}/nodes", getTestNetNodes).Methods("GET")
    router.HandleFunc("/testnets/{id}/nodes/", getTestNetNodes).Methods("GET")

    router.HandleFunc("/testnets/{id}/node/", addTestNetNode).Methods("POST")
    router.HandleFunc("/testnets/{id}/node/{nid}",deleteTestNetNode).Methods("DELETE")
    
    /**Management Functions**/
    router.HandleFunc("/status/nodes",nodesStatus).Methods("GET")
    router.HandleFunc("/status/nodes/",nodesStatus).Methods("GET")

    router.HandleFunc("/status/build",buildStatus).Methods("GET")
    router.HandleFunc("/status/build/",buildStatus).Methods("GET")

    router.HandleFunc("/exec/{server}/{node}",dockerExec).Methods("POST")

    router.HandleFunc("/params/{blockchain}",getBlockChainParams).Methods("GET")
    router.HandleFunc("/params/{blockchain}/",getBlockChainParams).Methods("GET")

    router.HandleFunc("/state/{blockchain}",getBlockChainState).Methods("GET")
    router.HandleFunc("/state/{blockchain}/",getBlockChainState).Methods("GET")

    router.HandleFunc("/defaults/{blockchain}",getBlockChainDefaults).Methods("GET")
    router.HandleFunc("/defaults/{blockchain}/",getBlockChainDefaults).Methods("GET")

    router.HandleFunc("/log/{server}/{node}",getBlockChainLog).Methods("GET")
    router.HandleFunc("/log/{server}/{node}/",getBlockChainLog).Methods("GET")

    router.HandleFunc("/nodes",getLastNodes).Methods("GET")
    router.HandleFunc("/nodes/",getLastNodes).Methods("GET")

    router.HandleFunc("/build",stopBuild).Methods("DELETE")
    router.HandleFunc("/build/",stopBuild).Methods("DELETE")

    router.HandleFunc("/emulate/{server}",stopNet).Methods("DELETE")
    router.HandleFunc("/emulate/{server}",handleNet).Methods("POST")
    router.HandleFunc("/emulate/all/{server}",handleNetAll).Methods("POST")

    http.ListenAndServe(conf.Listen, router)
}

func getAllServerInfo(w http.ResponseWriter, r *http.Request) {
    servers,err := db.GetAllServers()
    if err != nil {
        log.Println(err.Error())
        http.Error(w,err.Error(),204)
        return
    }
    json.NewEncoder(w).Encode(servers)
}

func addNewServer(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    var server db.Server
    err := json.NewDecoder(r.Body).Decode(&server)
    if err != nil {
        http.Error(w,err.Error(),400)
        return
    }
    err = server.Validate()
    if err != nil {
        http.Error(w,err.Error(),400)
        return
    }
    log.Println(fmt.Sprintf("Adding server: %+v",server))
    
    id,err := db.InsertServer(params["name"], server)
    if err != nil {
        http.Error(w,err.Error(),500)
        return
    }
    w.Write([]byte(strconv.Itoa(id)))
}

func getServerInfo(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)

    id, err := strconv.Atoi(params["id"])
    if err != nil {
        json.NewEncoder(w).Encode(err)
        return
    }
    server, _, err := db.GetServer(id)
    if err != nil {
        w.Write([]byte(err.Error()))
        return
    }
    err = json.NewEncoder(w).Encode(server)
    if err != nil {
        log.Println(err.Error())
    }
}

func deleteServer(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w,"Invalid id",400)
        return
    }
    db.DeleteServer(id)
    w.Write([]byte("Success"))
}

func updateServerInfo(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)

    var server db.Server

    err := json.NewDecoder(r.Body).Decode(&server)
    if err != nil {
        http.Error(w,err.Error(),400)
        return
    }
    err = server.Validate()
    if err != nil {
        http.Error(w,err.Error(),400)
        return
    }

    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w,"Invalid id",400)
        return
    }

    err = db.UpdateServer(id, server)
    if err != nil {
        http.Error(w,err.Error(),500)
        return
    }
    w.Write([]byte("Success"))
}


func getAllSwitchesInfo(w http.ResponseWriter, r *http.Request) {
    switches,err := db.GetAllSwitches()
    if err != nil{
        log.Println(err)
        http.Error(w,err.Error(),204)
        return
    }
    json.NewEncoder(w).Encode(switches)
}

func getAllTestNets(w http.ResponseWriter, r *http.Request) {
    testNets,err := db.GetAllTestNets()
    if err != nil{
        log.Println(err)
        http.Error(w,"There are no test nets",204)
        return
    }
    json.NewEncoder(w).Encode(testNets)
}

func createTestNet(w http.ResponseWriter, r *http.Request) {
    //params := mux.Vars(r)
    var testnet DeploymentDetails
    decoder := json.NewDecoder(r.Body)
    decoder.UseNumber()
    err := decoder.Decode(&testnet)
    if err != nil {
        log.Println(err.Error())
        http.Error(w,err.Error(),400)
        return
    }
    err = state.AcquireBuilding()
    if err != nil {
        log.Println(err.Error())
        http.Error(w,err.Error(),409)
        return
    }
    next,_ := GetNextTestNetId()
    w.Write([]byte(next))

    go AddTestNet(testnet)
    
    //log.Println("Created a test net successfully!");
}

func getTestNetInfo(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["id"])
    //log.Println("Received raw id \""+params["id"]+"\"")
    if err != nil {
        http.Error(w,"Invalid id",400)
        return
    }
    //log.Println(fmt.Sprintf("Attempting to find testnet with id %d",id))
    testNet, err := db.GetTestNet(id)
    if err != nil {
        http.Error(w,"Test net does not exist",404)
        return
    }
    err = json.NewEncoder(w).Encode(testNet)
    if err != nil {
        log.Println(err)
    }

}

func deleteTestNet(w http.ResponseWriter, r *http.Request) {
    //params := mux.Vars(r)
    //TODO handle the deletion of the test net
    http.Error(w,"Currently not supported",501)
}

func getTestNetNodes(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w,"Invalid id",400)
        return
    }
    nodes,err := db.GetAllNodesByTestNet(id)
    if err != nil {
        log.Println(err.Error())
        http.Error(w,err.Error(),204)
        return
    }
    json.NewEncoder(w).Encode(nodes)
}


func addTestNetNode(w http.ResponseWriter, r *http.Request) {
    http.Error(w,"Currently not supported",501)
}

func deleteTestNetNode(w http.ResponseWriter, r *http.Request) {
    http.Error(w,"Currently not supported",501)
}

func dockerExec(w http.ResponseWriter, r *http.Request) {
    if !conf.AllowExec {
        w.Write([]byte("This function is currently disabled"))
        return
    }
    
    params := mux.Vars(r)
    serverId, err := strconv.Atoi(params["server"])
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
        return
    }
    node,err := strconv.Atoi(params["node"])
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
        return
    }
    server,_,err := db.GetServer(serverId)
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
        return
    }
    cmd, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
        return
    }
    res,err := util.DockerExec(server.Addr,node,string(cmd))
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(fmt.Sprintf("%s %s",res,err.Error())))
        return
    }
    w.Write([]byte(res))
}

func nodesStatus(w http.ResponseWriter, r *http.Request) {
    out, err := status.CheckTestNetStatus()
    if err != nil {
        log.Println(err.Error())
        http.Error(w,err.Error(),500)
        return
    }
    json.NewEncoder(w).Encode(out)
}

func buildStatus(w http.ResponseWriter,r *http.Request){
    w.Write([]byte(status.CheckBuildStatus())) 
}

func getBlockChainParams(w http.ResponseWriter,r *http.Request){

    params := mux.Vars(r)
    log.Println("GET PARAMS : "+params["blockchain"])
    w.Write([]byte(GetParams(params["blockchain"])))
}

func getBlockChainState(w http.ResponseWriter,r *http.Request){
    params := mux.Vars(r)
    blockchain := params["blockchain"]
    switch blockchain {
        case "eos":
            data := state.GetEosState()
            if data == nil{
                w.Write([]byte("No state availible for eos"))
                return
            }
            json.NewEncoder(w).Encode(*data)
            return
    }
    w.Write([]byte("Unknown blockchain "+ blockchain))
}

func getBlockChainDefaults(w http.ResponseWriter,r *http.Request){
    params := mux.Vars(r)
    w.Write([]byte(GetDefaults(params["blockchain"])))
}

func getBlockChainLog(w http.ResponseWriter,r *http.Request){
    params := mux.Vars(r)
    serverId, err := strconv.Atoi(params["server"])
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
        return
    }
    node,err := strconv.Atoi(params["node"])
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
        return
    }
    server,_,err := db.GetServer(serverId)
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
        return
    }

    res,err := util.DockerRead(server.Addr,node,conf.DockerOutputFile)
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(fmt.Sprintf("%s %s",res,err.Error())))
        return
    }
    w.Write([]byte(res))
}

func getLastNodes(w http.ResponseWriter,r *http.Request) {
    id, err := GetLastTestNetId()
    if err != nil {
        log.Println(err)
        w.Write([]byte(err.Error()))
        return
    }
    nodes,err := db.GetAllNodesByTestNet(id)
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
        return
    }
    json.NewEncoder(w).Encode(nodes)
}

func stopBuild(w http.ResponseWriter,r *http.Request){
    err := state.SignalStop()
    if err != nil{
        w.Write([]byte(err.Error()))
        return
    }
    w.Write([]byte("Stop signal has been sent"))
}

func handleNet(w http.ResponseWriter,r *http.Request){
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["server"])

    var net_conf []netem.Netconf
    decoder := json.NewDecoder(r.Body)
    decoder.UseNumber()
    err = decoder.Decode(&net_conf)


    servers, err := db.GetServers([]int{id})
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
        return
    }
    server := servers[0]
    client,err := util.NewSshClient(server.Addr)
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
        return
    }
    defer client.Close()
    //fmt.Printf("GIVEN %v\n",net_conf)
    err = netem.ApplyAll(client,net_conf)
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
        return
    }
    w.Write([]byte("Success"))
}

func handleNetAll(w http.ResponseWriter,r *http.Request){
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["server"])

    var net_conf netem.Netconf
    decoder := json.NewDecoder(r.Body)
    decoder.UseNumber()
    err = decoder.Decode(&net_conf)


    servers, err := db.GetServers([]int{id})
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
    }
    server := servers[0]
    client,err := util.NewSshClient(server.Addr)
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
    }
    defer client.Close()

    id, err = GetLastTestNetId()
    if err != nil {
        log.Println(err)
        w.Write([]byte(err.Error()))
        return
    }

    nodes,err := db.GetAllNodesByTestNet(id)
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
        return
    }

    netem.RemoveAll(client,len(nodes))
    err = netem.ApplyToAll(client,net_conf,len(nodes))
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
    }
    w.Write([]byte("Success"))
}

func stopNet(w http.ResponseWriter,r *http.Request){
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["server"])

    servers, err := db.GetServers([]int{id})
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
    }
    server := servers[0]
    client,err := util.NewSshClient(server.Addr)
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
    }
    defer client.Close()

    id, err = GetLastTestNetId()
    if err != nil {
        log.Println(err)
        w.Write([]byte(err.Error()))
        return
    }

    nodes,err := db.GetAllNodesByTestNet(id)
    if err != nil {
        log.Println(err.Error())
        w.Write([]byte(err.Error()))
        return
    }

    netem.RemoveAll(client,len(nodes))
    
    w.Write([]byte("Success"))
}