package main

import (
	db "./db"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func StartServer() {
	router := mux.NewRouter()

	router.HandleFunc("/servers/", getAllServerInfo).Methods("GET")

	router.HandleFunc("/servers/{name}", addNewServer).Methods("PUT") //Private

	router.HandleFunc("/servers/{id}", getServerInfo).Methods("GET") 
	router.HandleFunc("/servers/{id}", deleteServer).Methods("DELETE") //Private
	router.HandleFunc("/servers/{id}", updateServerInfo).Methods("UPDATE") //Private

	router.HandleFunc("/testnets/", getAllTestNets).Methods("GET")
	router.HandleFunc("/testnets/", createTestNet).Methods("POST") //Create new test net

	router.HandleFunc("/switches/", getAllSwitchesInfo).Methods("GET")

	router.HandleFunc("/testnets/{id}", getTestNetInfo).Methods("GET")
	router.HandleFunc("/testnets/{id}", deleteTestNet).Methods("DELETE")

	router.HandleFunc("/testnets/{id}/nodes/", getTestNetNodes).Methods("GET")
	router.HandleFunc("/testnets/{id}/node/", addTestNetNode).Methods("POST")
	router.HandleFunc("/testnets/{id}/node/{nid}",deleteTestNetNode).Methods("DELETE")

	http.ListenAndServe(":8000", router)
}

func getAllServerInfo(w http.ResponseWriter, r *http.Request) {
	servers := db.GetAllServers()
	json.NewEncoder(w).Encode(servers)
}

func addNewServer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var server db.Server
	_ = json.NewDecoder(r.Body).Decode(&server)
	id := db.InsertServer(params["name"], server)
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
	json.NewEncoder(w).Encode(server)
}

func deleteServer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		json.NewEncoder(w).Encode(err)
		return
	}
	db.DeleteServer(id)
	w.Write([]byte("Success"))
}

func updateServerInfo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var server db.Server

	_ = json.NewDecoder(r.Body).Decode(&server)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		w.Write([]byte("Invalid id"))
		return
	}

	db.UpdateServer(id, server)
	w.Write([]byte("Success"))
}


func getAllSwitchesInfo(w http.ResponseWriter, r *http.Request) {
	switches := db.GetAllSwitches()
	json.NewEncoder(w).Encode(switches)
}



func getAllTestNets(w http.ResponseWriter, r *http.Request) {
	testNets := db.GetAllTestNets()
	json.NewEncoder(w).Encode(testNets)
}

func createTestNet(w http.ResponseWriter, r *http.Request) {
	//params := mux.Vars(r)
	var testnet DeploymentDetails
	_ = json.NewDecoder(r.Body).Decode(&testnet)
	
	err := AddTestNet(testnet)
	if(err != nil){
		json.NewEncoder(w).Encode(err)
		return
	}
	w.Write([]byte("Success"))

}

func getTestNetInfo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		w.Write([]byte("Invalid id"))
		return
	}
	testNet, err := db.GetTestNet(id)
	if err == nil {
		w.Write([]byte("Does Not Exist"))
	} else {
		json.NewEncoder(w).Encode(testNet)
	}

}

func deleteTestNet(w http.ResponseWriter, r *http.Request) {
	//params := mux.Vars(r)
	//TODO handle the deletion of the test net

}

func getTestNetNodes(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		w.Write([]byte("Invalid id"))
		return
	}
	nodes := db.GetAllNodesByTestNet(id)
	json.NewEncoder(w).Encode(nodes)
}


func addTestNetNode(w http.ResponseWriter, r *http.Request) {

}

func deleteTestNetNode(w http.ResponseWriter, r *http.Request) {

}