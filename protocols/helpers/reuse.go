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

package helpers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/state"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"io/ioutil"
) //log "github.com/sirupsen/logrus"

// ScpAndDeferRemoval Copy a file over to a server, and then defer it for removal after the build is completed
func ScpAndDeferRemoval(client ssh.Client, buildState *state.BuildState, src string, dst string) {
	buildState.Defer(func() { client.Run(fmt.Sprintf("rm -rf %s", dst)) })
	err := client.Scp(src, dst)
	if err != nil {
		buildState.ReportError(err)
		return
	}
}

// GetDefaults get any available default value for a given term.
// will be nil,false if it is not found
func GetDefaults(details *db.DeploymentDetails, term string) (interface{}, bool) {
	if details.Extras == nil {
		return nil, false
	}
	idefaults, ok := details.Extras["defaults"]
	if !ok {
		return nil, false
	}
	defaults, ok := idefaults.(map[string]interface{})
	if !ok {
		return nil, false
	}

	return defaults[term], true
}

// CheckDeployFlag checks for the presence of an extras flag.
func CheckDeployFlag(details *db.DeploymentDetails, flag string) bool {
	if details.Extras == nil {
		return false
	}
	iflags, ok := details.Extras["flags"]
	if !ok {
		return false
	}
	flags, ok := iflags.(map[string]interface{})
	if !ok {
		return false
	}
	flagVal, ok := flags[flag]
	if !ok {
		return false
	}
	return flagVal.(bool)
}

// GetFileDefault gets the default value for a file if it exists in the extras
func GetFileDefault(details *db.DeploymentDetails, file string) (string, bool) {
	ifileDefaults, ok := GetDefaults(details, "files")
	if !ok {
		return "", false
	}
	fileDefaults, ok := ifileDefaults.(map[string]interface{})
	if !ok {
		return "", false
	}

	iOut, ok := fileDefaults[file]
	if !ok {
		return "", false
	}
	out, ok := iOut.(string)
	return out, ok

}

// GetStaticBlockchainConfig fetches a static file resource for a blockchain, which will never change
func GetStaticBlockchainConfig(blockchain string, file string) ([]byte, error) {
	return ioutil.ReadFile(fmt.Sprintf("%s/%s/%s", conf.ResourceDir, blockchain, file))
}

// GetGlobalBlockchainConfig fetches a static file resource for a blockchain, which will be the same for all of the nodes
func GetGlobalBlockchainConfig(tn *testnet.TestNet, file string) ([]byte, error) {
	res, exists := GetFileDefault(&tn.CombinedDetails, file)
	if exists && len(res) != 0 {
		return base64.StdEncoding.DecodeString(res)
	}
	return GetStaticBlockchainConfig(tn.LDD.Blockchain, file)
}

// GetBlockchainConfig fetches dynamic config template files for the blockchain. Should be used in most cases instead of
// GetStaticBlockchainConfig as it provides the user the functionality for `-t..` in the build command for the CLI
func GetBlockchainConfig(blockchain string, node int, file string, details *db.DeploymentDetails) ([]byte, error) {

	if details.Files != nil {
		if len(details.Files) > node && details.Files[node] != nil {
			res, exists := details.Files[node][file]
			if exists && len(res) != 0 {
				return base64.StdEncoding.DecodeString(res)
			}
		} else {
			res, exists := GetFileDefault(details, file)
			if exists && len(res) != 0 {
				return base64.StdEncoding.DecodeString(res)
			}
		}
	}
	return ioutil.ReadFile(fmt.Sprintf("%s/%s/%s", conf.ResourceDir, blockchain, file))
}

// HandleBlockchainConfig handles the creation of a blockchain configuration from the defaults and given
// data from the deployment details
func HandleBlockchainConfig(blockchain string, data map[string]interface{}, out interface{}) error {
	dat, err := GetStaticBlockchainConfig(blockchain, "defaults.json")
	if err != nil {
		return util.LogError(err)
	}
	err = json.Unmarshal(dat, out)
	if data == nil {
		return util.LogError(err)
	}
	tmp, err := json.Marshal(data)
	if err != nil {
		return util.LogError(err)
	}
	return json.Unmarshal(tmp, out)
}

// getError retrieves the error value from the build state, depending on the settings.
func getError(tn *testnet.TestNet, s settings) error {
	if s.reportError {
		return tn.BuildState.GetError()
	}
	var err error
	hasErr := tn.BuildState.GetP("error", &err)
	if !hasErr {
		return nil
	}
	return err
}
func fetchPreGeneratedKeys(tn *testnet.TestNet, file string) ([]string, error) {
	rawPrivateKeys, err := GetGlobalBlockchainConfig(tn, file)
	if err != nil {
		return nil, util.LogError(err)
	}
	var out []string
	return out, util.LogError(json.Unmarshal(rawPrivateKeys, &out))
}

// FetchPreGeneratedPrivateKeys gets the pregenerated private keys for a blockchain from privatekeys.json
func FetchPreGeneratedPrivateKeys(tn *testnet.TestNet) ([]string, error) {
	return fetchPreGeneratedKeys(tn, "privatekeys.json")
}

// FetchPreGeneratedPublicKeys gets the pregenerated public keys for a blockchain from publickeys.json
func FetchPreGeneratedPublicKeys(tn *testnet.TestNet) ([]string, error) {
	return fetchPreGeneratedKeys(tn, "publickeys.json")
}
