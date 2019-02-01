package rest

import(
    "log"
    "net/http"
    "encoding/json"
    "strconv"
    "github.com/gorilla/mux"
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
    err = state.AcquireBuilding()
    if err != nil {
        log.Println(err)
        http.Error(w,"There is a build already in progress",409)
        return
    }
    next,_ := testnet.GetNextTestNetId()
    w.Write([]byte(next))

    go testnet.AddTestNet(tn)
    
    //log.Println("Created a test net successfully!");
}

func getTestNetInfo(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id, err := strconv.Atoi(params["id"])
    if err != nil {
        http.Error(w,"Invalid id",400)
        return
    }
    //log.Println(fmt.Sprintf("Attempting to find tn with id %d",id))
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


func addNodes(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    num, err := strconv.Atoi(params["num"])
    if err != nil {
        http.Error(w,"Invalid id",400)
        return
    }
    var tn db.DeploymentDetails
    tn.Nodes = num
    decoder := json.NewDecoder(r.Body)
    decoder.UseNumber()
    err = decoder.Decode(&tn)
    if err != nil {
        log.Println(err)
        //http.Error(w,err.Error(),400)
        //return
    }
    err = state.AcquireBuilding()
    if err != nil {
        log.Println(err)
        http.Error(w,"There is a build in progress",409)
        return
    }
    w.Write([]byte("Adding the nodes"))
    go testnet.AddNodes(tn)
}

func delNodes(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    num, err := strconv.Atoi(params["num"])
    if err != nil {
        http.Error(w,"Invalid id",400)
        return
    }
    
    err = state.AcquireBuilding()
    if err != nil {
        log.Println(err)
        http.Error(w,"There is a build in progress",409)
        return
    }
    w.Write([]byte("Deleting the nodes"))
    go testnet.DelNodes(num)
}
