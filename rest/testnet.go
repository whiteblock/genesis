package rest

import (
	db "../db"
	manager "../manager"
	state "../state"
	status "../status"
	util "../util"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func getAllTestNets(w http.ResponseWriter, r *http.Request) {
	testNets, err := db.GetAllTestNets()
	if err != nil {
		log.Println(err)
		http.Error(w, "There are no test nets", 204)
		return
	}
	json.NewEncoder(w).Encode(testNets)
}

func createTestNet(w http.ResponseWriter, r *http.Request) {
	tn := &db.DeploymentDetails{}
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	err := decoder.Decode(tn)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}
	jwt, err := util.ExtractJwt(r)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 403)
		return
	}
	tn.SetJwt(jwt)
	id, err := status.GetNextTestNetId()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error Generating a new UUID", 500)
		return
	}

	err = state.AcquireBuilding(tn.Servers, id)
	if err != nil {
		log.Println(err)
		http.Error(w, "There is a build already in progress", 409)
		return
	}

	go manager.AddTestNet(tn, id)
	w.Write([]byte(id))

}

func getTestNetInfo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	//log.Println(fmt.Sprintf("Attempting to find tn with id %d",id))
	testNet, err := db.GetTestNet(params["id"])
	if err != nil {
		log.Println(err)
		http.Error(w, "Test net does not exist", 404)
		return
	}
	err = json.NewEncoder(w).Encode(testNet)
	if err != nil {
		log.Println(err)
	}
}

func deleteTestNet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	err := manager.DeleteTestNet(params["id"])
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte("Success"))
}

func getTestNetNodes(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	nodes, err := db.GetAllNodesByTestNet(params["id"])
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 404)
		return
	}
	json.NewEncoder(w).Encode(nodes)
}

func addNodes(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	testnetId := params["testnetid"]

	tn, err := db.GetBuildByTestnet(testnetId)
	if err != nil {
		log.Println(err)
		http.Error(w, "Could not find the given testnet id", 400)
		return
	}

	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	err = decoder.Decode(&tn)
	if err != nil {
		log.Println(err)
		//Ignore error and continue
	}
	bs, err := state.GetBuildStateById(testnetId)
	if err != nil {
		log.Println(err)
		http.Error(w, "Testnet is down, build a new one", 409)
		return
	}
	bs.Reset()
	w.Write([]byte("Adding the nodes"))
	go manager.AddNodes(&tn, testnetId)
}

func delNodes(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	num, err := strconv.Atoi(params["num"])
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid id", 400)
		return
	}

	testnetId := params["id"]

	tn, err := db.GetBuildByTestnet(testnetId)
	if err != nil {
		log.Println(err)
		http.Error(w, "Could not find the given testnet id", 400)
		return
	}

	err = state.AcquireBuilding(tn.Servers, testnetId) //TODO: THIS IS WRONG
	if err != nil {
		log.Println(err)
		http.Error(w, "There is a build in progress", 409)
		return
	}
	w.Write([]byte("Deleting the nodes"))
	go testnet.DelNodes(num, testnetId)
}

func restartNode(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	testnetId := params["id"]
	node := params["num"]
	log.Printf("%s %s\n", testnetId, node)
	bs, err := state.GetBuildStateById(testnetId)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 404)
		return
	}
	cmdRaw, ok := bs.Get(node)
	fmt.Printf("%#v\n", bs.GetExtras())
	if !ok {
		log.Printf("Node %s not found", node)
		http.Error(w, fmt.Sprintf("Node %s not found", node), 404)
		return
	}
	cmd := cmdRaw.(util.Command)

	client, err := status.GetClient(cmd.ServerId)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	cmdgexCmd := fmt.Sprintf("ps aux | grep '%s' | grep -v grep|  awk '{print $2}'| tail -n 1", strings.Split(cmd.Cmdline, " ")[0])
	pid, err := client.DockerExec(cmd.Node, cmdgexCmd)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	res, err := client.DockerExec(cmd.Node, fmt.Sprintf("kill -INT %s", pid))
	if err != nil {
		log.Println(err)
		log.Println(res)
		http.Error(w, err.Error(), 500)
		return
	}

	for {
		_, err = client.DockerExec(cmd.Node, fmt.Sprintf("ps aux | grep '%s' | grep -v grep", strings.Split(cmd.Cmdline, " ")[0]))
		if err != nil {
			break
		}
	}
	err = client.DockerExecdLogAppend(cmd.Node, cmd.Cmdline)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte("Success"))

}
