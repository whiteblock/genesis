package main

import (
	"encoding/json"
    "net/http"
    "github.com/gorilla/mux"
    db "./db"
  	"strconv"
)



func StartServer(){
	router := mux.NewRouter()

	router.HandleFunc("/servers/",getAllServerInfo).Methods("GET")

	router.HandleFunc("/servers/{name}",addNewServer).Methods("PUT")

	router.HandleFunc("/servers/{id}",getServerInfo).Methods("GET")
	router.HandleFunc("/servers/{id}",deleteServer).Methods("DELETE")
	router.HandleFunc("/servers/{id}",updateServerInfo).Methods("UPDATE")
	

	router.HandleFunc("/testnet/",getAllTestNets).Methods("GET")
	router.HandleFunc("/testnet/",createTestNet).Methods("POST")//Create new test net

	router.HandleFunc("/testnet/{id}",getTestNetInfo).Methods("GET")
	router.HandleFunc("/testnet/{id}",deleteTestNet).Methods("DELETE")
	router.HandleFunc("/testnet/{id}/",updateTestNet).Methods("UPDATE")

	router.HandleFunc("/testnet/{id}/nodes/",getTestNetNodes).Methods("GET")
	router.HandleFunc("/testnet/{id}/nodes/",addNodesToTestNet).Methods("POST")
	router.HandleFunc("/testnet/{id}/nodes/",removeNodesFromTestNet).Methods("DELETE")

	http.ListenAndServe(":8000", router)
}

func getAllServerInfo(w http.ResponseWriter, r *http.Request){
	servers := db.GetAllServers()
	json.NewEncoder(w).Encode(servers)
}

func addNewServer(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	var server db.Server
	_ = json.NewDecoder(r.Body).Decode(&server)
	id := InsertServer(params["name"],server)
	w.Write(strconv.Itoa(id))
}

func getServerInfo(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	server,err := db.GetServer(params["id"])
	json.NewEncoder(w).Encode(server)
}

func deleteServer(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	db.DeleteServer(params["id"])
}

func updateServerInfo(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	var server db.Server
	_ = json.NewDecoder(r.Body).Decode(&server)
	db.UpdateServer(params["id"],server)
	w.Write("Success")
}


func getAllTestNets(w http.ResponseWriter, r *http.Request){
	testNets := GetAllTestNets()
	json.NewEncoder(w).Encode(testNets)
}

func createTestNet(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	var testnet db.TestNet
	_ = json.NewDecoder(r.Body).Decode(&testnet)
	//TODO handle the creation of a testnet
	
	//Handle the nodes, and everything...
	id := db.InsertTestNet(testnet)
	w.Write(strconv.Itoa(id))
}

func getTestNetInfo(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	testnet, err := db.GetTestNet(params["id"])
	if err == nil {
		w.Write("Does Not Exist")
	}else{
		json.NewEncoder(w).Encode(testNets)
	}

}

func deleteTestNet(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	//TODO handle the deletion of the test net
	db.DeleteTestNet(params["id"])
}	

func updateTestNet(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	var testnet db.TestNet
	_ = json.NewDecoder(r.Body).Decode(&testnet)
	//TODO handle the update of a testnet
	
	//Handle the nodes, and everything...
	id := db.UpdateTestNet(params["id"],testnet)
	w.Write(strconv.Itoa(id))
}

func getTestNetNodes(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	nodes := db.GetAllNodesByTest(params["id"])
	json.NewEncoder(w).Encode(nodes)
}

func addNodesToTestNet(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
}

func removeNodesFromTestNet(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
}
