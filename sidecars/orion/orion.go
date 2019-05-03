package orion

import (
	"../../db"
	"../../ssh"
	"../../testnet"
	"../../util"
	// "../../state"
	"../helpers"
	"../registrar"
	"fmt"
	"log"
	// "strings"
	// "sync"
	// "github.com/Whiteblock/mustache"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
	blockchain := "orion"
	registrar.RegisterBuild(blockchain, Build)
	registrar.RegisterAddNodes(blockchain, Add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, GetDefaults)
	registrar.RegisterParams(blockchain, GetParams)
}

func Build(tn *testnet.TestNet) ([]string, error) {
	// mux := sync.Mutex{}

	orionconf, err := newConf(tn.LDD.Params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	fmt.Println(orionconf)

	tn.BuildState.SetBuildSteps(6*tn.LDD.Nodes + 2)
	tn.BuildState.IncrementBuildProgress()

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, localNodeNum int, _ int) error {
		defer tn.BuildState.IncrementBuildProgress()
		_, err = client.DockerExec(localNodeNum, "mkdir /orion/data")
		return err
	})

	err = helpers.CreateConfigs(tn, "/orion/data/orion.conf",
		func(serverId int, localNodeNum int, absoluteNodeNum int) ([]byte, error) {
			defer tn.BuildState.IncrementBuildProgress()
			orionNodeConfig, err := makeNodeConfig(orionconf, absoluteNodeNum, tn.LDD)
			return []byte(orionNodeConfig), err
		})

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, localNodeNum int, absoluteNodeNum int) error {
		defer tn.BuildState.IncrementBuildProgress()
		_, err := client.DockerExec(localNodeNum, "bash -c 'cd /orion/data && orion -g nodeKey'")
		return err
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, localNodeNum int, absoluteNodeNum int) error {
		defer tn.BuildState.IncrementBuildProgress()
		return client.DockerExecdLog(localNodeNum, "orion /orion/data/orion.conf")
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return nil, err
}

func Add(tn *testnet.TestNet) ([]string, error) {
	return nil, nil
}
