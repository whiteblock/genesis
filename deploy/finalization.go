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
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/ssh"
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

	handlePostBuild(tn)
	return nil
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
