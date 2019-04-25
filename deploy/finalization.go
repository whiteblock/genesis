package deploy

import (
	helpers "../blockchains/helpers"
	db "../db"
	ssh "../ssh"
	state "../state"
	status "../status"
	testnet "../testnet"
	util "../util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

/*
   Finalization methods for the docker build process. Will be run immediately following their deployment
*/
func finalize(tn *testnet.TestNet) error {
	if conf.HandleNodeSshKeys {
		err := copyOverSshKeys(tn)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

/*
   Finalization methods for the docker build process. Will be run immediately following their deployment
*/
func finalizeNewNodes(tn *testnet.TestNet) error {
	if conf.HandleNodeSshKeys {
		err := copyOverSshKeysToNewNodes(tn)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

func alwaysRunFinalize(tn *testnet.TestNet) {

	tn.BuildState.Async(func() {
		for _, node := range tn.NewlyBuiltNodes {
			err := declareNode(&node, tn.LDD)
			if err != nil {
				log.Println(err)
			}
		}
	})

	tn.BuildState.Defer(func() {
		for i, node := range tn.NewlyBuiltNodes {
			err := finalizeNode(node, tn.LDD, tn.BuildState, i)
			if err != nil {
				log.Println(err)
			}
		}
	})

}

/*
   Copy over the ssh public key to each node to allow for the user to ssh into each node.
   The public key comes from the nodes public key specified in the configuration
*/
func copyOverSshKeys(tn *testnet.TestNet) error {
	tmp, err := ioutil.ReadFile(conf.NodesPublicKey)
	pubKey := string(tmp)
	pubKey = strings.Trim(pubKey, "\t\n\v\r")
	if err != nil {
		log.Println(err)
		return err
	}

	privKey, err := ioutil.ReadFile(conf.NodesPrivateKey)
	if err != nil {
		log.Println(err)
		return err
	}

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, localNodeNum int, _ int) error {
		defer tn.BuildState.IncrementDeployProgress()

		_, err := client.DockerExec(localNodeNum, "mkdir -p /root/.ssh/")
		if err != nil {
			log.Println(err)
			return err
		}
		_, err = client.DockerExec(localNodeNum, fmt.Sprintf(`bash -c 'echo "%s" >> /root/.ssh/authorized_keys'`, pubKey))
		if err != nil {
			log.Println(err)
			return err
		}

		_, err = client.DockerExecd(localNodeNum, "service ssh start")
		return err
	})
	if err != nil {
		log.Println(err)
		return err
	}

	return helpers.CopyBytesToAllNodes(tn, string(privKey), "/root/.ssh/id_rsa")
}

/*
   Functions like copyOverSshKeys, but with the add nodes format.
*/
func copyOverSshKeysToNewNodes(tn *testnet.TestNet) error {
	tmp, err := ioutil.ReadFile(conf.NodesPublicKey)
	pubKey := string(tmp)
	pubKey = strings.Trim(pubKey, "\t\n\v\r")
	if err != nil {
		log.Println(err)
		return err
	}

	privKey, err := ioutil.ReadFile(conf.NodesPrivateKey)
	if err != nil {
		log.Println(err)
		return err
	}

	err = helpers.AllNewNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, localNodeNum int, _ int) error { //TODO only run on new nodes
		defer tn.BuildState.IncrementDeployProgress()

		_, err := client.DockerExec(localNodeNum, "mkdir -p /root/.ssh/")
		if err != nil {
			log.Println(err)
			return err
		}
		_, err = client.DockerExec(localNodeNum, fmt.Sprintf(`bash -c 'echo "%s" >> /root/.ssh/authorized_keys'`, pubKey))
		if err != nil {
			log.Println(err)
			return err
		}

		_, err = client.DockerExecd(localNodeNum, "service ssh start")
		return err
	})
	if err != nil {
		log.Println(err)
		return err
	}

	return helpers.CopyBytesToAllNodes(tn, string(privKey), "/root/.ssh/id_rsa")
}

func declareNode(node *db.Node, details *db.DeploymentDetails) error {
	data := map[string]interface{}{
		"id":         node.TestNetID,
		"ip_address": node.Ip,
	}
	rawData, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = util.JwtHttpRequest("POST", "https://api.whiteblock.io/testnets/"+node.TestNetID+"/nodes", details.GetJwt(), string(rawData))
	return err
}

func finalizeNode(node db.Node, details *db.DeploymentDetails, buildState *state.BuildState, absNum int) error {
	client, err := status.GetClient(node.Server)
	if err != nil {
		log.Println(err)
		return err
	}
	files := conf.DockerOutputFile
	if details.Logs != nil && len(details.Logs) > 0 {
		var logFiles map[string]string
		if len(details.Logs) == 1 || len(details.Logs) <= absNum {
			logFiles = details.Logs[0]
		} else {
			logFiles = details.Logs[absNum]
		}
		for _, file := range logFiles { //Eventually may need to handle the names as well
			files += " " + file
		}
	}

	buildState.Defer(func() {
		_, err := client.DockerExecd(node.LocalID,
			fmt.Sprintf("nibbler --node-type %s --jwt %s --testnet %s --node %s %s",
				details.Blockchain, details.GetJwt(), node.TestNetID, node.ID, files))
		if err != nil {
			log.Println(err)
		}
	})
	return nil
}
