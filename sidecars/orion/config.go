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
	Nodeurl                string `json:"nodeurl"`
	Nodeport               int64  `json:"nodeport"`
	Clienturl              string `json:"clienturl"`
	Clientport             int64  `json:"clientport"`
	Tls                    string `json:"tls"`
	Nodenetworkinterface   string `json:"nodenetworkinterface"`
	Clientnetworkinterface string `json:"clientnetworkinterface"`
}

/**
 * Fills in the defaults for missing parts,
 */
func newConf(tn *testnet.TestNet) (*orionConf, error) {

	out := new(orionConf)
	err := json.Unmarshal([]byte(GetDefaults()), out)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	tn.BuildState.GetP("nodeurl", &out.Nodeurl)
	tn.BuildState.GetP("nodeport", &out.Nodeport)
	tn.BuildState.GetP("clienturl", &out.Clienturl)
	tn.BuildState.GetP("clientport", &out.Clientport)
	tn.BuildState.GetP("tls", &out.Tls)
	tn.BuildState.GetP("nodenetworkinterface", &out.Nodenetworkinterface)
	tn.BuildState.GetP("clientnetworkinterface", &out.Clientnetworkinterface)

	return out, nil
}

// GetParams fetchs pantheon related parameters
func GetParams() string {
	dat, err := helpers.GetStaticBlockchainConfig("orion", "params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetDefaults fetchs pantheon related parameter defaults
func GetDefaults() string {
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
	err = json.Unmarshal([]byte(GetDefaults()), &filler)
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
