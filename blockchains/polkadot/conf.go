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
	// "encoding/json"
	// "fmt"
	"github.com/whiteblock/genesis/blockchains/helpers"
)

type dotConf struct {
	validatorMode  bool  `json:"validatorMode"`
}

/**
 * Fills in the defaults for missing parts,
 */
func newConf(data map[string]interface{}) (*dotConf, error) {
	out := new(dotConf)
	err := helpers.HandleBlockchainConfig(blockchain, data, out)
	if err != nil || data == nil {
		return out, err
	}

	// vMode := data["validatorMode"]
	// if vMode == true {
	// 	filler := util.ConvertToStringMap(dotConf)
	// 	filler["validatorMode"] = "validator"
	// 	}
	// }
	
	return out, nil
}

// GetServices returns the services which are used by artemis
func GetServices() []helpers.Service {
	return nil
}
// ``