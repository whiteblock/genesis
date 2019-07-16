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

package rchain

import (
	"encoding/base64"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/services"
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
	return out, helpers.HandleBlockchainConfig(blockchain, data, out)
}

// GetServices returns the services which are used by rchain
func GetServices() []services.Service {
	return []services.Service{
		services.SimpleService{
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
