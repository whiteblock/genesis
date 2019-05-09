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

package rchain

import (
	"github.com/Whiteblock/genesis/util"
	"github.com/Whiteblock/genesis/blockchains/helpers"
	"encoding/base64"
	"encoding/json"
	"log"
)

type rChainConf struct {
	NoUpnp               bool   `json:"noUpnp"`
	DefaultTimeout       int64  `json:"defaultTimeout"`
	MapSize              int64  `json:"mapSize"`
	CasperBlockStoreSize int64  `json:"casperBlockStoreSize"`
	InMemoryStore        bool   `json:"inMemoryStore"`
	MaxNumOfConnections  int64  `json:"maxNumOfConnections"`
	Validators           int64  `json:"validators"`
	ValidatorCount       int64  `json:"validatorCount"`
	SigAlgorithm         string `json:"sigAlgorithm"`
	Command              string `json:"command"`
	BondsValue           int64  `json:"bondsValue"`
}

func newRChainConf(data map[string]interface{}) (*rChainConf, error) {
	out := new(rChainConf)
	err := json.Unmarshal([]byte(GetDefaults()), out)
	if data == nil {
		return out, util.LogError(err)
	}
	log.Printf("Default %+v\n", *out)
	tmp, err := json.Marshal(data)
	if err != nil {
		return nil, util.LogError(err)
	}
	return out, json.Unmarshal(tmp, out)
}

// GetServices returns the services which are used by rchain
func GetServices() []util.Service {
	return []util.Service{
		{
			Name:  "wb_influx_proxy",
			Image: "gcr.io/wb-genesis/bitbucket.org/whiteblockio/influx-proxy:master",
			Env: map[string]string{
				"BASIC_AUTH_BASE64": base64.StdEncoding.EncodeToString([]byte(conf.InfluxUser + ":" + conf.InfluxPassword)),
				"INFLUXDB_URL":      conf.Influx,
				"BIND_PORT":         "8086",
			},
		},
	}
}

// GetParams fetchs rchain related parameters
func GetParams() string {
	dat, err := helpers.GetStaticBlockchainConfig(blockchain, "params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetDefaults fetchs rchain related parameter defaults
func GetDefaults() string {
	dat, err := helpers.GetStaticBlockchainConfig(blockchain, "defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}
