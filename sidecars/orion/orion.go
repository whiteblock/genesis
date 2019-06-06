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

package orion

import (
	"fmt"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/registrar"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"github.com/whiteblock/mustache"
)

var conf *util.Config

const sidecar = "orion"

func init() {
	conf = util.GetConfig()
	registrar.RegisterSideCar(sidecar, registrar.SideCar{
		Image: "gcr.io/whiteblock/orion:dev",
		BuildStepsCalc: func(nodes int, _ int) int {
			return 4 * nodes
		},
	})
	registrar.RegisterBuildSideCar(sidecar, build)
	registrar.RegisterAddSideCar(sidecar, add)
}

func build(tn *testnet.Adjunct) error {

	helpers.AllNodeExecConSC(tn, func(client ssh.Client, _ *db.Server, node ssh.Node) error { //ignore err
		defer tn.BuildState.IncrementSideCarProgress()
		_, err := client.DockerExec(node, "mkdir -p /orion/data")
		return err
	})

	err := helpers.CreateConfigsSC(tn, "/orion/data/orion.conf", func(node ssh.Node) ([]byte, error) {
		defer tn.BuildState.IncrementSideCarProgress()
		return makeNodeConfig(node)
	})
	if err != nil {
		return util.LogError(err)
	}

	err = helpers.AllNodeExecConSC(tn, func(client ssh.Client, server *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementSideCarProgress()
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

	return helpers.AllNodeExecConSC(tn, func(client ssh.Client, server *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementSideCarProgress()
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
