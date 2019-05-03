package beam

import (
	"../../db"
	"../../util"
	"../helpers"
	"github.com/Whiteblock/mustache"
	"io/ioutil"
)

type beamConf struct {
	Validators int64 `json:"validators"`
	TxNodes    int64 `json:"txNodes"`
	NilNodes   int64 `json:"nilNodes"`
}

func newConf(data map[string]interface{}) (*beamConf, error) {
	out := new(beamConf)

	err := util.GetJSONInt64(data, "validators", &out.Validators)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "txNodes", &out.TxNodes)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "nilNodes", &out.NilNodes)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetParams fetchs beam related parameters
func GetParams() string {
	dat, err := ioutil.ReadFile("./resources/beam/params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetDefaults fetchs beam related parameter defaults
func GetDefaults() string {
	dat, err := ioutil.ReadFile("./resources/beam/defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetServices returns the services which are used by artemis
func GetServices() []util.Service {
	return nil
}

func makeNodeConfig(bconf *beamConf, keyOwner string, keyMine string, details *db.DeploymentDetails, node int) (string, error) {

	filler := util.ConvertToStringMap(map[string]interface{}{
		"keyOwner": keyOwner,
		"keyMine":  keyMine,
	})
	dat, err := helpers.GetBlockchainConfig("beam", node, "beam-node.cfg.mustache", details)
	if err != nil {
		return "", err
	}
	data, err := mustache.Render(string(dat), filler)
	return data, err
}
