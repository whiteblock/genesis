package deploy

import (
	helpers "../blockchains/helpers"
	ssh "../ssh"
	testnet "../testnet"
	util "../util"
	"encoding/base64"
	"fmt"
	"log"
	"sync"
)

func distributeNibbler(tn *testnet.TestNet) {
	tn.BuildState.Async(func() {
		nibbler, err := util.HttpRequest("GET", "https://storage.googleapis.com/genesis-public/nibbler/master/bin/linux/amd64/nibbler", "")
		if err != nil {
			log.Println(err)
		}
		err = tn.BuildState.Write("nibbler", string(nibbler))
		if err != nil {
			log.Println(err)
		}
		err = helpers.CopyToAllNodes(tn, "nibbler", "/usr/local/bin/nibbler")
		if err != nil {
			log.Println(err)
		}
		err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ int, localNodeNum int, _ int) error {
			_, err := client.DockerExec(localNodeNum, "chmod +x /usr/local/bin/nibbler")
			return err
		})
		if err != nil {
			log.Println(err)
		}
	})
}

func handleDockerBuildRequest(tn *testnet.TestNet, prebuild map[string]interface{}) error {

	_, hasDockerfile := prebuild["dockerfile"] //Must be base64
	if !hasDockerfile {
		return fmt.Errorf("Cannot build without being given a dockerfile")
	}

	dockerfile, err := base64.StdEncoding.DecodeString(prebuild["dockerfile"].(string))
	if err != nil {
		log.Println(err)
		return err
	}
	err = tn.BuildState.Write("Dockerfile", string(dockerfile))
	if err != nil {
		log.Println(err)
	}

	err = helpers.CopyAllToServers(tn, "Dockerfile", "/home/appo/Dockerfile")
	if err != nil {
		log.Println(err)
		return err
	}

	tag, err := util.GetUUIDString()
	if err != nil {
		log.Println(err)
		return err
	}
	tn.BuildState.SetBuildStage("Building your custom image")
	imageName := fmt.Sprintf("%s:%s", tn.LDD().Blockchain, tag)
	wg := sync.WaitGroup{}
	for _, client := range tn.Clients {
		wg.Add(1)
		go func(client *ssh.Client) {
			defer wg.Done()

			_, err := client.Run(fmt.Sprintf("docker build /home/appo/ -t %s", imageName))
			tn.BuildState.Defer(func() { client.Run(fmt.Sprintf("docker rmi %s", imageName)) })
			if err != nil {
				log.Println(err)
				tn.BuildState.ReportError(err)
				return
			}

		}(client)
	}
	wg.Wait()
	if !tn.BuildState.ErrorFree() {
		return tn.BuildState.GetError()
	}
	return nil
}

func handlePreBuildExtras(tn *testnet.TestNet) error {
	if tn.LDD().Extras == nil {
		return nil //Nothing to do
	}
	_, exists := tn.LDD().Extras["prebuild"]
	if !exists {
		return nil //Nothing to do
	}
	prebuild, ok := tn.LDD().Extras["prebuild"].(map[string]interface{})
	if !ok || prebuild == nil {
		return nil //Nothing to do
	}
	//

	dockerBuild, ok := prebuild["build"] //bool to see if a manual build was requested.
	if ok && dockerBuild.(bool) {
		err := handleDockerBuildRequest(tn, prebuild)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	dockerPull, ok := prebuild["pull"]
	if ok && dockerPull.(bool) {
		wg := sync.WaitGroup{}
		for _, image := range tn.LDD().Images { //OPTMZ
			wg.Add(1)
			go func(image string) {
				defer wg.Done()
				err := DockerPull(tn.GetFlatClients(), image) //OPTMZ
				if err != nil {
					log.Println(err)
					tn.BuildState.ReportError(err)
					return
				}
			}(image)
		}
		wg.Wait()
	}

	return tn.BuildState.GetError()
}
