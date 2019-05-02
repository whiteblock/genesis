package orion

import (
	"../../db"
	"../../util"
	"encoding/json"
	"io/ioutil"
	"../helpers"
	"github.com/Whiteblock/mustache"
)

type orionConf struct {
	Nodeurl                     string  `json:"nodeurl"`
	Nodeport                    int64  `json:"nodeport"`
	Clienturl                   string `json:"clienturl"`
	Clientport                  int64  `json:"clientport"`
	Tls                         string  `json:"tls"`
	Nodenetworkinterface        string  `json:"nodenetworkinterface"`
	Clientnetworkinterface      string  `json:"clientnetworkinterface"`
}

/**
 * Fills in the defaults for missing parts,
 */
func newConf(data map[string]interface{}) (*orionConf, error) {

	out := new(orionConf)
	err := json.Unmarshal([]byte(GetDefaults()), out)

	if data == nil {
		return out, nil
	}

	err = util.GetJSONString(data, "nodeurl", &out.Nodeurl)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "nodeport", &out.Nodeport)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONString(data, "clienturl", &out.Clienturl)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "clientport", &out.Clientport)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONString(data, "tls", &out.Tls)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONString(data, "nodenetworkinterface", &out.Nodenetworkinterface)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONString(data, "clientnetworkinterface", &out.Clientnetworkinterface)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// GetParams fetchs pantheon related parameters
func GetParams() string {
	dat, err := ioutil.ReadFile("./resources/orion/params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetDefaults fetchs pantheon related parameter defaults
func GetDefaults() string {
	dat, err := ioutil.ReadFile("./resources/orion/defaults.json")
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
	filler := details.Params
	dat, err := helpers.GetBlockchainConfig("orion", node, "orion.config.mustache", details)
	if err != nil {
		return "", err
	}
	data, err := mustache.Render(string(dat), filler)
	return data, err
}
