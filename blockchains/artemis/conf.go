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

package artemis

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/whiteblock/genesis/blockchains/helpers"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/util"
	"github.com/whiteblock/mustache"
)

type artemisConf map[string]interface{}

func newConf(data map[string]interface{}) (artemisConf, error) {
	rawDefaults := GetDefaults()
	defaults := map[string]interface{}{}

	err := json.Unmarshal([]byte(rawDefaults), &defaults)
	if err != nil {
		return nil, util.LogError(err)
	}
	var val int64
	err = util.GetJSONInt64(data, "validators", &val) //Check provided validators
	if err == nil {
		if val < 4 || val%2 != 0 {
			return nil, fmt.Errorf("invalid number of validators (%d): must be an even number and greater than 3", val)
		}
	}
	out := new(artemisConf)
	*out = artemisConf(util.MergeStringMaps(defaults, data))

	return *out, nil
}

// GetDefaults fetches artemis related parameter defaults
func GetDefaults() string {
	dat, err := helpers.GetStaticBlockchainConfig(blockchain, "defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetServices returns the services which are used by artemis
func GetServices() []util.Service {
	return []util.Service{
		{
			Name:  "wb_influx_proxy",
			Image: "gcr.io/whiteblock/influx-proxy:master",
			Env: map[string]string{
				"BASIC_AUTH_BASE64": base64.StdEncoding.EncodeToString([]byte(conf.InfluxUser + ":" + conf.InfluxPassword)),
				"INFLUXDB_URL":      conf.Influx,
				"BIND_PORT":         "8086",
			},
		},
	}
}

func makeNodeConfig(aconf artemisConf, identity string, peers string, node int, details *db.DeploymentDetails, constantsRaw string) (string, error) {

	artConf, err := util.CopyMap(aconf)
	if err != nil {
		return "", util.LogError(err)
	}
	artConf["identity"] = identity
	filler := util.ConvertToStringMap(artConf)
	filler["peers"] = peers
	filler["numNodes"] = fmt.Sprintf("%d", details.Nodes)
	filler["constants"] = constantsRaw
	var validators int64
	err = util.GetJSONInt64(details.Params, "validators", &validators)
	if err != nil {
		return "", util.LogError(err)
	}

	filler["validators"] = fmt.Sprintf("%d", validators)
	dat, err := helpers.GetBlockchainConfig("artemis", node, "artemis-config.toml.mustache", details)
	if err != nil {
		return "", util.LogError(err)
	}
	return mustache.Render(string(dat), filler)
}
