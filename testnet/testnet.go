package testnet

import(
	db "../db"
	ssh "../ssh"
	state "../state"
	"log"
)

/*
	Represents a testnet and some state on that testnet. Should also contain the details needed to 
	rebuild this testnet.
 */
type TestNet struct {
	TestNetID		string
	Servers 		[]db.Server
	Nodes			[]db.Node
	NewlyBuiltNodes []db.Node
	Clients			map[int]*ssh.Client
	BuildState  	*state.BuildState
	Details 		[]db.DeploymentDetails
	CombinedDetails	db.DeploymentDetails
}

func NewTestNet(details db.DeploymentDetails, buildID string) (*TestNet,error) {
	var err error
	out := new()

	this.TestNetID = buildID
	this.Nodes = []db.Node{}
	this.NewlyBuiltNodes = []db.Node{}
	this.Details = []db.DeploymentDetails{details}
	this.CombinedDetails = details

	this.BuildState,err = state.GetBuildStateById(buildID)
	if err != nil {
		log.Println(err)
		return nil,err
	}
	
	// FETCH THE SERVERS
	this.Servers, err = db.GetServers(details.Servers)
	if err != nil {
		log.Println(err)
		this.BuildState.ReportError(err)
		return nil,err
	}
	fmt.Println("Got the Servers")

	//OPEN UP THE RELEVANT SSH CONNECTIONS
	this.Clients = map[int]*ssh.Client{}

	for _,server := range this.Servers {
		this.Clients[server.Id],err = state.GetClient(server.Id)
		if err != nil {
			log.Println(err)
			this.BuildState.ReportError(err)
			return nil,err
		}
	}
}