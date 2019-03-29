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

	buildState.SetBuildStage("Creating node configuration files")
	/**Create node config files**/
	
	for i, server := range servers {
		identity := "0x"
		if i < 10 {
			identity = identity + "0"
		}

		artemisNodeConfig,err := makeNodeConfig(artemisConf, identity) 
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

		for j, _ := range server.Ips {
			buildState.IncrementBuildProgress()

			err = clients[i].DockerCp(j,"/home/appo/config.toml","/artemis/config/config.toml")
			if err != nil {
				log.Println(err)
				return nil,err
			}
		}
	}
	return nil, nil
}

func Add(details db.DeploymentDetails,servers []db.Server,clients []*util.SshClient,
	newNodes map[int][]string,buildState *state.BuildState) ([]string,error) {
	return nil,nil
}