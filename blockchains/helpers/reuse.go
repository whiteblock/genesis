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

package helpers

import (
	"encoding/base64"
	"fmt"
	"github.com/Whiteblock/genesis/db"
	"github.com/Whiteblock/genesis/ssh"
	"github.com/Whiteblock/genesis/state"
	"io/ioutil"
)

// ScpAndDeferRemoval Copy a file over to a server, and then defer it for removal after the build is completed
func ScpAndDeferRemoval(client *ssh.Client, buildState *state.BuildState, src string, dst string) {
	buildState.Defer(func() { client.Run(fmt.Sprintf("rm -rf %s", dst)) })
	err := client.Scp(src, dst)
	if err != nil {
		buildState.ReportError(err)
		return
	}
}

// GetDefaults get any availible default value for a given term.
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
	return ioutil.ReadFile(fmt.Sprintf("./resources/%s/%s", blockchain, file))
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
	return ioutil.ReadFile(fmt.Sprintf("./resources/%s/%s", blockchain, file))
}
