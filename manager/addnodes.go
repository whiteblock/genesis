package manager

import (
	"../blockchains/beam"
	"../blockchains/eos"
	"../blockchains/geth"
	"../blockchains/rchain"
	sys "../blockchains/syscoin"
	"../db"
	"../deploy"
	"../state"
	"../testnet"
	"fmt"
	"log"
)

/*
   AddNodes allows for nodes to be added to the network.
   The nodes don't need to be of the same type of the original build.
   It is worth noting that any missing information from the given
   deployment details will be filled in from the origin build.
*/
func AddNodes(details *db.DeploymentDetails, testnetID string) error {
	buildState, err := state.GetBuildStateById(testnetID)
	if err != nil {
		log.Println(err)
		return err
	}

	tn, err := testnet.RestoreTestNet(testnetID)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}
	defer tn.FinishedBuilding()

	err = tn.AddDetails(*details)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
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
		buildState.ReportError(fmt.Errorf("too many nodes"))
		return fmt.Errorf("too many nodes")
	}

	err = deploy.AddNodes(tn)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return err
	}
	var labels []string = nil
	switch details.Blockchain {
	case "eos":
		labels, err = eos.Add(tn)
		if err != nil {
			buildState.ReportError(err)
			log.Println(err)
			return err
		}
	case "ethereum":
		fallthrough
	case "geth":
		labels, err = geth.Add(tn)
		if err != nil {
			buildState.ReportError(err)
			log.Println(err)
			return err
		}
	case "syscoin":
		labels, err = sys.Add(tn)
		if err != nil {
			buildState.ReportError(err)
			log.Println(err)
			return err
		}
	case "rchain":
		labels, err = rchain.Add(tn)
		if err != nil {
			buildState.ReportError(err)
			log.Println(err)
			return err
		}
	case "beam":
		labels, err = beam.Add(tn)
		if err != nil {
			buildState.ReportError(err)
			log.Println(err)
			return err
		}
	case "generic":
		log.Println("Built in generic mode")
	default:
		buildState.ReportError(fmt.Errorf("unknown blockchain"))
		return fmt.Errorf("unknown blockchain")
	}
	err = tn.StoreNodes(labels)
	if err != nil {
		log.Println(err.Error())
		buildState.ReportError(err)
		return err
	}

	return nil
}
