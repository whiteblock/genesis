package testnet

import (
	db "../db"
	ssh "../ssh"
	state "../state"
	status "../status"
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

/*
	Represents a testnet and some state on that testnet. Should also contain the details needed to
	rebuild this testnet.
*/
type TestNet struct {
	TestNetID       string
	Servers         []db.Server
	Nodes           []db.Node
	NewlyBuiltNodes []db.Node
	Clients         map[int]*ssh.Client
	BuildState      *state.BuildState
	Details         []db.DeploymentDetails
	CombinedDetails db.DeploymentDetails
	LDD             *db.DeploymentDetails //ptr to latest deployment details
	mux             *sync.RWMutex
}

func RestoreTestNet(buildID string) (*TestNet, error) {
	out := new(TestNet)
	err := db.GetMetaP("testnet_"+buildID, out)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	bs, err := state.GetBuildStateById(buildID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	out.BuildState = bs
	out.mux = &sync.RWMutex{}
	out.LDD = out.GetLastestDeploymentDetails()

	out.Clients = map[int]*ssh.Client{}
	for _, server := range out.Servers {
		out.Clients[server.Id], err = status.GetClient(server.Id)
		if err != nil {
			log.Println(err)
			out.BuildState.ReportError(err)
			return nil, err
		}
	}
	return out, nil
}

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

	out.BuildState, err = state.GetBuildStateById(buildID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// FETCH THE SERVERS
	out.Servers, err = db.GetServers(details.Servers)
	if err != nil {
		log.Println(err)
		out.BuildState.ReportError(err)
		return nil, err
	}
	fmt.Println("Got the Servers")

	//OPEN UP THE RELEVANT SSH CONNECTIONS
	out.Clients = map[int]*ssh.Client{}

	for _, server := range out.Servers {
		out.Clients[server.Id], err = status.GetClient(server.Id)
		if err != nil {
			log.Println(err)
			out.BuildState.ReportError(err)
			return nil, err
		}
	}
	return out, nil
}

func (this *TestNet) AddNode(node db.Node) int {
	this.mux.Lock()
	defer this.mux.Unlock()
	node.AbsoluteNum = len(this.Nodes)
	this.NewlyBuiltNodes = append(this.NewlyBuiltNodes, node)
	this.Nodes = append(this.Nodes, node)
	return node.AbsoluteNum
}

func (this *TestNet) AddDetails(dd db.DeploymentDetails) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.Details = append(this.Details, dd)
	//MERGE
	tmp, err := json.Marshal(dd)
	if err != nil {
		log.Println(err)
		return err
	}
	this.LDD = &this.Details[len(this.Details)-1]
	return json.Unmarshal(tmp, &this.CombinedDetails)
}

func (this *TestNet) FinishedBuilding() {
	this.BuildState.DoneBuilding()
	this.NewlyBuiltNodes = []db.Node{}
	this.Store()
}

func (this *TestNet) GetFlatClients() []*ssh.Client {
	out := []*ssh.Client{}
	this.mux.RLock()
	defer this.mux.RUnlock()
	for _, client := range this.Clients {
		out = append(out, client)
	}
	return out
}

func (this *TestNet) GetServer(id int) *db.Server {
	for _, server := range this.Servers {
		if server.Id == id {
			return &server
		}
	}
	return nil
}

func (this *TestNet) GetLastestDeploymentDetails() *db.DeploymentDetails {
	this.mux.RLock()
	defer this.mux.RUnlock()
	return &this.Details[len(this.Details)-1]
}

func (this *TestNet) PreOrderNodes() map[int][]db.Node {
	this.mux.RLock()
	defer this.mux.RUnlock()

	out := make(map[int][]db.Node)
	for _, server := range this.Servers {
		out[server.Id] = []db.Node{}
	}

	for _, node := range this.Nodes {
		out[node.Server] = append(out[node.Server], node)
	}
	return out
}

func (this *TestNet) PreOrderNewNodes() map[int][]db.Node {
	this.mux.RLock()
	defer this.mux.RUnlock()

	out := make(map[int][]db.Node)
	for _, server := range this.Servers {
		out[server.Id] = []db.Node{}
	}

	for _, node := range this.NewlyBuiltNodes {
		out[node.Server] = append(out[node.Server], node)
	}
	return out
}

func (this *TestNet) Store() {
	db.SetMeta("testnet_"+this.TestNetID, *this)
}

func (this *TestNet) StoreNodes(labels []string) error {
	for i, node := range this.NewlyBuiltNodes {
		if labels != nil {
			node.Label = labels[i]
		}

		_, err := db.InsertNode(node)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}
