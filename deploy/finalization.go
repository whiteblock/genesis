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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/registrar"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/status"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
)

/*
   Finalization methods for the docker build process. Will be run immediately following their deployment
*/
func finalize(tn *testnet.TestNet) error {
	if conf.HandleNodeSSHKeys {
		err := copyOverSSHKeys(tn, false)
		if err != nil {
			return util.LogError(err)
		}
	}
	alwaysRunFinalize(tn)
	handlePostBuild(tn)
	return nil
}

/*
   Finalization methods for the docker build process. Will be run immediately following their deployment
*/
func finalizeNewNodes(tn *testnet.TestNet) error {
	if conf.HandleNodeSSHKeys {
		err := copyOverSSHKeys(tn, true)
		if err != nil {
			return util.LogError(err)
		}
	}
	alwaysRunFinalize(tn)
	handlePostBuild(tn)
	return nil
}

func alwaysRunFinalize(tn *testnet.TestNet) {

	tn.BuildState.Async(func() {
		for _, node := range tn.NewlyBuiltNodes {
			err := declareNode(&node, tn)
			if err != nil {
				log.WithFields(log.Fields{"node": node.AbsoluteNum}).Error(err)
			}
		}
	})
	newNodes := make([]db.Node, len(tn.NewlyBuiltNodes))
	copy(newNodes, tn.NewlyBuiltNodes)
	tn.BuildState.Defer(func() {
		for i, node := range newNodes {
			err := finalizeNode(node, tn.LDD, i)
			if err != nil {
				tn.BuildState.ReportError(err)
			}
		}
	})
}

/*
   Copy over the ssh public key to each node to allow for the user to ssh into each node.
   The public key comes from the nodes public key specified in the configuration
*/
func copyOverSSHKeys(tn *testnet.TestNet, newOnly bool) error {
	tmp, err := ioutil.ReadFile(conf.NodesPublicKey)
	if err != nil {
		log.WithFields(log.Fields{"loc": conf.NodesPublicKey, "error": err}).Error("failed to read the public key file")
		return util.LogError(err)
	}
	pubKey := string(tmp)
	pubKey = strings.Trim(pubKey, "\t\n\v\r")

	privKey, err := ioutil.ReadFile(conf.NodesPrivateKey)
	if err != nil {
		return util.LogError(err)
	}

	fn := func(client ssh.Client, _ *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementDeployProgress()

		_, err := client.DockerExec(node, "mkdir -p /root/.ssh/")
		if err != nil {
			return util.LogError(err)
		}
		_, err = client.DockerExec(node, fmt.Sprintf(`sh -c 'echo "%s" >> /root/.ssh/authorized_keys'`, pubKey))
		if err != nil {
			return util.LogError(err)
		}

		_, err = client.DockerExecd(node, "sh -c '/sbin/apk update && /sbin/apk add --no-cache openrc openssh && ssh-keygen -A && /usr/sbin/sshd || true'")
		if err != nil {
			log.Warn(err)
			_, err = client.DockerExecd(node, "sh -c 'apt-get update && apt-get install openssh-server && service ssh start || true'")
			if err != nil {
				log.Warn(err)
			}
		}
		return nil
	}

	if newOnly {
		err = helpers.AllNewNodeExecCon(tn, fn)
	} else {
		err = helpers.AllNodeExecCon(tn, fn)
	}
	if err != nil {
		return util.LogError(err)
	}
	if newOnly {
		return helpers.CopyBytesToAllNewNodes(tn, string(privKey), "/root/.ssh/id_rsa")
	}
	return helpers.CopyBytesToAllNodes(tn, string(privKey), "/root/.ssh/id_rsa")
}

func declareNode(node *db.Node, tn *testnet.TestNet) error {
	if conf.DisableTestnetReporting {
		log.Info("skipping node declaration since testnet reporting is disabled")
		return nil
	}
	if len(tn.LDD.GetJwt()) == 0 { //If there isn't a JWT, return immediately
		return nil
	}
	image := tn.LDD.Images[0]

	if len(tn.LDD.Images) > node.AbsoluteNum {
		image = tn.LDD.Images[node.AbsoluteNum]
	}

	data := map[string]interface{}{
		"id":         node.ID,
		"ip_address": node.IP,
		"image":      image,
		"kind":       tn.LDD.Blockchain,
		"version":    "unknown",
	}
	rawData, err := json.Marshal(data)
	if err != nil {
		return util.LogError(err)
	}

	_, err = util.JwtHTTPRequest("POST", conf.APIEndpoint+"/testnets/"+node.TestNetID+"/nodes", tn.LDD.GetJwt(), string(rawData))
	return err
}

func finalizeNode(node db.Node, details *db.DeploymentDetails, absNum int) error {
	if conf.DisableNibbler {
		log.Info("skipping nibbler setup as it is disabled")
		return nil
	}
	client, err := status.GetClient(node.Server)
	if err != nil {
		return util.LogError(err)
	}
	files := details.Blockchain + " " + conf.DockerOutputFile
	if details.Logs != nil && len(details.Logs) > 0 {
		var logFiles map[string]string
		if len(details.Logs) == 1 || len(details.Logs) <= absNum {
			logFiles = details.Logs[0]
		} else {
			logFiles = details.Logs[absNum]
		}
		for name, file := range logFiles { //Eventually may need to handle the names as well
			files += " " + name + " " + file
		}
	}
	logFiles := registrar.GetAdditionalLogs(details.Blockchain)
	for name, logFile := range logFiles {
		files += " " + name + " " + logFile
	}

	_, err = client.DockerExecd(node,
		fmt.Sprintf("sh -c 'nibbler --node-type %s --api %s --jwt %s --testnet %s --node %s %s 2>&1 >> /nibbler.log'",
			details.Blockchain, conf.APIEndpoint, details.GetJwt(), node.TestNetID, node.ID, files))
	return util.LogError(err)
}
