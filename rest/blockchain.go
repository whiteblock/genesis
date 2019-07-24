/*
	Copyright 2019 whiteblock Inc.
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

package rest

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/manager"
	"github.com/whiteblock/genesis/protocols/registrar"
	"github.com/whiteblock/genesis/state"
	"github.com/whiteblock/genesis/status"
	"github.com/whiteblock/genesis/util"
	"io/ioutil"
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
		http.Error(w, util.LogError(err).Error(), 400)
		return
	}

	files, err := util.Lsr(fmt.Sprintf("./resources/" + params["blockchain"]))
	if err != nil {
		log.WithFields(log.Fields{"error": err, "blockchain": params["blockchain"]}).Error("not found")
		http.Error(w, fmt.Sprintf("Nothing available for \"%s\"", params["blockchain"]), 500)
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
		http.Error(w, util.LogError(err).Error(), 400)
		return
	}
	err = util.ValidateFilePath(params["file"])
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 400)
		return
	}
	if strings.Contains(params["blockchain"], "..") || strings.Contains(params["file"], "./") {
		http.Error(w, "relative path operators not allowed", 401)
		return
	}
	path := "./resources/" + params["blockchain"] + "/" + params["file"]

	log.WithFields(log.Fields{"path": path, "blockchain": params["blockchain"], "file": params["file"]}).Debug("got the file path")

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.WithFields(log.Fields{"path": path, "error": err}).Error("error reading the requested config")
		http.Error(w, "File not found", 404)
		return
	}
	util.LogError(json.NewEncoder(w).Encode(string(data)))
}

func getBlockChainParams(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	log.WithFields(log.Fields{"blockchain": params["blockchain"]}).Debug("getting params")
	blockchainParams, err := manager.GetParams(params["blockchain"])
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}
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
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}
	w.Write(defaults)
}

func getBlockChainLog(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	nodeNum, err := strconv.Atoi(params["node"])
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 400)
		return
	}
	lines := -1
	_, ok := params["lines"]
	if ok {
		lines, err = strconv.Atoi(params["lines"])
		if err != nil {
			http.Error(w, util.LogError(err).Error(), 400)
			return
		}
	}
	nodes, err := db.GetAllNodesByTestNet(params["testnetID"])
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}

	node, err := db.GetNodeByLocalID(nodes, nodeNum)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}

	client, err := status.GetClient(node.Server)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}
	res, err := client.DockerRead(node, conf.DockerOutputFile, lines)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s %s", res, util.LogError(err).Error()), 500)
		return
	}
	w.Write([]byte(res))
}

func getAllSupportedBlockchains(w http.ResponseWriter, r *http.Request) {
	util.LogError(json.NewEncoder(w).Encode(registrar.GetSupportedBlockchains()))
}
