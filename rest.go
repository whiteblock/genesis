package main

import (
	"encoding/json"
    "log"
    "net/http"
    "github.com/gorilla/mux"
)



func StartServer(){
	router := mux.NewRouter();

	router.HandleFunc("/servers/",getAllServerInfo).Methods("GET")
	router.HandleFunc("/servers/",addNewServer).Methods("POST")

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


}

func GetPerson(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
   
}


allServers["bravo"] =
		Server{	
			addr:"172.16.2.5",
			iaddr:Iface{ip:"10.254.2.100",gateway:"10.254.2.1",subnet:24},
			nodes:0,
			max:100,
			id:2,
			iface:"eno1",
			ips:[]string{},
			switches:[]Switch{ Switch{addr:"172.16.2.1",iface:"eth0",brand:HP} }}