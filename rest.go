package main

import (
	"encoding/json"
    "log"
    "net/http"
    "github.com/gorilla/mux"
    db "./db"
  	"strconv"
)



func StartServer(){
	router := mux.NewRouter();

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

	router.HandleFunc("/testnet/{id}/nodes/",getTestNetNodes).Methods("GET");
	router.HandleFunc("/testnet/{id}/nodes/",addNodesToTestNet).Methods("POST")
	router.HandleFunc("/testnet/{id}/nodes/",removeNodesFromTestNet).Methods("DELETE")

	router.HandleFunc("/switch/",addSwitch).Methods("POST")
	router.HandleFunc("/switch/",getSwitches).Methods("GET")

	router.HandleFunc("/switch/{id}",getSwitch).Methods("GET")
	router.HandleFunc("/switch/{id}",updateSwitch).Methods("UPDATE")
	router.HandleFunc("/switch/{id}",deleteSwitch).Methods("DELETE")

	http.ListenAndServe(":8000", router)
}

func getAllServerInfo(w http.ResponseWriter, r *http.Request){
	servers := db.GetAllServers()
	json.NewEncoder(w).Encode(servers);

}

func addNewServer(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	var server Server
	_ = json.NewDecoder(r.Body).Decode(&server)
	id := InsertServer(params["name"],server)
	_, _ := w.Write(strconv.Itoa(id))
}

func getServerInfo(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	server := db.GetServer(params["id"])
	json.NewEncoder(w).Encode(server);
}

func deleteServer(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	db.DeleteServer(params["id"])
}

func updateServerInfo(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	var server Server
	_ = json.NewDecoder(r.Body).Decode(&server)
	db.UpdateServer(params["id"],server)
	w.Write("Success")
}


func getAllTestNets(w http.ResponseWriter, r *http.Request){

}

func createTestNet(w http.ResponseWriter, r *http.Request){

}

func getTestNetInfo(w http.ResponseWriter, r *http.Request){

}


func deleteTestNet(w http.ResponseWriter, r *http.Request){

}

func updateTestNet(w http.ResponseWriter, r *http.Request){

}

func getTestNetNodes(w http.ResponseWriter, r *http.Request){

}

func addNodesToTestNet(w http.ResponseWriter, r *http.Request){

}

func removeNodesFromTestNet(w http.ResponseWriter, r *http.Request){

}

func addSwitch(w http.ResponseWriter, r *http.Request){

}

func updateSwitch(w http.ResponseWriter, r *http.Request){

}

func deleteSwitch(w http.ResponseWriter, r *http.Request){
	
}