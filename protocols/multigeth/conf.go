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

// MgethConf represents the settings for the multi-geth build
type MgethConf struct {
	Network               string `json:"network"`
	ExtraAccounts         int64  `json:"extraAccounts"`
	NetworkID             int64  `json:"networkId"`
	Difficulty            int64  `json:"difficulty"`
	InitBalance           string `json:"initBalance"`
	MaxPeers              int64  `json:"maxPeers"`
	GasLimit              int64  `json:"gasLimit"`
	ExtraData             string `json:"extraData"`
	Consensus             string `json:"consensus"`
	BlockPeriodSeconds    int64  `json:"blockPeriodSeconds"`
	Epoch                 int64  `json:"epoch"`
	Mode                  string `json:"mode"`
	Verbosity             int64  `json:"verbosity"`
	Unlock                bool   `json:"unlock"`
	HomesteadBlock        int64  `json:"homesteadBlock"`
	EIP7FBlock            int64  `json:"eip7FBlock"`
	EIP150Block           int64  `json:"eip150Block"`
	EIP155Block           int64  `json:"eip155Block"`
	EIP158Block           int64  `json:"eip158Block"`
	ByzantiumBlock        int64  `json:"byzantiumBlock"`
	DisposalBlock         int64  `json:"disposalBlock"`
	ConstantinopleBlock   int64  `json:"constantinopleBlock"`
	ECIP1017EraRounds     int64  `json:"ecip1017EraRounds"`
	EIP160FBlock           int64  `json:"eip160FBlock"`
}

/**
 * Fills in the defaults for missing parts,
 */
func newConf(data map[string]interface{}) (*MgethConf, error) {
	out := new(MgethConf)
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
