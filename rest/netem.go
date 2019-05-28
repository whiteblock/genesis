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
	"github.com/whiteblock/genesis/db"
	netem "github.com/whiteblock/genesis/net"
	"github.com/whiteblock/genesis/status"
	"github.com/whiteblock/genesis/util"
	"net/http"
	"strconv"
)

func handleNet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var netConf []netem.Netconf
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	err := decoder.Decode(&netConf)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 400)
		return
	}

	nodes, err := db.GetAllNodesByTestNet(params["testnetID"])
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}

	err = netem.ApplyAll(netConf, nodes)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
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
		http.Error(w, util.LogError(err).Error(), 400)
		return
	}

	nodes, err := db.GetAllNodesByTestNet(params["testnetID"])
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}

	netem.RemoveAll(nodes)
	err = netem.ApplyToAll(netConf, nodes)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
	}
	w.Write([]byte("Success"))
}

func stopNet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	nodes, err := db.GetAllNodesByTestNet(params["testnetID"])
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}

	netem.RemoveAll(nodes)

	w.Write([]byte("Success"))
}

func getNet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	servers, err := status.GetLatestServers(params["testnetID"])
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}
	out := []netem.Netconf{}
	for _, server := range servers {
		client, err := status.GetClient(server.ID)
		if err != nil {
			http.Error(w, util.LogError(err).Error(), 404)
			return
		}
		confs, err := netem.GetConfigOnServer(client)
		if err != nil {
			http.Error(w, util.LogError(err).Error(), 500)
			return
		}
		out = append(out, confs...)
	}
	json.NewEncoder(w).Encode(out)
}

func removeOrAddOutage(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	testnetID := params["testnetID"]
	nodeNum1, err := strconv.Atoi(params["node1"])
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 400)
		return
	}

	nodeNum2, err := strconv.Atoi(params["node2"])
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 400)
		return
	}

	nodes, err := db.GetAllNodesByTestNet(testnetID)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}

	node1, err := db.GetNodeByAbsNum(nodes, nodeNum1)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}

	node2, err := db.GetNodeByAbsNum(nodes, nodeNum2)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}
	switch r.Method {
	case "POST":
		err = netem.MakeOutage(node1, node2)
	case "DELETE":
		err = netem.RemoveOutage(node1, node2)
	default:
		err = fmt.Errorf("unexpected http method")
	}
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}
	w.Write([]byte("Success"))
}

func partitionOutage(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	nodeNums := []int{}
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	err := decoder.Decode(&nodeNums)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 400)
		return
	}
	nodes, err := db.GetAllNodesByTestNet(params["testnetID"])
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}
	side1, side2, err := db.DivideNodesByAbsMatch(nodes, nodeNums)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 400)
		return
	}
	netem.CreatePartitionOutage(side1, side2)
	w.Write([]byte("success"))
}

func removeAllOutages(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	servers, err := status.GetLatestServers(params["testnetID"])
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}

	for _, server := range servers {
		client, err := status.GetClient(server.ID)
		if err != nil {
			http.Error(w, util.LogError(err).Error(), 404)
			return
		}
		err = netem.RemoveAllOutages(client)
		if err != nil {
			http.Error(w, util.LogError(err).Error(), 500)
			return
		}
	}
	w.Write([]byte("Success"))
}

func getAllOutages(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	servers, err := status.GetLatestServers(params["testnetID"])
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}
	out := []netem.Connection{}
	for _, server := range servers {
		client, err := status.GetClient(server.ID)
		if err != nil {
			http.Error(w, util.LogError(err).Error(), 404)
			return
		}
		conns, err := netem.GetCutConnections(client)
		if err != nil {
			http.Error(w, util.LogError(err).Error(), 500)
			return
		}
		out = append(out, conns...)
	}
	nodeRaw, exists := params["node"]
	if exists {
		node, err := strconv.Atoi(nodeRaw)
		if err != nil {
			http.Error(w, util.LogError(err).Error(), 400)
			return
		}
		filteredOut := []netem.Connection{}
		for _, conn := range out {
			if conn.To == node || conn.From == node {
				filteredOut = append(filteredOut, conn)
			}
		}
		json.NewEncoder(w).Encode(filteredOut)
		return
	}
	json.NewEncoder(w).Encode(out)
}

func getAllPartitions(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	nodes, err := db.GetAllNodesByTestNet(params["testnetID"])
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 404)
		return
	}

	out, err := netem.CalculatePartitions(nodes)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(out)
}
