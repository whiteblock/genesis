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

package prismatic

import (
	"github.com/Whiteblock/genesis/util"
	"github.com/Whiteblock/genesis/blockchains/helpers"
	"encoding/base64"
	"encoding/json"
	"log"
)

type prismaticConf struct {
}

func newConf(data map[string]interface{}) (*prismaticConf, error) {
	out := new(prismaticConf)
	err := json.Unmarshal([]byte(GetDefaults()), out)
	if data == nil {
		return out, util.LogError(err)
	}
	tmp, err := json.Marshal(data)
	if err != nil {
		return nil, util.LogError(err)
	}
	return out, json.Unmarshal(tmp, out)
}

// GetServices returns the services which are used by rchain
func GetServices() []util.Service {
	return []util.Service{}
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
