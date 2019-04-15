package deploy

import (
	helpers "../blockchains/helpers"
	db "../db"
	state "../state"
	util "../util"
	"encoding/base64"
	"fmt"
	"log"
	"sync"
)

func handleDockerBuildRequest(blockchain string, prebuild map[string]interface{},
	clients []*util.SshClient, buildState *state.BuildState) error {

	_, hasDockerfile := prebuild["dockerfile"] //Must be base64
	if !hasDockerfile {
		return fmt.Errorf("Cannot build without being given a dockerfile")
	}

	dockerfile, err := base64.StdEncoding.DecodeString(prebuild["dockerfile"].(string))
	if err != nil {
		log.Println(err)
		return err
	}
	err = buildState.Write("Dockerfile", string(dockerfile))
	if err != nil {
		log.Println(err)
	}

	err = helpers.CopyAllToServers(clients, buildState, "Dockerfile", "/home/appo/Dockerfile")
	if err != nil {
		log.Println(err)
		return err
	}

	tag, err := util.GetUUIDString()
	if err != nil {
		log.Println(err)
		return err
	}
	buildState.SetBuildStage("Building your custom image")
	imageName := fmt.Sprintf("%s:%s", blockchain, tag)
	wg := sync.WaitGroup{}
	for _, client := range clients {
		wg.Add(1)
		go func(client *util.SshClient) {
			defer wg.Done()

			_, err := client.Run(fmt.Sprintf("docker build /home/appo/ -t %s", imageName))
			buildState.Defer(func() { client.Run(fmt.Sprintf("docker rmi %s", imageName)) })
			if err != nil {
				log.Println(err)
				buildState.ReportError(err)
				return
			}

		}(client)
	}
	wg.Wait()
	if !buildState.ErrorFree() {
		return buildState.GetError()
	}
	return nil
}

func handlePreBuildExtras(buildConf *db.DeploymentDetails, clients []*util.SshClient, buildState *state.BuildState) error {
	if buildConf.Extras == nil {
		return nil //Nothing to do
	}
	_, exists := buildConf.Extras["prebuild"]
	if !exists {
		return nil //Nothing to do
	}
	prebuild, ok := buildConf.Extras["prebuild"].(map[string]interface{})
	if !ok || prebuild == nil {
		return nil //Nothing to do
	}
	//

	dockerBuild, ok := prebuild["build"] //bool to see if a manual build was requested.
	if ok && dockerBuild.(bool) {
		err := handleDockerBuildRequest(buildConf.Blockchain, prebuild, clients, buildState)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	dockerPull, ok := prebuild["pull"]
	if ok && dockerPull.(bool) {
		err := DockerPull(clients, buildConf.Image)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}
