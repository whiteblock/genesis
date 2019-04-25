package helpers

import (
	db "../../db"
	ssh "../../ssh"
	state "../../state"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
)

func ScpAndDeferRemoval(client *ssh.Client, buildState *state.BuildState, src string, dst string) {
	buildState.Defer(func() { client.Run(fmt.Sprintf("rm -rf %s", dst)) })
	err := client.Scp(src, dst)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return
	}
}

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

func GetStaticBlockchainConfig(blockchain string, file string) ([]byte, error) {
	return ioutil.ReadFile(fmt.Sprintf("./resources/%s/%s", blockchain, file))
}

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
