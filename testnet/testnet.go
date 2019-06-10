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

//Package testnet helps to manage and control current testnets
package testnet

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/state"
	"github.com/whiteblock/genesis/status"
	"sync"
)

// TestNet represents a testnet and some state on that testnet. Should also contain the details needed to
// rebuild tn testnet.
type TestNet struct {
	TestNetID string
	// Servers contains the servers on which the testnet resides
	Servers []db.Server
	// Nodes contains the active nodes in the network, including the newly built nodes
	Nodes []db.Node
	// NewlyBuiltNodes contains only the nodes created in the last/in progress build event
	NewlyBuiltNodes []db.Node

	SideCars [][]db.SideCar

	NewlyBuiltSideCars [][]db.SideCar

	// Clients is a map of server ids to ssh clients
	Clients map[int]ssh.Client `json:"-"`
	// BuildState is the build state for the test net
	BuildState *state.BuildState `json:"-"`
	// Details contains all of the past deployments to tn test net
	Details []db.DeploymentDetails
	// CombinedDetails contains all of the deployments merged into one
	CombinedDetails db.DeploymentDetails
	// LDD is a pointer to latest deployment details
	LDD *db.DeploymentDetails `json:"-"`
	mux *sync.RWMutex
}

// RestoreTestNet fetches a testnet which already exists.
func RestoreTestNet(buildID string) (*TestNet, error) {
	out := new(TestNet)
	err := db.GetMetaP("testnet_"+buildID, out)
	if err != nil {
		log.WithFields(log.Fields{"build": buildID}).Error("failed to restore the testnet")
		return nil, err
	}
	bs, err := state.GetBuildStateByID(buildID)
	if err != nil {
		log.WithFields(log.Fields{"build": buildID}).Error("failed to restore the build state")
		return nil, err
	}
	out.BuildState = bs
	out.mux = &sync.RWMutex{}
	out.LDD = out.GetLastestDeploymentDetails()

	out.Clients = map[int]ssh.Client{}
	for _, server := range out.Servers {
		out.Clients[server.ID], err = status.GetClient(server.ID)
		if err != nil {
			log.WithFields(log.Fields{"build": buildID, "server": server.ID}).Error("failed to get ssh connection")
			out.BuildState.ReportError(err)
			return nil, err
		}
	}
	return out, nil
}

// NewTestNet creates a new TestNet
func NewTestNet(details db.DeploymentDetails, buildID string) (*TestNet, error) {
	var err error
	out := new(TestNet)

	out.TestNetID = buildID
	out.Nodes = []db.Node{}
	out.NewlyBuiltNodes = []db.Node{}
	out.Details = []db.DeploymentDetails{details}
	out.CombinedDetails = details
	out.LDD = &details
	out.mux = &sync.RWMutex{}

	out.BuildState, err = state.GetBuildStateByID(buildID)
	if err != nil {
		log.WithFields(log.Fields{"build": buildID}).Error("failed to create the build state")
		return nil, err
	}

	// FETCH THE SERVERS
	out.Servers, err = db.GetServers(details.Servers)
	if err != nil {
		log.WithFields(log.Fields{"build": buildID}).Error("failed to fetch the servers")
		out.BuildState.ReportError(err)
		return nil, err
	}
	log.WithFields(log.Fields{"build": buildID}).Trace("fetched the servers")

	//OPEN UP THE RELEVANT SSH CONNECTIONS
	out.Clients = map[int]ssh.Client{}

	for _, server := range out.Servers {
		out.Clients[server.ID], err = status.GetClient(server.ID)
		if err != nil {
			log.WithFields(log.Fields{"build": buildID, "server": server.ID}).Error("failed to get ssh connection")
			out.BuildState.ReportError(err)
			return nil, err
		}
	}
	return out, nil
}

// AddNode adds a node to the testnet and returns a pointer to that node.
func (tn *TestNet) AddNode(node db.Node) *db.Node {
	tn.mux.Lock()
	defer tn.mux.Unlock()
	node.AbsoluteNum = len(tn.Nodes)
	tn.NewlyBuiltNodes = append(tn.NewlyBuiltNodes, node)
	tn.Nodes = append(tn.Nodes, node)
	return &tn.Nodes[node.AbsoluteNum]
}

