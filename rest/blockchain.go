package rest

import(
    "fmt"
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "github.com/gorilla/mux"
    "strings"
    "io/ioutil"
    util "../util"
    state "../state"
    status "../status"
    testnet "../testnet"
)

func getConfFiles(w http.ResponseWriter,r *http.Request) {
    params := mux.Vars(r)
    
    err := util.ValidateFilePath(params["blockchain"])
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),400)
        return
    }

    files,err := util.Lsr(fmt.Sprintf("./resources/"+params["blockchain"]))//TODO 
    if err != nil {
        log.Println(err)
        http.Error(w,fmt.Sprintf("Nothing availible for \"%s\"",params["blockchain"]),500)
        return
    }

    for i,file := range files {
        index := strings.LastIndex(file,"/")
        files[i] = file[index+1:]
    }

    json.NewEncoder(w).Encode(files)
}

func getConfFile(w http.ResponseWriter,r *http.Request) {
    params := mux.Vars(r)

    err := util.ValidateFilePath(params["blockchain"])
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),400)
        return
    }
    err = util.ValidateFilePath(params["file"])
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),400)
        return
    }
    if strings.Contains(params["blockchain"],"..") || strings.Contains(params["file"],"..") {
        http.Error(w,"relative path operators not allowed",401)
        return
    }
    if !strings.HasSuffix(params["file"],"mustache") && !strings.HasSuffix(params["file"],"json") {
        http.Error(w,"Cannot read non mustache/json files",403)
        return
    }
    path := "./resources/"+params["blockchain"]+"/"+params["file"]
    fmt.Println(path)
    data,err := ioutil.ReadFile(path)
    if err != nil{
        http.Error(w,"File not found",404)
        return
    }
    json.NewEncoder(w).Encode(string(data))
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
    lines := -1
    _,ok := params["lines"]
    if ok {
        lines,err = strconv.Atoi(params["lines"])
        if err != nil {
            log.Println(err)
            http.Error(w,err.Error(),400)
            return
        }
    }
  
    client,err := status.GetClient(serverId)
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),404)
        return
    }
    res,err := client.DockerRead(node,conf.DockerOutputFile,lines)
    if err != nil {
        log.Println(err)
        http.Error(w,fmt.Sprintf("%s %s",res,err.Error()),500)
        return
    }
    w.Write([]byte(res))
}