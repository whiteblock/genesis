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

package polkadot

import (
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/protocols/services"
)

type dotConf struct {
	ValidatorMode           bool   `json:"validatorMode"`
	InPeers                 int64  `json:"inPeers"`
	ListenAddr              string `json:"listenAddr"`
	Log                     string `json:"log"`
	OffchainWorker          string `json:"offChainWorker"`
	OffchainWorkerExecution string `json:"offChainWokerExecution"`
	OtherExecution          string `json:"otherExecution"`
	OutPeers                int64  `json:"outPeers"`
	PoolKbytes              int64  `json:"poolKbytes"`
	PoolLimit               int64  `json:"poolLimit"`
	Pruning                 int64  `json:"pruning"`
	StateCacheSize          int64  `json:"stateCacheSize"`
	TelemetryURL            int64  `json:"telemetryUrl"`
}

/**
 * Fills in the defaults for missing parts,
 */
func newConf(data map[string]interface{}) (*dotConf, error) {
	out := new(dotConf)
	return out, helpers.HandleBlockchainConfig(blockchain, data, out)
}

// GetServices returns the services which are used by artemis
func GetServices() []services.Service {
	return nil
}

// ``
