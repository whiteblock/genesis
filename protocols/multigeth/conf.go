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

package multigeth

import (
	"encoding/json"
	"fmt"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/services"
)

type mgethConf struct {
	Network            string `json:"network"`
}

type ethConf struct {
	Network               string `json:"network"`
	ExtraAccounts         int64  `json:"extraAccounts"`
	NetworkID             int64  `json:"networkId"`
	Difficulty            int64  `json:"difficulty"`
	InitBalance           string `json:"initBalance"`
	MaxPeers              int64  `json:"maxPeers"`
	GasLimit              int64  `json:"gasLimit"`
	Consensus             string `json:"consensus"`
	BlockPeriodSeconds    int64  `json:"blockPeriodSeconds"`
	Epoch                 int64  `json:"epoch"`

	HomesteadBlock        int64  `json:"homesteadBlock"`
	EthEip155Block        int64  `json:"eip155Block"`
	EthEip158Block        int64  `json:"eip158Block"`
	EtcIdentity           string `json:"identity"`
	EtcName               string `json:"name"`
	EtcEIP150Block        int64  `json:"eip150Block"`
	EtcDAOHFBlock         int64  `json:"daoHFBlock"`
	EtcEIP155_160Block    int64  `json:"eip155_160Block"`
	EtcECIP1010Length     int64  `json:"ecip1010Length"`
	EtcECIP1017Block      int64  `json:"ecip1017Block"`
	EtcECIP1017Era        int64  `json:"ecip1017Era"`
}

/**
 * Fills in the defaults for missing parts,
 */
func newConf(data map[string]interface{}) (*ethConf, error) {
	out := new(ethConf)
	err := helpers.HandleBlockchainConfig(blockchain, data, out)
	if err != nil || data == nil {
		return out, err
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

// GetServices returns the services which are used by artemis
func GetServices() []services.Service {
	return []services.Service{
		services.SimpleService{
			Name:    "ethNetStats",
			Image:   "gcr.io/whiteblock/ethnetstats:dev",
			Env:     nil,
			Network: "host",
		},
	}
}
