package artemis

import (
	"fmt"
	"log"
	db "../../db"
	util "../../util"
	state "../../state"
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
	buildState.SetBuildSteps(0+(details.Nodes*4))

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
			identity := fmt.Sprintf("0x0%x", j)

			artemisNodeConfig,err := makeNodeConfig(artemisConf, identity, peers, details.Nodes, details.Params) 
			if err != nil {
				log.Println(err)
				return nil, err
			}

			fmt.Println("Writing Configuration File")
			err = util.Write("config.toml", artemisNodeConfig)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			defer util.Rm("./config.toml")

			err = clients[i].Scp("./config.toml", "/home/appo/config.toml")
			if err != nil {
				log.Println(err)
				return nil, err
			}
			defer clients[i].Run("rm -f /home/appo/config.toml")

			err = clients[i].DockerCp(j,"/home/appo/config.toml","/artemis/config/config.toml")
			if err != nil {
				log.Println(err)
				return nil,err
			}
		}
	}

	buildState.SetBuildStage("Starting Artemis")
	for i, server := range servers {
		for localId, _ := range server.Ips {
			artemisCmd := fmt.Sprintf(
				`artemis -c /artemis/config/config.toml -o /artemis/data/data.json 2>&1 | tee /output.log`,
			)
			clients[i].DockerExecd(localId,"tmux new -s whiteblock -d")
			clients[i].DockerExecd(localId,fmt.Sprintf("tmux send-keys -t whiteblock '%s' C-m",artemisCmd))
			buildState.IncrementBuildProgress()
		}
	}

	return nil, nil
}

func Add(details db.DeploymentDetails,servers []db.Server,clients []*util.SshClient,
	newNodes map[int][]string,buildState *state.BuildState) ([]string,error) {
	return nil,nil
}

