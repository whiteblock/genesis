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

package pantheon

import (
	"github.com/whiteblock/genesis/blockchains/helpers"
	"github.com/whiteblock/genesis/util"
)

type panConf struct {
	NetworkID             int64  `json:"networkId"`
	Difficulty            int64  `json:"difficulty"`
	InitBalance           string `json:"initBalance"`
	MaxPeers              int64  `json:"maxPeers"`
	GasLimit              int64  `json:"gasLimit"`
	Consensus             string `json:"consensus"`
	FixedDifficulty       int64  `json:"fixedDifficulty"`
	BlockPeriodSeconds    int64  `json:"blockPeriodSeconds"`
	Epoch                 int64  `json:"epoch"`
	RequestTimeoutSeconds int64  `json:"requesttimeoutseconds"`
	Accounts              int64  `json:"accounts"`
	Orion                 bool   `json:"orion"`
}

/**
 * Fills in the defaults for missing parts,
 */
func newConf(data map[string]interface{}) (*panConf, error) {
	out := new(panConf)
	return out, helpers.HandleBlockchainConfig(blockchain, data, out)
}

// GetServices returns the services which are used by artemis
func GetServices() []util.Service {
	return []util.Service{
		{ //Include a geth node for transaction signing
			Name:  "geth",
			Image: "gcr.io/whiteblock/geth:master",
			Env:   nil,
		},
	}
}
