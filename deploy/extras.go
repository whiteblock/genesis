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
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/docker"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"sync"
)

func distributeNibbler(tn *testnet.TestNet) {
	if conf.DisableNibbler {
		log.Info("nibbler is disabled")
		return
	}
	tn.BuildState.Async(func() {
		var err error
		for i := uint(0); i < conf.NibblerRetries; i++ {
			var nibbler []byte
			nibbler, err = util.HTTPRequest("GET", conf.NibblerEndPoint, "")
			if err != nil {
				log.WithFields(log.Fields{"error": err, "attempt": i}).Error("failed to download nibbler. retrying...")
				continue
			}
			if nibbler == nil || len(nibbler) == 0 {
				log.WithFields(log.Fields{"error": err, "attempt": i}).Error("downloaded an empty nibbler")
				continue
			}

			err = tn.BuildState.Write("nibbler", string(nibbler))
			if err != nil {
				log.Error(err)
				continue
			}

			err = helpers.CopyToAllNewNodesDR(tn, "nibbler", "/usr/local/bin/nibbler")
			if err != nil {
				log.Error(err)
				continue
			}

			err = helpers.AllNewNodeExecConDR(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
				_, err := client.DockerExec(node, "chmod +x /usr/local/bin/nibbler")
				return err
			})
			if err != nil {
				log.Error(err)
				continue
			}
			break
		}
	})
}

func handleDockerBuildRequest(tn *testnet.TestNet, prebuild map[string]interface{}) error {
	if !conf.EnableImageBuilding {
		log.Warn("got a request to build an image, when it is disabled")
		return fmt.Errorf("image building is disabled")
	}
	_, hasDockerfile := prebuild["dockerfile"] //Must be base64
	if !hasDockerfile {
		return fmt.Errorf("cannot build without being given a dockerfile")
	}

	dockerfile, err := base64.StdEncoding.DecodeString(prebuild["dockerfile"].(string))
	if err != nil {
		return util.LogError(err)
	}
	err = tn.BuildState.Write("Dockerfile", string(dockerfile))
	if err != nil {
		return util.LogError(err)
	}

	dir, err := util.GetUUIDString()
	if err != nil {
		return util.LogError(err)
	}

	err = helpers.AllServerExecCon(tn, func(client ssh.Client, _ *db.Server) error {
		tn.BuildState.Defer(func() { client.Run(fmt.Sprintf("rm /tmp/%s/", dir)) })
		_, err := client.Run(fmt.Sprintf("mkdir /tmp/%s/", dir))
		return err
	})

	if err != nil {
		return util.LogError(err)
	}

	err = helpers.CopyAllToServers(tn, "Dockerfile", fmt.Sprintf("/tmp/%s/Dockerfile", dir))
	if err != nil {
		return util.LogError(err)
	}

	tag, err := util.GetUUIDString()
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.SetBuildStage("Building your custom image")
	imageName := fmt.Sprintf("%s:%s", tn.LDD.Blockchain, tag)
	wg := sync.WaitGroup{}
	for _, client := range tn.Clients {
		wg.Add(1)
		go func(client ssh.Client) {
			defer wg.Done()

			_, err := client.Run(fmt.Sprintf("docker build /tmp/%s -t %s", dir, imageName))
			tn.BuildState.Defer(func() { client.Run(fmt.Sprintf("docker rmi %s", imageName)) })
			if err != nil {
				tn.BuildState.ReportError(err)
				return
			}

		}(client)
	}
	wg.Wait()
	tn.UpdateAllImages(imageName)
	return util.LogError(tn.BuildState.GetError())
}

func handleDockerAuth(tn *testnet.TestNet, auth map[string]interface{}) error {
	wg := sync.WaitGroup{}
	for _, client := range tn.Clients {
		wg.Add(1)
		go func(client ssh.Client) { //TODO add validation
			err := docker.Login(client, auth["username"].(string), auth["password"].(string))
			if err != nil {
				tn.BuildState.ReportError(err)
			}
			tn.BuildState.Defer(func() { docker.Logout(client) })
		}(client)
	}

	wg.Wait()
	return util.LogError(tn.BuildState.GetError())
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
			return util.LogError(err)
		}
	}
	//Handle docker build
	dockerBuild, ok := prebuild["build"] //bool to see if a manual build was requested.
	if ok && dockerBuild.(bool) {
		err := handleDockerBuildRequest(tn, prebuild)
		if err != nil {
			return util.LogError(err)
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
					tn.BuildState.ReportError(err)
					return
				}
			}(image)
		}
		wg.Wait()
	}

	return tn.BuildState.GetError()
}
