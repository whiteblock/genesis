package artemis

import (
	db "../../db"
	ssh "../../ssh"
	state "../../state"
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
func Build(details *db.DeploymentDetails, servers []db.Server, clients []*ssh.Client,
	buildState *state.BuildState) ([]string, error) {

	artemisConf, err := NewConf(details.Params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	buildState.SetBuildSteps(0 + (details.Nodes * 4))

	err = helpers.AllNodeExecCon(servers, buildState, func(serverNum int, localNodeNum int, absoluteNodeNum int) error {
		defer buildState.IncrementBuildProgress()
		_, err := clients[serverNum].DockerExec(localNodeNum, "rm /artemis/config/config.toml")
		return err
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	port := 9000
	peers := "["
	var peer string
	for _, server := range servers {
		for i, ip := range server.Ips {
			peer = fmt.Sprintf("%s://whiteblock-node%d@%s:%d",
				artemisConf["networkMode"],
				i,
				ip,
				port,
			)
			if i < details.Nodes-1 {
				peers = peers + "\"" + peer + "\"" + ","
			} else {
				peers = peers + "\"" + peer + "\""
			}
			buildState.IncrementBuildProgress()
		}
	}
	peers = peers + "]"
	fmt.Println(peers)

	buildState.SetBuildStage("Creating node configuration files")
	/**Create node config files**/

	err = helpers.CreateConfigs(servers, clients, buildState, "/artemis/config/config.toml",
		func(serverNum int, localNodeNum int, absoluteNodeNum int) ([]byte, error) {
			defer buildState.IncrementBuildProgress()
			identity := fmt.Sprintf("0x%.8x", absoluteNodeNum)
			artemisNodeConfig, err := makeNodeConfig(artemisConf, identity, peers, absoluteNodeNum, details)
			return []byte(artemisNodeConfig), err
		})

	buildState.SetBuildStage("Starting Artemis")
	err = helpers.AllNodeExecCon(servers, buildState, func(serverNum int, localNodeNum int, absoluteNodeNum int) error {
		artemisCmd := `artemis -c /artemis/config/config.toml -o /artemis/data/data.json 2>&1 | tee /output.log`
		_, err = clients[serverNum].DockerExecd(localNodeNum, "tmux new -s whiteblock -d")
		if err != nil {
			log.Println(err)
			return err
		}

		_, err = clients[serverNum].DockerExecd(localNodeNum, fmt.Sprintf("tmux send-keys -t whiteblock '%s' C-m", artemisCmd))
		if err != nil {
			log.Println(err)
			return err
		}

		buildState.IncrementBuildProgress()

		_, err = clients[serverNum].DockerExecd(localNodeNum,
			fmt.Sprintf("bash -c 'while :;do artemis-log-parser --influx \"http://%s:8086\" --node \"%s%d\" "+
				"/artemis/data/data.json 2>&1 >> /parser.log; done'",
				util.GetGateway(servers[serverNum].SubnetID, localNodeNum), conf.NodePrefix, absoluteNodeNum))
		return err
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return nil, nil
}

func Add(details *db.DeploymentDetails, servers []db.Server, clients []*ssh.Client,
	newNodes map[int][]string, buildState *state.BuildState) ([]string, error) {
	return nil, nil
}
