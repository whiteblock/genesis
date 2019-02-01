/*
    Implements the REST interface which is used to communicate with this module
 */
package rest

import (
    "encoding/json"
    "github.com/gorilla/mux"
    "net/http"
    "strconv"
    "log"
    "fmt"
    util "../util"
    db "../db"
    state "../state"
    status "../status"
    testnet "../testnet"
)

var conf *util.Config

func init() {
    conf = util.GetConfig()
}

/*
    Starts the rest server, blocking the calling thread from returning
 */
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
    
    /**Management Functions**/
    router.HandleFunc("/status/nodes",nodesStatus).Methods("GET")
    router.HandleFunc("/status/nodes/",nodesStatus).Methods("GET")

    router.HandleFunc("/status/build",buildStatus).Methods("GET")
    router.HandleFunc("/status/build/",buildStatus).Methods("GET")

    router.HandleFunc("/status/servers",getLatestServers).Methods("GET")
    router.HandleFunc("/status/servers/",getLatestServers).Methods("GET")

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

    router.HandleFunc("/nodes/{num}",addNodes).Methods("POST")
    router.HandleFunc("/nodes/{num}/",addNodes).Methods("POST")

    router.HandleFunc("/nodes/{num}",delNodes).Methods("DELETE")
    router.HandleFunc("/nodes/{num}/",delNodes).Methods("DELETE")

    router.HandleFunc("/build",stopBuild).Methods("DELETE")
    router.HandleFunc("/build/",stopBuild).Methods("DELETE")

    router.HandleFunc("/build",getAllBuilds).Methods("GET")
    router.HandleFunc("/build/",getAllBuilds).Methods("GET")

    router.HandleFunc("/build/{id}",getBuild).Methods("GET")
    router.HandleFunc("/build/{id}/",getBuild).Methods("GET")

    router.HandleFunc("/emulate/{server}",stopNet).Methods("DELETE")
    router.HandleFunc("/emulate/{server}",handleNet).Methods("POST")
    router.HandleFunc("/emulate/all/{server}",handleNetAll).Methods("POST")

    http.ListenAndServe(conf.Listen, router)
}

func nodesStatus(w http.ResponseWriter, r *http.Request) {
    out, err := status.CheckNodeStatus()
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
    w.Write([]byte(testnet.GetParams(params["blockchain"])))
}

func getBlockChainState(w http.ResponseWriter,r *http.Request){
    params := mux.Vars(r)
    blockchain := params["blockchain"]
    switch blockchain {
        case "eos":
            data := state.GetEosState()
            if data == nil{
                http.Error(w,"No state availible for eos",410)
                return
            }
            json.NewEncoder(w).Encode(*data)
            return
    }
    w.Write([]byte("Unknown blockchain "+ blockchain))
}

func getBlockChainDefaults(w http.ResponseWriter,r *http.Request){
    params := mux.Vars(r)
    w.Write([]byte(testnet.GetDefaults(params["blockchain"])))
}

func getBlockChainLog(w http.ResponseWriter,r *http.Request){
    params := mux.Vars(r)
    serverId, err := strconv.Atoi(params["server"])
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),400)
        return
    }
    node,err := strconv.Atoi(params["node"])
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),400)
        return
    }
  
    client,err := testnet.GetClient(serverId)
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),404)
        return
    }
    res,err := client.DockerRead(node,conf.DockerOutputFile)
    if err != nil {
        log.Println(err)
        http.Error(w,fmt.Sprintf("%s %s",res,err.Error()),500)
        return
    }
    w.Write([]byte(res))
}

func getLastNodes(w http.ResponseWriter,r *http.Request) {
    nodes,err := status.GetLatestTestnetNodes()
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),404)
        return
    }
    json.NewEncoder(w).Encode(nodes)
}

func stopBuild(w http.ResponseWriter,r *http.Request){
    err := state.SignalStop()
    if err != nil{
        http.Error(w,err.Error(),412)
        return
    }
    w.Write([]byte("Stop signal has been sent"))
}


func getLatestServers(w http.ResponseWriter, r *http.Request) {
    servers,err := status.GetLatestServers()
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),404)
        return
    }
    json.NewEncoder(w).Encode(servers)
}

func getAllBuilds(w http.ResponseWriter, r *http.Request) {
    builds,err := db.GetAllBuilds()
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),404)
        return
    }
    json.NewEncoder(w).Encode(builds)
}

func getBuild(w http.ResponseWriter,r *http.Request){
    params := mux.Vars(r)

    id, err := strconv.Atoi(params["id"])
    if err != nil {
        json.NewEncoder(w).Encode(err)
        return
    }
    build, err := db.GetBuildByTestnet(id)
    if err != nil {
        http.Error(w,err.Error(),404)
        return
    }
    err = json.NewEncoder(w).Encode(build)
    if err != nil {
        log.Println(err.Error())
    }
}