// AddSideCar adds a side car to the testnet
func (tn *TestNet) AddSideCar(node db.SideCar, index int) {
	tn.mux.Lock()
	defer tn.mux.Unlock()
	if len(tn.NewlyBuiltSideCars) <= index {
		tn.NewlyBuiltSideCars = append(tn.NewlyBuiltSideCars, []db.SideCar{node})
	} else {
		tn.NewlyBuiltSideCars[index] = append(tn.NewlyBuiltSideCars[index], node)
	}
	if len(tn.SideCars) <= index {
		tn.SideCars = append(tn.SideCars, []db.SideCar{node})
	} else {
		tn.SideCars[index] = append(tn.SideCars[index], node)
	}
}

// AddDetails adds the details of a new deployment to the TestNet
func (tn *TestNet) AddDetails(dd db.DeploymentDetails) error {
	tn.mux.Lock()
	defer tn.mux.Unlock()
	tn.Details = append(tn.Details, dd)
	//MERGE
	tmp, err := json.Marshal(dd)
	if err != nil {
		log.WithFields(log.Fields{"build": tn.TestNetID}).Error(
			"failed to marshal the deploymentdetails into json")
		return err
	}
	tn.LDD = &tn.Details[len(tn.Details)-1]

	oldCD := tn.CombinedDetails
	err = json.Unmarshal(tmp, &tn.CombinedDetails)
	if err != nil {
		log.WithFields(log.Fields{"build": tn.TestNetID}).Error(
			"failed to combine the build details into all the other build details")
	}

	/**Handle Files**/
	tn.CombinedDetails.Files = oldCD.Files
	if dd.Files != nil && len(dd.Files) > 0 {
		if tn.CombinedDetails.Files == nil {
			tn.CombinedDetails.Files = make([]map[string]string, oldCD.Nodes)
		}
		if len(tn.CombinedDetails.Files) < oldCD.Nodes {
			for i := len(tn.CombinedDetails.Files); i < oldCD.Nodes; i++ {
				tn.CombinedDetails.Files = append(tn.CombinedDetails.Files, map[string]string{})
			}
		}
		for _, files := range dd.Files {
			tn.CombinedDetails.Files = append(tn.CombinedDetails.Files, files)
		}
	}

	/**Handle Nodes**/
	tn.CombinedDetails.Nodes = oldCD.Nodes + dd.Nodes

	/**Handle Images***/
	if dd.Images != nil && len(dd.Images) > 0 {
		if tn.CombinedDetails.Images == nil {
			tn.CombinedDetails.Images = make([]string, oldCD.Nodes)
		}
		if len(tn.CombinedDetails.Images) < oldCD.Nodes {
			for i := len(tn.CombinedDetails.Images); i < oldCD.Nodes; i++ {
				tn.CombinedDetails.Images = append(tn.CombinedDetails.Images, tn.CombinedDetails.Images[0])
			}
		}
		for _, image := range dd.Images {
			tn.CombinedDetails.Images = append(tn.CombinedDetails.Images, image)
		}
	}
	return nil
}

// FinishedBuilding empties the NewlyBuiltNodes, signals DoneBuilding on the BuildState, and
// stores the current data of tn testnet
func (tn *TestNet) FinishedBuilding() {
	tn.BuildState.DoneBuilding()
	tn.NewlyBuiltNodes = []db.Node{}
	tn.Store()
}

// GetFlatClients takes the clients map and turns it into an array
// for easy iterator
func (tn *TestNet) GetFlatClients() []ssh.Client {
	out := []ssh.Client{}
	tn.mux.RLock()
	defer tn.mux.RUnlock()
	for _, client := range tn.Clients {
		out = append(out, client)
	}
	return out
}

// GetServer retrieves a server the TestNet resides on by id
func (tn *TestNet) GetServer(id int) *db.Server {
	for _, server := range tn.Servers {
		if server.ID == id {
			return &server
		}
	}
	return nil
}

// GetLastestDeploymentDetails gets a pointer to the latest deployment details
func (tn *TestNet) GetLastestDeploymentDetails() *db.DeploymentDetails {
	tn.mux.RLock()
	defer tn.mux.RUnlock()
	return &tn.Details[len(tn.Details)-1]
}

