/*
	Copyright 2019 whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Genesis is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package deploy

import (
	"encoding/base64"
	"fmt"
	"github.com/whiteblock/genesis/blockchains/helpers"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/docker"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"log"
	"sync"
)

func distributeNibbler(tn *testnet.TestNet) {
	tn.BuildState.Async(func() {
		nibbler, err := util.HTTPRequest("GET", "https://storage.googleapis.com/genesis-public/nibbler/dev/bin/linux/amd64/nibbler", "")
		//nibbler, err := util.HTTPRequest("GET", "http://127.0.0.1/nibbler", "")
		if err != nil {
			log.Println(err)
		}
		err = tn.BuildState.Write("nibbler", string(nibbler))
		if err != nil {
			log.Println(err)
		}
		err = helpers.CopyToAllNewNodes(tn, "nibbler", "/usr/local/bin/nibbler")
		if err != nil {
			log.Println(err)
		}
		err = helpers.AllNewNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
			_, err := client.DockerExec(node, "chmod +x /usr/local/bin/nibbler")
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
		return fmt.Errorf("cannot build without being given a dockerfile")
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

	err = helpers.CopyAllToServers(tn, "Dockerfile", "~/Dockerfile")
	if err != nil {
		log.Println(err)
		return err
	}

	tag := util.GetUUIDString()

	tn.BuildState.SetBuildStage("Building your custom image")
	imageName := fmt.Sprintf("%s:%s", tn.LDD.Blockchain, tag)
	wg := sync.WaitGroup{}
	for _, client := range tn.Clients {
		wg.Add(1)
		go func(client *ssh.Client) {
			defer wg.Done()

			_, err := client.Run(fmt.Sprintf("docker build ~ -t %s", imageName))
			tn.BuildState.Defer(func() { client.Run(fmt.Sprintf("docker rmi %s", imageName)) })
			if err != nil {
				log.Println(err)
				tn.BuildState.ReportError(err)
				return
			}

		}(client)
	}
	wg.Wait()
	return tn.BuildState.GetError()
}

func handleDockerAuth(tn *testnet.TestNet, auth map[string]interface{}) error {
	wg := sync.WaitGroup{}
	for _, client := range tn.Clients {
		wg.Add(1)
		go func(client *ssh.Client) { //TODO add validation
			err := docker.Login(client, auth["username"].(string), auth["password"].(string))
			if err != nil {
				log.Println(err)
				tn.BuildState.ReportError(err)
			}
			tn.BuildState.Defer(func() { docker.Logout(client) })
		}(client)
	}

	wg.Wait()
	return tn.BuildState.GetError()
}

func handlePreBuildExtras(tn *testnet.TestNet) error {
	if tn.LDD.Extras == nil {
		return nil //Nothing to do
	}
	_, exists := tn.LDD.Extras["prebuild"]
	if !exists {
		return nil //Nothing to do
	}
	prebuild, ok := tn.LDD.Extras["prebuild"].(map[string]interface{})
	if !ok || prebuild == nil {
		return nil //Nothing to do
	}
	//Handle docker Auth
	iDockerAuth, ok := prebuild["auth"]
	if ok {
		err := handleDockerAuth(tn, iDockerAuth.(map[string]interface{}))
		if err != nil {
			log.Println(err)
			return err
		}
	}
	//Handle docker build
	dockerBuild, ok := prebuild["build"] //bool to see if a manual build was requested.
	if ok && dockerBuild.(bool) {
		err := handleDockerBuildRequest(tn, prebuild)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	//Force docker pull
	dockerPull, ok := prebuild["pull"]
	if ok && dockerPull.(bool) { //Slightly frail
		tn.BuildState.SetBuildStage("Pulling your images")
		wg := sync.WaitGroup{}
		images := util.GetUniqueStrings(tn.LDD.Images)
		for _, image := range images {
			wg.Add(1)
			go func(image string) {
				defer wg.Done()
				err := docker.Pull(tn.GetFlatClients(), image) //OPTMZ
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
