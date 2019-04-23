package testnet

import (
	beam "../blockchains/beam"
	eos "../blockchains/eos"
	geth "../blockchains/geth"
	rchain "../blockchains/rchain"
	sys "../blockchains/syscoin"
	db "../db"
	deploy "../deploy"
	state "../state"
	status "../status"
	util "../util"
	"fmt"
	"log"
)

/*
   AddNodes allows for nodes to be added to the network.
   The nodes don't need to be of the same type of the original build.
   It is worth noting that any missing information from the given
   deployment details will be filled in from the origin build.
*/
func AddNodes(details *db.DeploymentDetails, testnetId string) error {
	buildState, err := state.GetBuildStateById(testnetId)
	if err != nil {
		log.Println(err)
		return err
	}
	defer buildState.DoneBuilding()

	//STEP 1: MERGE IN MISSING INFO FROM ORIGINAL BUILD
	prevDetails, err := db.GetBuildByTestnet(testnetId)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}
	if details.Servers == nil || len(details.Servers) == 0 {
		details.Servers = prevDetails.Servers
	}

	if len(details.Blockchain) == 0 {
		details.Blockchain = prevDetails.Blockchain
	}

	if len(details.Images) == 0 {
		details.Images = prevDetails.Images
	}

	if details.Params == nil {
		details.Params = prevDetails.Params
	}

	//STEP 2: VALIDATE
	for i, res := range details.Resources {
		err = res.ValidateAndSetDefaults()
		if err != nil {
			log.Println(err)
			err = fmt.Errorf("%s. For node %d", err.Error(), i)
			buildState.ReportError(err)
			return err
		}
	}

	if details.Nodes > conf.MaxNodes {
		buildState.ReportError(fmt.Errorf("Too many nodes"))
		return fmt.Errorf("Too many nodes")
	}
	//STEP 3: FETCH THE SERVERS
	servers, err := status.GetLatestServers(testnetId)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}
	fmt.Println("Got the Servers")

	//STEP 4: OPEN UP THE RELEVANT SSH CONNECTIONS
	clients, err := status.GetClients(details.Servers)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}

	nodes, err := deploy.AddNodes(details, servers, clients, buildState)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}
	var labels []string = nil
	switch details.Blockchain {
	case "eos":
		labels, err = eos.Add(details, servers, clients, nodes, buildState)
		if err != nil {
			buildState.ReportError(err)
			log.Println(err)
			return err
		}
	case "ethereum":
		fallthrough
	case "geth":
		labels, err = geth.Add(details, servers, clients, nodes, buildState)
		if err != nil {
			buildState.ReportError(err)
			log.Println(err)
			return err
		}
	case "syscoin":
		labels, err = sys.Add(details, servers, clients, nodes, buildState)
		if err != nil {
			buildState.ReportError(err)
			log.Println(err)
			return err
		}
	case "rchain":
		labels, err = rchain.Add(details, servers, clients, nodes, buildState)
		if err != nil {
			buildState.ReportError(err)
			log.Println(err)
			return err
		}
	case "beam":
		labels, err = beam.Add(details, servers, clients, nodes, buildState)
		if err != nil {
			buildState.ReportError(err)
			log.Println(err)
			return err
		}
	case "generic":
		log.Println("Built in generic mode")
	default:
		buildState.ReportError(fmt.Errorf("Unknown blockchain"))
		return fmt.Errorf("Unknown blockchain")
	}

	i := 0
	for serverId, ips := range nodes {
		for _, ip := range ips {
			id, err := util.GetUUIDString()
			if err != nil {
				log.Println(err.Error())
				buildState.ReportError(err)
				return err
			}
			node := db.Node{Id: id, TestNetId: testnetId, Server: serverId, LocalId: i, Ip: ip}
			if labels != nil {
				node.Label = labels[i]
			}
			_, err = db.InsertNode(node)
			if err != nil {
				log.Println(err.Error())
			}
			i++
		}

	}

	return nil
}
