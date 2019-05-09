/*
	Copyright 2019 Whiteblock Inc.
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

package orion

import (
	"github.com/Whiteblock/genesis/blockchains/helpers"
	"github.com/Whiteblock/genesis/blockchains/registrar"
	"github.com/Whiteblock/genesis/db"
	"github.com/Whiteblock/genesis/ssh"
	"github.com/Whiteblock/genesis/testnet"
	"github.com/Whiteblock/genesis/util"
	"fmt"
	"github.com/Whiteblock/mustache"
)

var conf *util.Config

const sidecar = "orion"

func init() {
	conf = util.GetConfig()
	registrar.RegisterSideCar(sidecar, registrar.SideCar{
		Image: "gcr.io/whiteblock/orion:dev",
	})
	registrar.RegisterBuildSideCar(sidecar, build)
	registrar.RegisterAddSideCar(sidecar, add)
}

func build(tn *testnet.Adjunct) error {

	helpers.AllNodeExecConSC(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error { //ignore err
		_, err := client.DockerExec(node, "mkdir -p /orion/data")
		return err
	})

	err := helpers.CreateConfigsSC(tn, "/orion/data/orion.conf", func(node ssh.Node) ([]byte, error) {
		return makeNodeConfig(node)
	})
	if err != nil {
		return util.LogError(err)
	}

	err = helpers.AllNodeExecConSC(tn, func(client *ssh.Client, server *db.Server, node ssh.Node) error {
		_, err := client.DockerExec(node, "bash -c 'cd /orion/data && echo \"\" | orion -g nodeKey'")
		return err
	})
	if err != nil {
		return util.LogError(err)
	}
	ips := make([]string, len(tn.Nodes))
	for i, node := range tn.Nodes {
		ips[i] = node.GetIP()
	}
	tn.BuildState.SetExt("orion", ips)

	return helpers.AllNodeExecConSC(tn, func(client *ssh.Client, server *db.Server, node ssh.Node) error {
		return client.DockerExecdLog(node, "orion /orion/data/orion.conf")
	})
}

func add(tn *testnet.Adjunct) error {
	return nil
}

func makeNodeConfig(node ssh.Node) ([]byte, error) {

	dat, err := helpers.GetStaticBlockchainConfig(sidecar, "orion.conf.mustache")
	if err != nil {
		return nil, util.LogError(err)
	}
	data, err := mustache.Render(string(dat), util.ConvertToStringMap(map[string]interface{}{
		"nodeurl":   fmt.Sprintf("http://%s:8080/", node.GetIP()),
		"clienturl": fmt.Sprintf("http://%s:8080/", node.GetIP()),
	}))
	return []byte(data), err
}
