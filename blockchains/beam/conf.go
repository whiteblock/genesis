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

package beam

import (
	"github.com/whiteblock/genesis/blockchains/helpers"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/util"
	"github.com/whiteblock/mustache"
)

type beamConf struct {
	Validators int64 `json:"validators"`
	TxNodes    int64 `json:"txNodes"`
	NilNodes   int64 `json:"nilNodes"`
}

func newConf(data map[string]interface{}) (*beamConf, error) {
	out := new(beamConf)
	return out, helpers.HandleBlockchainConfig(blockchain, data, out)
}

// GetServices returns the services which are used by artemis
func GetServices() []helpers.Service {
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
