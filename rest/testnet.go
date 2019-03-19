package rest

import(
    "log"
    "net/http"
    "encoding/json"
    "strconv"
    "github.com/gorilla/mux"
    status "../status"
    state "../state"
    db "../db"
    testnet "../testnet"
)


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
    var tn db.DeploymentDetails
    decoder := json.NewDecoder(r.Body)
    decoder.UseNumber()
    err := decoder.Decode(&tn)
    if err != nil {
        log.Println(err.Error())
        http.Error(w,err.Error(),400)
        return
    }
    id,err := status.GetNextTestNetId()
    if err != nil {
        log.Println(err)
        http.Error(w,"Error Generating a new UUID",500)
        return
    }

    err = state.AcquireBuilding(tn.Servers,id)
    if err != nil {
        log.Println(err)
        http.Error(w,"There is a build already in progress",409)
        return
    }
    
    go testnet.AddTestNet(tn,id)
    w.Write([]byte(id))

}

func getTestNetInfo(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)

    //log.Println(fmt.Sprintf("Attempting to find tn with id %d",id))
    testNet, err := db.GetTestNet(params["id"])
    if err != nil {
        log.Println(err)
        http.Error(w,"Test net does not exist",404)
        return
    }
    err = json.NewEncoder(w).Encode(testNet)
    if err != nil {
        log.Println(err)
    }
}

func deleteTestNet(w http.ResponseWriter, r *http.Request) {
    //TODO handle the deletion of the test net
    http.Error(w,"Currently not supported",501)
}

func getTestNetNodes(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)

    nodes,err := db.GetAllNodesByTestNet(params["id"])
    if err != nil {
        log.Println(err.Error())
        http.Error(w,err.Error(),404)
        return
    }
    json.NewEncoder(w).Encode(nodes)
}


func addNodes(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    num, err := strconv.Atoi(params["num"])
    if err != nil {
        log.Println(err)
        http.Error(w,"Invalid number of nodes",400)
        return
    }

    testnetId := params["id"]

    tn,err := db.GetBuildByTestnet(testnetId)
    if err != nil{
        log.Println(err)
        http.Error(w,"Could not find the given testnet id",400)
        return
    }

    tn.Nodes = num
    decoder := json.NewDecoder(r.Body)
    decoder.UseNumber()
    err = decoder.Decode(&tn)
    if err != nil {
        log.Println(err)
        //Ignore error and continue
    }
    err = state.AcquireBuilding(tn.Servers,testnetId)
    if err != nil {
        log.Println(err)
        http.Error(w,"There is a build in progress",409)
        return
    }
    w.Write([]byte("Adding the nodes"))
    go testnet.AddNodes(tn,testnetId)
}

func delNodes(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    num, err := strconv.Atoi(params["num"])
    if err != nil {
        log.Println(err)
        http.Error(w,"Invalid id",400)
        return
    }
    
    testnetId := params["id"]
    
    tn,err := db.GetBuildByTestnet(testnetId)
    if err != nil{
        log.Println(err)
        http.Error(w,"Could not find the given testnet id",400)
        return
    }

    err = state.AcquireBuilding(tn.Servers,testnetId)//TODO: THIS IS WRONG
    if err != nil {
        log.Println(err)
        http.Error(w,"There is a build in progress",409)
        return
    }
    w.Write([]byte("Deleting the nodes"))
    go testnet.DelNodes(num,testnetId)
}