// PreOrderNodes sorts the nodes into buckets by server id
func (tn *TestNet) PreOrderNodes(newNodes bool, sidecar bool, index int) map[int][]ssh.Node {
	tn.mux.RLock()
	defer tn.mux.RUnlock()

	out := make(map[int][]ssh.Node)
	for _, server := range tn.Servers {
		out[server.ID] = []ssh.Node{}
	}
	if !newNodes && sidecar {
		for _, node := range tn.SideCars[index] {
			out[node.Server] = append(out[node.Server], node)
		}
	} else if !newNodes && !sidecar {
		for _, node := range tn.Nodes {
			out[node.Server] = append(out[node.Server], node)
		}
	} else if newNodes && sidecar {
		for _, node := range tn.NewlyBuiltSideCars[index] {
			out[node.Server] = append(out[node.Server], node)
		}
	} else {
		for _, node := range tn.NewlyBuiltNodes {
			out[node.Server] = append(out[node.Server], node)
		}
	}

	return out
}

// PreOrderNewNodes sorts the newly built nodes into buckets by server id
func (tn *TestNet) PreOrderNewNodes(sidecar bool) map[int][]ssh.Node {
	tn.mux.RLock()
	defer tn.mux.RUnlock()

	out := make(map[int][]ssh.Node)
	for _, server := range tn.Servers {
		out[server.ID] = []ssh.Node{}
	}

	return out
}

// Store stores the TestNets data for later retrieval
func (tn *TestNet) Store() {
	db.SetMeta("testnet_"+tn.TestNetID, *tn)
}

// Destroy removes all the testnets data
func (tn *TestNet) Destroy() error {
	return db.DeleteMeta("testnet_" + tn.TestNetID)
}

// StoreNodes stores the newly built nodes into the database with their labels.
func (tn *TestNet) StoreNodes() error {
	var err error
	for _, node := range tn.NewlyBuiltNodes {
		_, er := db.InsertNode(node)
		if er != nil {
			log.WithFields(log.Fields{"build": tn.TestNetID, "error": er,
				"node": node.ID}).Error("failed to store a node into db")
			err = er
		}
	}
	return err
}

// GetSSHNodes gets all nodes or sidecars wrapper in the
// ssh Node interface
func (tn *TestNet) GetSSHNodes(newNodes bool, sidecar bool, index int) []ssh.Node {
	out := []ssh.Node{}
	if !newNodes && sidecar {
		for _, node := range tn.SideCars[index] {
			out = append(out, node)
		}
	} else if !newNodes && !sidecar {
		for _, node := range tn.Nodes {
			out = append(out, node)
		}
	} else if newNodes && sidecar {
		for _, node := range tn.NewlyBuiltSideCars[index] {
			out = append(out, node)
		}
	} else {
		for _, node := range tn.NewlyBuiltNodes {
			out = append(out, node)
		}
	}
	return out
}

// SpawnAdjunct generates info on an adjunct new by index
func (tn *TestNet) SpawnAdjunct(newNodes bool, index int) (*Adjunct, error) {
	if index >= len(tn.SideCars) {
		return nil, fmt.Errorf("index out of range")
	}
	return &Adjunct{
		Main:       tn,
		Index:      index,
		Nodes:      tn.GetSSHNodes(newNodes, true, index),
		BuildState: tn.BuildState,
		LDD:        tn.LDD,
	}, nil
}

// GetNodesSideCar Get's a nodes sidecar by name
func (tn *TestNet) GetNodesSideCar(node ssh.Node, name string) (*db.SideCar, error) {
	index := -1
	for i := range tn.SideCars {
		if tn.SideCars[i][0].Type == name {
			index = i
			break
		}
	}
	if index == -1 {
		return nil, fmt.Errorf("could not find any side cars of type \"%s\"", name)
	}
	if node.GetAbsoluteNumber() >= len(tn.SideCars[index]) {
		return nil, fmt.Errorf("given node index out of range")
	}

	return &tn.SideCars[index][node.GetAbsoluteNumber()], nil
}
