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

package ethereum

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"reflect"
)

//BaseConfig contains the parameters which should be shared amongst clients
type BaseConfig struct {
	Consensus      string `json:"consensus"`
	EIP150Block    int64  `json:"eip150Block"`
	ExtraAccounts  int64  `json:"extraAccounts"`
	ExtraData      string `json:"extraData"`
	GasLimit       int64  `json:"gasLimit"`
	HomesteadBlock int64  `json:"homesteadBlock"`
	MaxPeers       int64  `json:"maxPeers"`
	NetworkID      int64  `json:"networkId"`
	Nonce          string `json:"nonce"`
	Timestamp      int64  `json:"timestamp"`

	Difficulty int64 `json:"difficulty"`

	BlockPeriodSeconds int64  `json:"blockPeriodSeconds"`
	Epoch              int64  `json:"epoch"`
	InitBalance        string `json:"initBalance"`

	ECIP1010Length int64  `json:"ecip1010Length"`
	MixHash        string `json:"mixHash"`
}

var sharedConfigParameters = []string{
	//standard ethereum
	"consensus",
	"eip150Block",
	"extraAccounts",
	"extraData",
	"gasLimit",
	"homesteadBlock",
	"maxPeers",
	"networkId",
	"nonce",
	"timestamp",

	//ethereum pow
	"difficulty",

	//ethereum non-pow
	"blockPeriodSeconds",
	"epoch",
	"initBalance",
}

var optionalSharedConfigParameters = []string{
	//ethereum classic
	"ecip1010Length",
	"mixHash",
}

const configPrefix = "config_"

//StoreConfigParameters stores the known shared ethereum parameters from your config. Allowing
//another compatible client to use it if needed.
func StoreConfigParameters(tn *testnet.TestNet, confObj interface{}) error {
	if confObj == nil {
		return fmt.Errorf("given a nil configuration")
	}
	/*if reflect.ValueOf(confObj).Type().Kind() == reflect.Ptr {
		return StoreConfigParameters(tn, *(confObj).(*interface{}))
	}*/
	tmp, err := json.Marshal(confObj)
	if err != nil {
		return util.LogError(err)
	}
	var conf map[string]interface{}
	err = json.Unmarshal(tmp, &conf)
	if err != nil {
		return util.LogError(err)
	}
	for _, paramName := range sharedConfigParameters {
		if _, ok := conf[paramName]; !ok {
			return fmt.Errorf("missing the required parameter \"%s\"", paramName)
		}
		log.WithFields(log.Fields{"name": paramName, "value": conf[paramName]}).Trace("storing parameter")
		tn.BuildState.Set(configPrefix+paramName, conf[paramName])
	}

	for _, paramName := range optionalSharedConfigParameters {
		if _, ok := conf[paramName]; !ok {
			log.WithFields(log.Fields{"name": paramName}).Debug("skipping missing parameter")
			continue
		}
		tn.BuildState.Set(configPrefix+paramName, conf[paramName])
	}
	return nil
}

//FetchConfigParameters merges the already stored config into your config object
//It must be given a pointer and will only fill in a limited set of fields
func FetchConfigParameters(tn *testnet.TestNet, outConf interface{}) error {
	if reflect.ValueOf(outConf).Type().Kind() != reflect.Ptr {
		return fmt.Errorf("FetchParameters expects a pointer to the output")
	}

	confToMerge := map[string]interface{}{}
	for _, paramName := range sharedConfigParameters {
		exists := false
		confToMerge[paramName], exists = tn.BuildState.Get(configPrefix + paramName)
		if !exists {
			return fmt.Errorf("missing the required parameter \"%s\"", paramName)
		}
	}

	for _, paramName := range optionalSharedConfigParameters {
		confToMerge[paramName], _ = tn.BuildState.Get(configPrefix + paramName)
	}

	tmp, err := json.Marshal(confToMerge)
	if err != nil {
		return util.LogError(err)
	}
	return util.LogError(json.Unmarshal(tmp, outConf))
}

/*
//etc
	Identity           string `json:"identity"`
	Name               string `json:"name"`
	ExtraData          string `json:"extraData"`

	DAOHFBlock         int64  `json:"daoHFBlock"`
	EIP155_160Block    int64  `json:"eip155_160Block"`
	ECIP1017Block      int64  `json:"ecip1017Block"`
	ECIP1017Era        int64  `json:"ecip1017Era"`

//multigeth
	ChainID            int64  `json:"chainId"`
	EIP155Block        int64  `json:"eip155Block"`
	EIP158Block        int64  `json:"eip158Block"`
	ByzantiumBlock     int64  `json:"byzantiumBlock"`
	DisposalBlock      int64  `json:"disposalBlock"`
	//  ConstantinopleBlock int64  `json:"constantinopleBlock"`
	ECIP1017EraRounds  int64 `json:"ecip1017EraRounds"`
	EIP160FBlock       int64 `json:"eip160FBlock"`
	ECIP1010PauseBlock int64 `json:"ecip1010PauseBlock"`

	//  TrustedCheckpoint   int64  `json:"trustedCheckpoint"`
	ExposedAccounts int64  `json:"exposedAccounts"`
	Verbosity       int    `json:"verbosity"`
*/
