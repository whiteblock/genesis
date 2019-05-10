/*
	Copyright 2019 Whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Genesis is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

// Package rest implements the REST interface which is used to communicate with this module
package rest

import (
	"encoding/json"
	"github.com/Whiteblock/genesis/db"
	"github.com/Whiteblock/genesis/state"
	"github.com/Whiteblock/genesis/status"
	"github.com/Whiteblock/genesis/util"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

// StartServer starts the rest server, blocking the calling thread from returning
func StartServer() {
	router := mux.NewRouter()
	router.Use(authN)
	router.HandleFunc("/servers", getAllServerInfo).Methods("GET")

	router.HandleFunc("/servers/{name}", addNewServer).Methods("PUT")

	router.HandleFunc("/servers/{id}", getServerInfo).Methods("GET")
	router.HandleFunc("/servers/{id}", deleteServer).Methods("DELETE")
	router.HandleFunc("/servers/{id}", updateServerInfo).Methods("UPDATE")

	router.HandleFunc("/testnets", createTestNet).Methods("POST") //Create new test net

	router.HandleFunc("/testnets/{id}", deleteTestNet).Methods("DELETE")

	router.HandleFunc("/testnets/{id}/nodes", getTestNetNodes).Methods("GET")

	/**Management Functions**/
	router.HandleFunc("/status/nodes/{testnetID}", nodesStatus).Methods("GET")

	router.HandleFunc("/status/build/{id}", buildStatus).Methods("GET")

	router.HandleFunc("/params/{blockchain}", getBlockChainParams).Methods("GET")

	router.HandleFunc("/state/{buildID}", getBlockChainState).Methods("GET")

	router.HandleFunc("/defaults/{blockchain}", getBlockChainDefaults).Methods("GET")

	router.HandleFunc("/log/{testnetID}/{node}", getBlockChainLog).Methods("GET")

	router.HandleFunc("/log/{testnetID}/{node}/{lines}", getBlockChainLog).Methods("GET")

	router.HandleFunc("/nodes/{id}", getTestNetNodes).Methods("GET")

	router.HandleFunc("/nodes/{testnetID}", addNodes).Methods("POST")

	router.HandleFunc("/nodes/{id}/{num}", delNodes).Methods("DELETE") //Completely remove x nodes

	router.HandleFunc("/nodes/restart/{testnetID}/{num}", restartNode).Methods("POST")

	router.HandleFunc("/nodes/raise/{testnetID}/{node}/{signal}", signalNode).Methods("POST")

	router.HandleFunc("/nodes/kill/{testnetID}/{node}", killNode).Methods("POST")

	router.HandleFunc("/build/{id}", stopBuild).Methods("DELETE")

	router.HandleFunc("/build", getPreviousBuild).Methods("GET")

	router.HandleFunc("/build/{id}", getBuild).Methods("GET")

	router.HandleFunc("/build/freeze/{id}", freezeBuild).Methods("POST")

	router.HandleFunc("/build/thaw/{id}", thawBuild).Methods("POST")
	router.HandleFunc("/build/freeze/{id}", thawBuild).Methods("DELETE")

	router.HandleFunc("/emulate/{testnetID}", getNet).Methods("GET")

	router.HandleFunc("/emulate/{testnetID}", stopNet).Methods("DELETE")

	router.HandleFunc("/emulate/{testnetID}", handleNet).Methods("POST")

	router.HandleFunc("/emulate/all/{testnetID}", handleNetAll).Methods("POST")

	router.HandleFunc("/resources/{blockchain}", getConfFiles).Methods("GET")

	router.HandleFunc("/resources/{blockchain}/{file}", getConfFile).Methods("GET")

	router.HandleFunc("/outage/{testnetID}/{node1}/{node2}", removeOrAddOutage).Methods("POST")

	router.HandleFunc("/outage/{testnetID}/{node1}/{node2}", removeOrAddOutage).Methods("DELETE")

	router.HandleFunc("/outage/{testnetID}", removeAllOutages).Methods("DELETE")

	router.HandleFunc("/outage/{testnetID}", getAllOutages).Methods("GET")

	router.HandleFunc("/outage/{testnetID}/{node}", getAllOutages).Methods("GET")

	router.HandleFunc("/partition/{testnetID}", partitionOutage).Methods("POST")

	router.HandleFunc("/partition/{testnetID}", getAllPartitions).Methods("GET")

	router.HandleFunc("/blockchains", getAllSupportedBlockchains).Methods("GET")

	http.ListenAndServe(conf.Listen, removeTrailingSlash(router))
}

func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		next.ServeHTTP(w, r)
	})
}

func nodesStatus(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	testnetID, ok := params["testnetID"]
	if !ok {
		http.Error(w, "Missing testnet id", 400)
		return
	}

	nodes, err := db.GetAllNodesByTestNet(testnetID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	out, err := status.CheckNodeStatus(nodes)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(out)
}

func buildStatus(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	buildID, ok := params["id"]
	if !ok {
		http.Error(w, "Missing build id", 400)
		return
	}
	res, err := status.CheckBuildStatus(buildID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 404)
		return
	}
	w.Write([]byte(res))
}

func stopBuild(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	buildID, ok := params["id"]
	if !ok {
		http.Error(w, "Missing build id", 400)
		return
	}
	err := state.SignalStop(buildID)
	if err != nil {
		http.Error(w, err.Error(), 412)
		return
	}
	w.Write([]byte("Stop signal has been sent"))
}

func freezeBuild(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	bState, err := state.GetBuildStateByID(params["id"])
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	err = bState.Freeze()
	if err != nil {
		http.Error(w, err.Error(), 409)
		return
	}
	w.Write([]byte("Build has been frozen"))
}

func thawBuild(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	bState, err := state.GetBuildStateByID(params["id"])
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	err = bState.Unfreeze()
	if err != nil {
		http.Error(w, err.Error(), 409)
		return
	}
	w.Write([]byte("Build has been resumed"))
}

func getPreviousBuild(w http.ResponseWriter, r *http.Request) {

	jwt, err := util.ExtractJwt(r)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 403)
		return
	}
	kid, err := util.GetKidFromJwt(jwt)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 403)
	}
	build, err := db.GetLastBuildByKid(kid)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 404)
		return
	}
	json.NewEncoder(w).Encode(build)
}

func getBuild(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id := params["id"]

	build, err := db.GetBuildByTestnet(id)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	err = json.NewEncoder(w).Encode(build)
	if err != nil {
		log.Println(err)
	}
}
