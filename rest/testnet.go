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
	"github.com/whiteblock/genesis/state"
	"github.com/whiteblock/genesis/status"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"net/http"
	"strconv"
	"strings"
)

func createTestNet(w http.ResponseWriter, r *http.Request) {
	tn := &db.DeploymentDetails{}
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	err := decoder.Decode(tn)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 400)
		return
	}
	jwt, err := util.ExtractJwt(r)
	/*if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 403)
		return
	}*/
	tn.SetJwt(jwt)

	id, err := util.GetUUIDString()
	if err != nil {
		util.LogError(err)
		http.Error(w, "Error Generating a new UUID", 500)
		return
	}

	err = state.AcquireBuilding(tn.Servers, id)
	if err != nil {
		util.LogError(err)
		http.Error(w, "There is a build already in progress", 409)
		return
	}

	go manager.AddTestNet(tn, id)
	w.Write([]byte(id))

}

func deleteTestNet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	err := manager.DeleteTestNet(params["id"])
	if err != nil {

		http.Error(w, util.LogError(err).Error(), 500)
		return
	}
	w.Write([]byte("Success"))
}

func getTestNetNodes(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	nodes, err := db.GetAllNodesByTestNet(params["id"])
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}
	json.NewEncoder(w).Encode(nodes)
}

func addNodes(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	testnetID := params["testnetID"]

	tn, err := db.GetBuildByTestnet(testnetID)
	if err != nil {
		util.LogError(err)
		http.Error(w, "Could not find the given testnet id", 400)
		return
	}

	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	err = decoder.Decode(&tn)
	if err != nil {
		util.LogError(err)
		//Ignore error and continue
	}
	bs, err := state.GetBuildStateByID(testnetID)
	if err != nil {
		util.LogError(err)
		http.Error(w, "Testnet is down, build a new one", 409)
		return
	}
	bs.Reset()
	w.Write([]byte("Adding the nodes"))
	go manager.AddNodes(&tn, testnetID)
}

func delNodes(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	num, err := strconv.Atoi(params["num"])
	if err != nil {
		util.LogError(err)
		http.Error(w, "Invalid id", 400)
		return
	}

	testnetID := params["id"]

	tn, err := db.GetBuildByTestnet(testnetID)
	if err != nil {
		util.LogError(err)
		http.Error(w, "Could not find the given testnet id", 400)
		return
	}

	err = state.AcquireBuilding(tn.Servers, testnetID) //TODO: THIS IS WRONG
	if err != nil {
		util.LogError(err)
		http.Error(w, "There is a build in progress", 409)
		return
	}
	w.Write([]byte("Deleting the nodes"))
	go manager.DelNodes(num, testnetID)
}

func restartNode(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	testnetID := params["id"]
	nodeNum := params["num"]
	log.Printf("%s %s\n", testnetID, nodeNum)
	tn, err := testnet.RestoreTestNet(testnetID)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}
	cmdRaw, ok := tn.BuildState.Get(nodeNum)
	log.WithFields(log.Fields{"extras": tn.BuildState.GetExtras()}).Debug("fetched the previous build state")
	if !ok {
		log.Printf("Node %s not found", nodeNum)
		http.Error(w, fmt.Sprintf("Node %s not found", nodeNum), 404)
		return
	}
	cmd := cmdRaw.(util.Command)

	client, err := status.GetClient(cmd.ServerID)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}
	cmdgexCmd := fmt.Sprintf("ps aux | grep '%s' | grep -v grep|  awk '{print $2}'| tail -n 1", strings.Split(cmd.Cmdline, " ")[0])
	node, err := db.GetNodeByLocalID(tn.Nodes, cmd.Node)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}
	pid, err := client.DockerExec(node, cmdgexCmd)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}
	_, err = client.DockerExec(node, fmt.Sprintf("kill -INT %s", pid))
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}

	for {
		_, err = client.DockerExec(node, fmt.Sprintf("ps aux | grep '%s' | grep -v grep", strings.Split(cmd.Cmdline, " ")[0]))
		if err != nil {
			break
		}
	}

	err = client.DockerExecdLogAppend(node, cmd.Cmdline)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}
	w.Write([]byte("Success"))
}

func signalNode(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	testnetID := params["testnetID"]
	node := params["node"]
	nodeNum, err := strconv.Atoi(node)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 400)
		return
	}
	signal := params["signal"]
	err = util.ValidateCommandLine(signal)
	if err != nil {
		util.LogError(err)
		http.Error(w, fmt.Sprintf("Invalid signal \"%s\", see `man 7 signal` for help", signal), 400)
	}

	tn, err := testnet.RestoreTestNet(testnetID)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}
	if nodeNum >= len(tn.Nodes) {
		http.Error(w, fmt.Sprintf("Node %d does not exist. Try node 0 through node %d", nodeNum, len(tn.Nodes)), 400)
		return
	}
	n := &tn.Nodes[nodeNum]
	cmdRaw, ok := tn.BuildState.Get(node)
	if !ok {
		log.Printf("Node %s not found", node)
		http.Error(w, fmt.Sprintf("Node %s not found", node), 404)
		return
	}
	cmd := cmdRaw.(util.Command)

	cmdgexCmd := fmt.Sprintf("ps aux | grep '%s' | grep -v grep|  awk '{print $2}'| tail -n 1", strings.Split(cmd.Cmdline, " ")[0])
	pid, err := tn.Clients[n.Server].DockerExec(n, cmdgexCmd)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}

	_, err = tn.Clients[n.Server].DockerExec(n, fmt.Sprintf("kill -%s %s", signal, pid))
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}
	w.Write([]byte(fmt.Sprintf("Sent signal %s to node %s", signal, node)))
}

func killNode(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	testnetID := params["testnetID"]

	log.Printf("%s %s\n", testnetID, params["node"])
	tn, err := testnet.RestoreTestNet(testnetID)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}
	cmdRaw, ok := tn.BuildState.Get(params["node"])
	if !ok {
		log.Printf("Node %s not found", params["node"])
		http.Error(w, fmt.Sprintf("Node %s not found", params["node"]), 404)
		return
	}
	cmd := cmdRaw.(util.Command)

	client, err := status.GetClient(cmd.ServerID)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}
	cmdgexCmd := fmt.Sprintf("ps aux | grep '%s' | grep -v grep|  awk '{print $2}'| tail -n 1", strings.Split(cmd.Cmdline, " ")[0])
	node, err := db.GetNodeByLocalID(tn.Nodes, cmd.Node)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	pid, err := client.DockerExec(node, cmdgexCmd)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}
	res, err := client.DockerExec(node, fmt.Sprintf("kill -INT %s", pid))
	if err != nil {
		log.Println(res)
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}

	for {
		_, err = client.DockerExec(node, fmt.Sprintf("ps aux | grep '%s' | grep -v grep", strings.Split(cmd.Cmdline, " ")[0]))
		if err != nil {
			break
		}
	}
	w.Write([]byte(fmt.Sprintf("Killed node %s", params["node"])))
}
