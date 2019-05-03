package rest

import (
	"../blockchains/registrar"
	"../db"
	"../manager"
	"../state"
	"../status"
	"../util"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

/*
   Returns a list of the commands in the response
*/
func getConfFiles(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	err := util.ValidateFilePath(params["blockchain"])
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	files, err := util.Lsr(fmt.Sprintf("./resources/" + params["blockchain"]))
	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("Nothing availible for \"%s\"", params["blockchain"]), 500)
		return
	}

	for i, file := range files {
		index := strings.LastIndex(file, "/")
		files[i] = file[index+1:]
	}

	json.NewEncoder(w).Encode(files)
}

/*
   Get a configuration file by blockchain and file name
*/
func getConfFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	err := util.ValidateFilePath(params["blockchain"])
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}
	err = util.ValidateFilePath(params["file"])
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}
	if strings.Contains(params["blockchain"], "..") || strings.Contains(params["file"], "..") {
		http.Error(w, "relative path operators not allowed", 401)
		return
	}
	if !strings.HasSuffix(params["file"], "mustache") && !strings.HasSuffix(params["file"], "json") {
		http.Error(w, "Cannot read non mustache/json files", 403)
		return
	}
	path := "./resources/" + params["blockchain"] + "/" + params["file"]
	fmt.Println(path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		http.Error(w, "File not found", 404)
		return
	}
	json.NewEncoder(w).Encode(string(data))
}

func getBlockChainParams(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	log.Println("GET PARAMS : " + params["blockchain"])
	blockchainParams, err := manager.GetParams(params["blockchain"])
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	log.Println(string(blockchainParams))
	w.Write(blockchainParams)
}

func getBlockChainState(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	buildID := params["buildID"]
	buildState, err := state.GetBuildStateByID(buildID)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	out, err := buildState.GetExtExtras()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(out)

}

func getBlockChainDefaults(w http.ResponseWriter, r *http.Request) {
	defaults, err := manager.GetDefaults(mux.Vars(r)["blockchain"])
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	w.Write(defaults)
}

func getBlockChainLog(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	nodeNum, err := strconv.Atoi(params["node"])
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}
	lines := -1
	_, ok := params["lines"]
	if ok {
		lines, err = strconv.Atoi(params["lines"])
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 400)
			return
		}
	}
	nodes, err := db.GetAllNodesByTestNet(params["testnetID"])
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 404)
		return
	}

	node, err := db.GetNodeByLocalID(nodes, nodeNum)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 404)
		return
	}

	client, err := status.GetClient(node.Server)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 404)
		return
	}
	res, err := client.DockerRead(node, conf.DockerOutputFile, lines)
	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("%s %s", res, err.Error()), 500)
		return
	}
	w.Write([]byte(res))
}

func getAllSupportedBlockchains(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(registrar.GetSupportedBlockchains())
}
