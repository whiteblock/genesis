package orion

import (
	"../../db"
	"../../ssh"
	"../../testnet"
	"../../util"
	// "../../state"
	"../../blockchains/helpers"
	"../../blockchains/registrar"
	"fmt"
	"log"
	// "strings"
	// "sync"
	// "github.com/Whiteblock/mustache"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
	sidecar := "orion"
	registrar.RegisterSideCar(sidecar, registrar.SideCar{
		Image: "gcr.io/whiteblock/orion:dev",
	})
	registrar.RegisterBuildSideCar(sidecar, Build)
	registrar.RegisterAddSideCar(sidecar, Add)
}

func Build(tn *testnet.TestNet) ( error) {
	// mux := sync.Mutex{}

	orionconf, err := newConf(tn)
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Println(orionconf)

	tn.BuildState.SetBuildSteps(6*tn.LDD.Nodes + 2)
	tn.BuildState.IncrementBuildProgress()

	err = helpers.AllNodeExecConSC(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		_, err = client.DockerExec(node, "mkdir /orion/data")
		return err
	})

	err = helpers.CreateConfigsSC(tn, "/orion/data/orion.conf",
		func(node ssh.Node) ([]byte, error) {
			defer tn.BuildState.IncrementBuildProgress()
			orionNodeConfig, err := makeNodeConfig(orionconf, node.GetAbsoluteNumber(), tn.LDD)
			return []byte(orionNodeConfig), err
		})

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		_, err := client.DockerExec(node, "bash -c 'cd /orion/data && echo \"\" | orion -g nodeKey'")
		return err
	})
	if err != nil {
		log.Println(err)
		return err
	}

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		return client.DockerExecdLog(node, "orion /orion/data/orion.conf")
	})
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func Add(tn *testnet.TestNet) (error) {
	return nil
}
