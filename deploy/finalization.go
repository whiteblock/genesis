package deploy

import (
	helpers "../blockchains/helpers"
	ssh "../ssh"
	testnet "../testnet"
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

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ int, localNodeNum int, _ int) error {
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

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ int, localNodeNum int, _ int) error {
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
