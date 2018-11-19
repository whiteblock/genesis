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
	
	/**Management Functions**/
	router.HandleFunc("/status/nodes/",nodesStatus).Methods("GET")
	router.HandleFunc("/exec/{server}/{node}",dockerExec).Methods("POST")

	http.ListenAndServe(conf.Listen, router)
}

func getAllServerInfo(w http.ResponseWriter, r *http.Request) {
	servers := db.GetAllServers()
	json.NewEncoder(w).Encode(servers)
}

func addNewServer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var server db.Server
	err := json.NewDecoder(r.Body).Decode(&server)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	err = server.Validate()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	log.Println(fmt.Sprintf("Adding server: %+v",server))
	
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
	err = json.NewEncoder(w).Encode(server)
	if err != nil {
		log.Println(err.Error())
	}
}

func deleteServer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		w.Write([]byte(err.Error()))
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
		w.Write([]byte(err.Error()))
		return
	}
	err = server.Validate()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

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
	
	w.Write([]byte(GetNextTestNetId()))

	go AddTestNet(testnet)
	
	//log.Println("Created a test net successfully!");
}

func getTestNetInfo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	//log.Println("Received raw id \""+params["id"]+"\"")
	if err != nil {
		w.Write([]byte("Invalid id"))
		return
	}
	//log.Println(fmt.Sprintf("Attempting to find testnet with id %d",id))
	testNet, err := db.GetTestNet(id)
	if err != nil {
		//log.Println("Error:",err)
		w.Write([]byte("Does Not Exist"))
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
	w.Write([]byte("Currently not supported"))
}

func getTestNetNodes(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		w.Write([]byte("Invalid id"))
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


func addTestNetNode(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Currently not supported"))
}

func deleteTestNetNode(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Currently not supported"))
}

func dockerExec(w http.ResponseWriter, r *http.Request) {
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
	res,err := util.SshExecCheck(server.Addr,fmt.Sprintf("docker exec whiteblock-node%d %s",node,cmd))
	if err != nil {
		log.Println(err.Error())
		w.Write([]byte(fmt.Sprintf("%s %s",res,err.Error())))
		return
	}
	w.Write([]byte(res))
}

func nodesStatus(w http.ResponseWriter, r *http.Request) {
	status, err := CheckTestNetStatus()
	if err != nil {
		log.Println(err.Error())
		w.Write([]byte(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(status)
}