package rest

import(
    "log"
    "net/http"
    "encoding/json"
    "strconv"
    "github.com/gorilla/mux"
    netem "../net"
    db "../db"
    util "../util"
    status "../status"
)


func handleNet(w http.ResponseWriter,r *http.Request){
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["server"])
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),400)
        return
    }
    var net_conf []netem.Netconf
    decoder := json.NewDecoder(r.Body)
    decoder.UseNumber()
    err = decoder.Decode(&net_conf)
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),400)
        return
    }

    servers, err := db.GetServers([]int{id})
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),404)
        return
    }
    server := servers[0]
    client,err := util.NewSshClient(server.Addr)
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),500)
        return
    }
    defer client.Close()
    //fmt.Printf("GIVEN %v\n",net_conf)
    err = netem.ApplyAll(client,net_conf,server.ServerID)
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),500)
        return
    }
    w.Write([]byte("Success"))
}

func handleNetAll(w http.ResponseWriter,r *http.Request){
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["server"])
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),400)
        return
    }

    var net_conf netem.Netconf
    decoder := json.NewDecoder(r.Body)
    decoder.UseNumber()

    err = decoder.Decode(&net_conf)
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),400)
        return
    }

    servers, err := db.GetServers([]int{id})
    if err != nil {
        log.Println(err.Error())
        http.Error(w,err.Error(),404)
    }
    server := servers[0]
    client,err := util.NewSshClient(server.Addr)
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),500)
    }
    defer client.Close()

    nodes,err := status.GetLatestTestnetNodes()
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),500)
        return
    }

    netem.RemoveAll(client,len(nodes))
    err = netem.ApplyToAll(client,net_conf,server.ServerID,len(nodes))
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),500)
    }
    w.Write([]byte("Success"))
}

func stopNet(w http.ResponseWriter,r *http.Request){
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["server"])
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),400)
        return
    }
    servers, err := db.GetServers([]int{id})
    if err != nil {
        log.Println(err.Error())
        http.Error(w,err.Error(),404)
        return
    }

    server := servers[0]
    client,err := util.NewSshClient(server.Addr)
    if err != nil {
        log.Println(err)
        http.Error(w,err.Error(),500)
        return
    }
    defer client.Close()

    nodes,err := status.GetLatestTestnetNodes()
    if err != nil {
        log.Println(err.Error())
        http.Error(w,err.Error(),500)
        return
    }

    netem.RemoveAll(client,len(nodes))
    
    w.Write([]byte("Success"))
}