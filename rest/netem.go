package rest

import (
	db "../db"
	netem "../net"
	status "../status"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func handleNet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var netConf []netem.Netconf
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	err := decoder.Decode(&netConf)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	nodes, err := db.GetAllNodesByTestNet(params["testnetId"])
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	//fmt.Printf("GIVEN %v\n",netConf)
	err = netem.ApplyAll(netConf, nodes)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte("Success"))
}

func handleNetAll(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var netConf netem.Netconf
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()

	err := decoder.Decode(&netConf)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	nodes, err := db.GetAllNodesByTestNet(params["testnetId"])
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	netem.RemoveAll(nodes)
	err = netem.ApplyToAll(netConf, nodes)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
	}
	w.Write([]byte("Success"))
}

func stopNet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	nodes, err := db.GetAllNodesByTestNet(params["testnetId"])
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	netem.RemoveAll(nodes)

	w.Write([]byte("Success"))
}

func getNet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	servers, err := status.GetLatestServers(params["testnetId"])
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 404)
		return
	}
	out := []netem.Netconf{}
	for _, server := range servers {
		client, err := status.GetClient(server.Id)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 404)
			return
		}
		confs, err := netem.GetConfigOnServer(client)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}
		out = append(out, confs...)
	}

	output, err := json.Marshal(out)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(output)
}
