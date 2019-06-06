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
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
)

func handleExtraPublicKeys(tn *testnet.TestNet) error {
	postBuild, ok := util.ExtractStringMap(tn.LDD.Extras, "postbuild")
	if !ok {
		return nil
	}
	log.WithFields(log.Fields{"postBuild": tn.LDD.Extras["postbuild"]}).Trace("extracted post build details")
	SSHDetails, ok := util.ExtractStringMap(postBuild, "ssh")
	if !ok || SSHDetails == nil {
		return nil
	}
	iPubKeys, ok := SSHDetails["pubKeys"]
	if !ok || iPubKeys == nil {
		return nil
	}
	pubKeys := iPubKeys.([]interface{})

	tn.BuildState.Async(func() {
		helpers.AllNewNodeExecCon(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error {
			for i := range pubKeys {
				_, err := client.DockerExec(node, fmt.Sprintf(`bash -c 'echo "%v" >> /root/.ssh/authorized_keys'`, pubKeys[i]))
				if err != nil {
					return util.LogError(err)
				}
			}
			return nil
		})
	})
	return nil
}

func handlePostBuild(tn *testnet.TestNet) error {
	return handleExtraPublicKeys(tn)
	//return nil
}
