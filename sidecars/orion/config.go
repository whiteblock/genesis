package orion

import (
	"../../blockchains/helpers"
	"../../db"
	"../../testnet"
	"../../util"
	"encoding/json"
	"github.com/Whiteblock/mustache"
	"log"
)

type orionConf struct {
	NodeURL                string `json:"nodeURL"`
	NodePort               int64  `json:"nodePort"`
	ClientURL              string `json:"clientURL"`
	ClientPort             int64  `json:"clientPort"`
	TLS                    string `json:"tls"`
	NodeNetworkInterface   string `json:"nodeNetworkInterface"`
	ClientNetworkInterface string `json:"clientNetworkInterface"`
}

/**
 * Fills in the defaults for missing parts,
 */
func newConf(tn *testnet.Adjunct) (*orionConf, error) {

	out := new(orionConf)
	err := json.Unmarshal([]byte(getDefaults()), out)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	tn.BuildState.GetP("nodeURL", &out.NodeURL)
	tn.BuildState.GetP("nodePort", &out.NodePort)
	tn.BuildState.GetP("clientURL", &out.ClientURL)
	tn.BuildState.GetP("clientPort", &out.ClientPort)
	tn.BuildState.GetP("tls", &out.TLS)
	tn.BuildState.GetP("nodeNetworkInterface", &out.NodeNetworkInterface)
	tn.BuildState.GetP("clientNetworkInterface", &out.ClientNetworkInterface)

	return out, nil
}

// getParams fetchs orion related parameters
func getParams() string {
	dat, err := helpers.GetStaticBlockchainConfig("orion", "params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// getDefaults fetchs orion related parameter defaults
func getDefaults() string {
	dat, err := helpers.GetStaticBlockchainConfig("orion", "defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetServices returns the services which are used by artemis
func GetServices() []util.Service {
	return nil
}

func makeNodeConfig(orionconf *orionConf, node int, details *db.DeploymentDetails) (string, error) {
	filler, err := util.CopyMap(details.Params)
	if err != nil {
		log.Println(err)
		return "", nil
	}
	err = json.Unmarshal([]byte(getDefaults()), &filler)
	if err != nil {
		log.Println(err)
		return "", nil
	}
	dat, err := helpers.GetBlockchainConfig("orion", node, "orion.conf.mustache", details)
	if err != nil {
		return "", err
	}
	data, err := mustache.Render(string(dat), util.ConvertToStringMap(filler))
	return data, err
}
