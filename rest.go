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

	router.HandleFunc("/servers/{name}", addNewServer).Methods("PUT")

	router.HandleFunc("/servers/{id}", getServerInfo).Methods("GET")
	router.HandleFunc("/servers/{id}", deleteServer).Methods("DELETE")
	router.HandleFunc("/servers/{id}", updateServerInfo).Methods("UPDATE")

	router.HandleFunc("/testnet/", getAllTestNets).Methods("GET")
	router.HandleFunc("/testnet/", createTestNet).Methods("POST") //Create new test net

	router.HandleFunc("/testnet/{id}", getTestNetInfo).Methods("GET")
	router.HandleFunc("/testnet/{id}", deleteTestNet).Methods("DELETE")

	router.HandleFunc("/testnet/{id}/nodes/", getTestNetNodes).Methods("GET")

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
		w.Write([]byte("Invalid id"))
		return
	}

	server, _, err := db.GetServer(id)

	json.NewEncoder(w).Encode(server)
}

func deleteServer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		w.Write([]byte("Invalid id"))
		return
	}
	db.DeleteServer(id)
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

func getAllTestNets(w http.ResponseWriter, r *http.Request) {
	testNets := db.GetAllTestNets()
	json.NewEncoder(w).Encode(testNets)
}

func createTestNet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var testnet db.TestNet
	_ = json.NewDecoder(r.Body).Decode(&testnet)
	//TODO handle the creation of a testnet

	//Handle the nodes, and everything...

}

func getTestNetInfo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		w.Write([]byte("Invalid id"))
		return
	}
	testnet, err := db.GetTestNet(id)
	if err == nil {
		w.Write([]byte("Does Not Exist"))
	} else {
		json.NewEncoder(w).Encode(testNets)
	}

}

func deleteTestNet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
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
