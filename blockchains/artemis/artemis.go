package artemis

import (
	db "../../db"
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
func Build(details db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
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
	err = helpers.AllNodeExecCon(servers, buildState, func(serverNum int, localNodeNum int, absoluteNodeNum int) error {

		identity := fmt.Sprintf("0x%.8x", absoluteNodeNum) // potential error if application reads the identity as a string literal

		artemisNodeConfig, err := makeNodeConfig(artemisConf, identity, peers, details.Nodes, details.Params)
		if err != nil {
			log.Println(err)
			return err
		}

		fmt.Println("Writing Configuration File")
		err = buildState.Write(fmt.Sprintf("config.toml%d", absoluteNodeNum), artemisNodeConfig)
		if err != nil {
			log.Println(err)
			return err
		}

		err = clients[serverNum].Scp(fmt.Sprintf("config.toml%d", absoluteNodeNum), fmt.Sprintf("/home/appo/config.toml%d", absoluteNodeNum))
		if err != nil {
			log.Println(err)
			return err
		}
		buildState.Defer(func() { clients[serverNum].Run(fmt.Sprintf("rm -f /home/appo/config.toml%d", absoluteNodeNum)) })

		err = clients[serverNum].DockerCp(localNodeNum, fmt.Sprintf("/home/appo/config.toml%d", absoluteNodeNum), "/artemis/config/config.toml")
		if err != nil {
			log.Println(err)
			return err
		}
		buildState.IncrementBuildProgress()
		return nil
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

func Add(details db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
	newNodes map[int][]string, buildState *state.BuildState) ([]string, error) {
	return nil, nil
}
