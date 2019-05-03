package helpers

import (
	"../../db"
	"../../ssh"
	"../../state"
	"encoding/base64"
	"fmt"
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
