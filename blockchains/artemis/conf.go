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
	"encoding/json"
	"fmt"
	"github.com/whiteblock/genesis/blockchains/helpers"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/util"
	"github.com/whiteblock/mustache"
	"reflect"
)

type artemisConf map[string]interface{}

func newConf(data map[string]interface{}) (artemisConf, error) {
	rawDefaults := helpers.DefaultGetDefaultsFn(blockchain)()
	defaults := map[string]interface{}{}

	err := json.Unmarshal([]byte(rawDefaults), &defaults)
	if err != nil {
		return nil, util.LogError(err)
	}
	finalData := util.MergeStringMaps(defaults, data)
	var val int64
	err = util.GetJSONInt64(finalData, "validators", &val) //Check provided validators
	if err == nil {
		if val < 4 || val%2 != 0 {
			return nil, fmt.Errorf("invalid number of validators (%d): must be an even number and greater than 3", val)
		}
	}
	out := new(artemisConf)
	*out = artemisConf(finalData)

	return *out, nil
}

// GetServices returns the services which are used by artemis
func GetServices() []helpers.Service {
	return []helpers.Service{
		helpers.RegisterPrometheus(),
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
	var outputFile string
	obj := details.Params["outputFile"]
	if obj != nil && reflect.TypeOf(obj).Kind() == reflect.String {
		outputFile = obj.(string)
	}
	if outputFile == "" {
		outputFile = "/artemis/data/log.json"
	}
	filler["outputFile"] = outputFile
	var providerType string
	obj = details.Params["providerType"]
	if obj != nil && reflect.TypeOf(obj).Kind() == reflect.String {
		providerType = obj.(string)
	}
	if providerType == "" {
		providerType = "JSON"
	}
	var prometheusInstrumentationPort string
	obj = details.Params["prometheusInstrumentationPort"]
	if obj != nil && reflect.TypeOf(obj).Kind() == reflect.String {
		prometheusInstrumentationPort = obj.(string)
	}
	if prometheusInstrumentationPort == "" {
		prometheusInstrumentationPort = "8088"
	}

	filler["providerType"] = providerType
	filler["metricsPort"] = prometheusInstrumentationPort
	filler["constants"] = constantsRaw

	filler["validators"] = fmt.Sprintf("%.0f", aconf["validators"])
	dat, err := helpers.GetBlockchainConfig("artemis", node, "artemis-config.toml.mustache", details)
	if err != nil {
		return "", util.LogError(err)
	}
	return mustache.Render(string(dat), filler)
}
