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
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/ssh"
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
	if err != nil && conf.RequireAuth {
		http.Error(w, util.LogError(err).Error(), 403)
		return
	}
	tn.SetJwt(jwt)

	id, err := util.GetUUIDString()
	if err != nil {
		util.LogError(err)
		http.Error(w, "Error Generating a new UUID", 500)
		return
	}
	_, ok := tn.Extras["forceUnlock"]
	if ok && tn.Extras["forceUnlock"].(bool) {
		state.ForceUnlockServers(tn.Servers)
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

func getNodePids(tn *testnet.TestNet, n ssh.Node, node string) ([]string, error) {
	cmdsToTry, err := helpers.GetCommandExprs(tn, node)
	if err != nil {
		return nil, util.LogError(err)
	}
	log.WithFields(log.Fields{"toTry": cmdsToTry}).Info("got the commands to try")
	out := []string{}
	for _, cmd := range cmdsToTry {
		pid, err := tn.Clients[n.GetServerID()].DockerExec(n, fmt.Sprintf(
			"ps aux | grep '%s' | grep -v grep | grep -v nibbler |  awk '{print $2}'", cmd))
		if err == nil {
			out = append(out, strings.Split(pid, "\n")...)
		}
	}
	return out, nil
}

func restartNode(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	testnetID := params["id"]
	nodeNum := params["num"]
	log.WithFields(log.Fields{"testnet": testnetID, "node": nodeNum}).Info("restarting a node")
	tn, err := testnet.RestoreTestNet(testnetID)
	if err != nil {
		util.LogError(err)
		http.Error(w, fmt.Sprintf("unable to restore testnet \"%s\"", testnetID), 404)
		return
	}
	var cmd util.Command
	ok := tn.BuildState.GetP(nodeNum, &cmd)
	log.WithFields(log.Fields{"extras": tn.BuildState.GetExtras()}).Debug("fetched the previous build state")
	if !ok {
		log.WithFields(log.Fields{"node": nodeNum}).Error("node not found")
		http.Error(w, fmt.Sprintf("Node %s not found", nodeNum), 404)
		return
	}

	client, err := status.GetClient(cmd.ServerID)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}
	node := &tn.Nodes[cmd.Node]
	procs, err := getNodePids(tn, node, nodeNum)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}
	log.WithFields(log.Fields{"procs": procs}).Debug("got the possible process ids")

	for _, pid := range procs {
		if pid == "" {
			continue
		}
		_, err = client.DockerExec(node, fmt.Sprintf("kill -INT %s", pid))
		if err != nil {
			http.Error(w, util.LogError(err).Error(), 500)
			return
		}
	}

	killedSuccessfully := false
	for i := uint(0); i < conf.KillRetries; i++ {
		_, err = client.DockerExec(node,
			fmt.Sprintf("ps aux | grep '%s' | grep -v grep | grep -v nibbler", strings.Split(cmd.Cmdline, " ")[0]))
		if err != nil {
			killedSuccessfully = true
			break
		}
	}

	if !killedSuccessfully {
		err := fmt.Errorf("Unable to kill the blockchain process")
		http.Error(w, util.LogError(err).Error(), 500)
		return
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
	log.WithFields(log.Fields{"testnet": testnetID, "node": nodeNum, "signal": signal}).Info("sending signal to node")
	err = util.ValidateCommandLine(signal)
	if err != nil {
		util.LogError(err)
		http.Error(w, fmt.Sprintf("invalid signal \"%s\", see `man 7 signal` for help", signal), 400)
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
	procs, err := getNodePids(tn, tn.Nodes[nodeNum], node)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}
	log.WithFields(log.Fields{"procs": procs}).Debug("got the possible process ids")

	for _, pid := range procs {
		if pid == "" {
			continue
		}
		_, err = tn.Clients[n.GetServerID()].DockerExec(n, fmt.Sprintf("kill -%s %s", signal, pid))
	}
	w.Write([]byte(fmt.Sprintf("Sent signal %s to node %s", signal, node)))
}

func killNode(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	testnetID := params["testnetID"]
	log.WithFields(log.Fields{"testnet": testnetID, "node": params["node"]}).Info("killing a node's main process")
	tn, err := testnet.RestoreTestNet(testnetID)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}
	var cmd util.Command
	ok := tn.BuildState.GetP(params["node"], &cmd)
	if !ok {
		log.WithFields(log.Fields{"node": params["node"]}).Warn("node not found")
		http.Error(w, fmt.Sprintf("Node %s not found", params["node"]), 404)
		return
	}

	client, err := status.GetClient(cmd.ServerID)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}
	cmdgexCmd := fmt.Sprintf("ps aux | grep '%s' | grep -v grep|  awk '{print $2}'| tail -n 1", strings.Split(cmd.Cmdline, " ")[0])
	node, err := db.GetNodeByLocalID(tn.Nodes, cmd.Node)
	if err != nil {
		log.WithFields(log.Fields{"node": cmd.Node, "error": err}).Error("error getting node from db")
		http.Error(w, err.Error(), 500)
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
		_, err = client.DockerExec(node, fmt.Sprintf("ps aux | grep '%s' | grep -v grep | grep -v nibbler", strings.Split(cmd.Cmdline, " ")[0]))
		if err != nil {
			break
		}
	}
	w.Write([]byte(fmt.Sprintf("Killed node %s", params["node"])))
}
