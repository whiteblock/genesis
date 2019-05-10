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

package geth

import (
	"encoding/json"
	"fmt"
	"github.com/Whiteblock/genesis/blockchains/helpers"
	"github.com/Whiteblock/genesis/util"
)

type ethConf struct {
	ExtraAccounts  int64  `json:"extraAccounts"`
	NetworkID      int64  `json:"networkId"`
	Difficulty     int64  `json:"difficulty"`
	InitBalance    string `json:"initBalance"`
	MaxPeers       int64  `json:"maxPeers"`
	GasLimit       int64  `json:"gasLimit"`
	HomesteadBlock int64  `json:"homesteadBlock"`
	Eip155Block    int64  `json:"eip155Block"`
	Eip158Block    int64  `json:"eip158Block"`
}

/**
 * Fills in the defaults for missing parts,
 */
func newConf(data map[string]interface{}) (*ethConf, error) {
	out := new(ethConf)
	err := json.Unmarshal([]byte(GetDefaults()), out)

	if data == nil {
		return out, nil
	}

	err = util.GetJSONInt64(data, "extraAccounts", &out.ExtraAccounts)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "networkId", &out.NetworkID)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "difficulty", &out.Difficulty)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "maxPeers", &out.MaxPeers)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "gasLimit", &out.GasLimit)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "eip155Block", &out.Eip155Block)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "homesteadBlock", &out.HomesteadBlock)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "eip158Block", &out.Eip158Block)
	if err != nil {
		return nil, err
	}

	initBalance, exists := data["initBalance"]
	if exists && initBalance != nil {
		switch initBalance.(type) {
		case json.Number:
			out.InitBalance = initBalance.(json.Number).String()
		case string:
			out.InitBalance = initBalance.(string)
		default:
			return nil, fmt.Errorf("incorrect type for initBalance given")
		}
	}

	return out, nil
}

// GetParams fetchs artemis related parameters
func GetParams() string {
	dat, err := helpers.GetStaticBlockchainConfig(blockchain, "params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetDefaults fetchs artemis related parameter defaults
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
			Name:    "ethNetStats",
			Image:   "gcr.io/whiteblock/ethnetstats:dev",
			Env:     nil,
			Network: "host",
		},
	}
}
