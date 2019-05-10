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

package beam

import (
	"github.com/Whiteblock/genesis/blockchains/helpers"
	"github.com/Whiteblock/genesis/db"
	"github.com/Whiteblock/genesis/util"
	"github.com/Whiteblock/mustache"
	"io/ioutil"
)

type beamConf struct {
	Validators int64 `json:"validators"`
	TxNodes    int64 `json:"txNodes"`
	NilNodes   int64 `json:"nilNodes"`
}

func newConf(data map[string]interface{}) (*beamConf, error) {
	out := new(beamConf)

	err := util.GetJSONInt64(data, "validators", &out.Validators)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "txNodes", &out.TxNodes)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "nilNodes", &out.NilNodes)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetParams fetchs beam related parameters
func GetParams() string {
	dat, err := ioutil.ReadFile("./resources/beam/params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetDefaults fetchs beam related parameter defaults
func GetDefaults() string {
	dat, err := ioutil.ReadFile("./resources/beam/defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetServices returns the services which are used by artemis
func GetServices() []util.Service {
	return nil
}

func makeNodeConfig(bconf *beamConf, keyOwner string, keyMine string, details *db.DeploymentDetails, node int) (string, error) {

	filler := util.ConvertToStringMap(map[string]interface{}{
		"keyOwner": keyOwner,
		"keyMine":  keyMine,
	})
	dat, err := helpers.GetBlockchainConfig("beam", node, "beam-node.cfg.mustache", details)
	if err != nil {
		return "", err
	}
	data, err := mustache.Render(string(dat), filler)
	return data, err
}
