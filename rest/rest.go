/*
	Implements the REST interface which is used to communicate with this module
*/
package rest

import (
	db "../db"
	state "../state"
	status "../status"
	util "../util"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

/*
	Starts the rest server, blocking the calling thread from returning
*/
func StartServer() {
	router := mux.NewRouter()
	router.Use(AuthN)
	router.HandleFunc("/servers", getAllServerInfo).Methods("GET")
	router.HandleFunc("/servers/", getAllServerInfo).Methods("GET")

	router.HandleFunc("/servers/{name}", addNewServer).Methods("PUT") //Private

	router.HandleFunc("/servers/{id}", getServerInfo).Methods("GET")
	router.HandleFunc("/servers/{id}", deleteServer).Methods("DELETE")     //Private
	router.HandleFunc("/servers/{id}", updateServerInfo).Methods("UPDATE") //Private

	router.HandleFunc("/testnets/", getAllTestNets).Methods("GET")
	router.HandleFunc("/testnets", getAllTestNets).Methods("GET")

	router.HandleFunc("/testnets/", createTestNet).Methods("POST") //Create new test net
	router.HandleFunc("/testnets", createTestNet).Methods("POST")  //Create new test net

	router.HandleFunc("/testnets/{id}", getTestNetInfo).Methods("GET")
	router.HandleFunc("/testnets/{id}/", getTestNetInfo).Methods("GET")

	router.HandleFunc("/testnets/{id}", deleteTestNet).Methods("DELETE")
	router.HandleFunc("/testnets/{id}/", deleteTestNet).Methods("DELETE")

	router.HandleFunc("/testnets/{id}/nodes", getTestNetNodes).Methods("GET")
	router.HandleFunc("/testnets/{id}/nodes/", getTestNetNodes).Methods("GET")

	/**Management Functions**/
	router.HandleFunc("/status/nodes/{testnetId}", nodesStatus).Methods("GET")
	router.HandleFunc("/status/nodes/{testnetId}/", nodesStatus).Methods("GET")

	router.HandleFunc("/status/build/{id}", buildStatus).Methods("GET")
	router.HandleFunc("/status/build/{id}/", buildStatus).Methods("GET")

	router.HandleFunc("/params/{blockchain}", getBlockChainParams).Methods("GET")
	router.HandleFunc("/params/{blockchain}/", getBlockChainParams).Methods("GET")

	router.HandleFunc("/state/{buildId}", getBlockChainState).Methods("GET")
	router.HandleFunc("/state/{buildId}/", getBlockChainState).Methods("GET")

	router.HandleFunc("/defaults/{blockchain}", getBlockChainDefaults).Methods("GET")
	router.HandleFunc("/defaults/{blockchain}/", getBlockChainDefaults).Methods("GET")

	router.HandleFunc("/log/{testnetId}/{node}", getBlockChainLog).Methods("GET")
	router.HandleFunc("/log/{testnetId}/{node}/", getBlockChainLog).Methods("GET")

	router.HandleFunc("/log/{testnetId}/{node}/{lines}", getBlockChainLog).Methods("GET")
	router.HandleFunc("/log/{testnetId}/{node}/{lines}/", getBlockChainLog).Methods("GET")

	router.HandleFunc("/nodes/{id}", getTestNetNodes).Methods("GET")
	router.HandleFunc("/nodes/{id}/", getTestNetNodes).Methods("GET")

	router.HandleFunc("/nodes/{id}/{num}", addNodes).Methods("POST")
	router.HandleFunc("/nodes/{id}/{num}/", addNodes).Methods("POST")

	router.HandleFunc("/nodes/{id}/{num}", delNodes).Methods("DELETE")
	router.HandleFunc("/nodes/{id}/{num}/", delNodes).Methods("DELETE")

	router.HandleFunc("/nodes/restart/{id}/{num}", restartNode).Methods("POST")
	router.HandleFunc("/nodes/restart/{id}/{num}/", restartNode).Methods("POST")

	router.HandleFunc("/build/{id}", stopBuild).Methods("DELETE")
	router.HandleFunc("/build/{id}/", stopBuild).Methods("DELETE")

	router.HandleFunc("/build", getAllBuilds).Methods("GET")
	router.HandleFunc("/build/", getAllBuilds).Methods("GET")

	router.HandleFunc("/build/{id}", getBuild).Methods("GET")
	router.HandleFunc("/build/{id}/", getBuild).Methods("GET")

	router.HandleFunc("/build/freeze/{id}", freezeBuild).Methods("POST")
	router.HandleFunc("/build/freeze/{id}/", freezeBuild).Methods("POST")

	router.HandleFunc("/build/thaw/{id}", thawBuild).Methods("POST")
	router.HandleFunc("/build/thaw/{id}/", thawBuild).Methods("POST")
	router.HandleFunc("/build/freeze/{id}", thawBuild).Methods("DELETE")
	router.HandleFunc("/build/freeze/{id}/", thawBuild).Methods("DELETE")

	router.HandleFunc("/emulate/{testnetId}", getNet).Methods("GET")
	router.HandleFunc("/emulate/{testnetId}/", getNet).Methods("GET")

	router.HandleFunc("/emulate/{testnetId}", stopNet).Methods("DELETE")
	router.HandleFunc("/emulate/{testnetId}/", stopNet).Methods("DELETE")

	router.HandleFunc("/emulate/{testnetId}", handleNet).Methods("POST")
	router.HandleFunc("/emulate/{testnetId}/", handleNet).Methods("POST")

	router.HandleFunc("/emulate/all/{testnetId}", handleNetAll).Methods("POST")
	router.HandleFunc("/emulate/all/{testnetId}/", handleNetAll).Methods("POST")

	router.HandleFunc("/resources/{blockchain}", getConfFiles).Methods("GET")
	router.HandleFunc("/resources/{blockchain}/", getConfFiles).Methods("GET")

	router.HandleFunc("/resources/{blockchain}/{file}", getConfFile).Methods("GET")
	router.HandleFunc("/resources/{blockchain}/{file}/", getConfFile).Methods("GET")

	http.ListenAndServe(conf.Listen, router)
}

func nodesStatus(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	testnetId, ok := params["testnetId"]
	if !ok {
		http.Error(w, "Missing testnet id", 400)
		return
	}

	nodes, err := db.GetAllNodesByTestNet(testnetId)
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
	buildId, ok := params["id"]
	if !ok {
		http.Error(w, "Missing build id", 400)
		return
	}
	res, err := status.CheckBuildStatus(buildId)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 404)
		return
	}
	w.Write([]byte(res))
}

func stopBuild(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	buildId, ok := params["id"]
	if !ok {
		http.Error(w, "Missing build id", 400)
		return
	}
	err := state.SignalStop(buildId)
	if err != nil {
		http.Error(w, err.Error(), 412)
		return
	}
	w.Write([]byte("Stop signal has been sent"))
}

func freezeBuild(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	bState, err := state.GetBuildStateById(params["id"])
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

	bState, err := state.GetBuildStateById(params["id"])
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

func getAllBuilds(w http.ResponseWriter, r *http.Request) {
	builds, err := db.GetAllBuilds()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 404)
		return
	}
	json.NewEncoder(w).Encode(builds)
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
