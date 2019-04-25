package artemis

import (
	db "../../db"
	ssh "../../ssh"
	testnet "../../testnet"
	util "../../util"
	helpers "../helpers"
	"fmt"
	"log"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

/*
Build builds out a fresh new artemis test network
*/
func Build(tn *testnet.TestNet) ([]string, error) {
	artemisConf, err := NewConf(tn.LDD.Params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	tn.BuildState.SetBuildSteps(0 + (tn.LDD.Nodes * 4))

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, localNodeNum int, absoluteNodeNum int) error {
		defer tn.BuildState.IncrementBuildProgress()
		_, err := client.DockerExec(localNodeNum, "rm /artemis/config/config.toml")
		return err
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	port := 9000
	peers := "["
	var peer string
	for i, node := range tn.Nodes {
		peer = fmt.Sprintf("%s://whiteblock-node%d@%s:%d",
			artemisConf["networkMode"],
			node.LocalID,
			node.Ip,
			port,
		)
		if i != len(tn.Nodes)-1 {
			peers = peers + "\"" + peer + "\"" + ","
		} else {
			peers = peers + "\"" + peer + "\""
		}
		tn.BuildState.IncrementBuildProgress()
	}

	peers = peers + "]"
	fmt.Println(peers)

	tn.BuildState.SetBuildStage("Creating node configuration files")
	/**Create node config files**/

	err = helpers.CreateConfigs(tn, "/artemis/config/config.toml",
		func(serverId int, localNodeNum int, absoluteNodeNum int) ([]byte, error) {
			defer tn.BuildState.IncrementBuildProgress()
			identity := fmt.Sprintf("0x%.8x", absoluteNodeNum)
			artemisNodeConfig, err := makeNodeConfig(artemisConf, identity, peers, absoluteNodeNum, tn.LDD)
			return []byte(artemisNodeConfig), err
		})

	tn.BuildState.SetBuildStage("Starting Artemis")
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, localNodeNum int, absoluteNodeNum int) error {
		artemisCmd := `artemis -c /artemis/config/config.toml -o /artemis/data/data.json 2>&1 | tee /output.log`

		_, err := client.DockerExecd(localNodeNum, "tmux new -s whiteblock -d")
		if err != nil {
			log.Println(err)
			return err
		}

		_, err = client.DockerExecd(localNodeNum, fmt.Sprintf("tmux send-keys -t whiteblock '%s' C-m", artemisCmd))
		if err != nil {
			log.Println(err)
			return err
		}

		tn.BuildState.IncrementBuildProgress()

		_, err = client.DockerExecd(localNodeNum,
			fmt.Sprintf("bash -c 'while :;do artemis-log-parser --influx \"http://%s:8086\" --node \"%s%d\" "+
				"/artemis/data/data.json 2>&1 >> /parser.log; done'",
				util.GetGateway(server.SubnetID, localNodeNum), conf.NodePrefix, absoluteNodeNum))
		return err
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return nil, nil
}

func Add(tn *testnet.TestNet) ([]string, error) {
	return nil, nil
}
