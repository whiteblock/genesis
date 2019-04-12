package artemis

import (
	db "../../db"
	state "../../state"
	util "../../util"
	"fmt"
	"log"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

func Build(details db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
	buildState *state.BuildState) ([]string, error) {

	artemisConf, err := NewConf(details.Params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	buildState.SetBuildSteps(0 + (details.Nodes * 4))

	for i, server := range servers {
		for localId, _ := range server.Ips {
			_, err := clients[i].DockerExec(localId, "rm /artemis/config/config.toml")
			if err != nil {
				log.Println(err)
				return nil, err
			}
			buildState.IncrementBuildProgress()
		}
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
	for i, server := range servers {
		for j, _ := range server.Ips {
			buildState.IncrementBuildProgress()

			// potential error if application reads the identity as a string literal
			identity := fmt.Sprintf("0x%.8x", j)

			artemisNodeConfig, err := makeNodeConfig(artemisConf, identity, peers, details.Nodes, details.Params)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			fmt.Println("Writing Configuration File")
			err = buildState.Write("config.toml", artemisNodeConfig)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			err = clients[i].Scp("config.toml", "/home/appo/config.toml")
			if err != nil {
				log.Println(err)
				return nil, err
			}
			defer clients[i].Run("rm -f /home/appo/config.toml")

			err = clients[i].DockerCp(j, "/home/appo/config.toml", "/artemis/config/config.toml")
			if err != nil {
				log.Println(err)
				return nil, err
			}
		}
	}

	buildState.SetBuildStage("Starting Artemis")
	node := 0
	for i, server := range servers {
		for localId, _ := range server.Ips {
			artemisCmd := fmt.Sprintf(
				`artemis -c /artemis/config/config.toml -o /artemis/data/data.json 2>&1 | tee /output.log`,
			)
			_, err = clients[i].DockerExecd(localId, "tmux new -s whiteblock -d")
			if err != nil {
				log.Println(err)
				return nil, err
			}

			_, err = clients[i].DockerExecd(localId, fmt.Sprintf("tmux send-keys -t whiteblock '%s' C-m", artemisCmd))
			if err != nil {
				log.Println(err)
				return nil, err
			}

			buildState.IncrementBuildProgress()

			_, err = clients[i].DockerExecd(localId,
				fmt.Sprintf("bash -c 'while :;do artemis-log-parser --influx \"http://%s:8086\" --node \"%s%d\" /artemis/data/data.json 2>&1 >> /parser.log; done'",
					util.GetGateway(server.SubnetID, localId), conf.NodePrefix, node))
			if err != nil {
				log.Println(err)
				return nil, err
			}
			node++
		}
	}

	return nil, nil
}

func Add(details db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
	newNodes map[int][]string, buildState *state.BuildState) ([]string, error) {
	return nil, nil
}
