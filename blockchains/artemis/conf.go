package artemis

import (
	"../../db"
	"../../util"
	"../helpers"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Whiteblock/mustache"
	"log"
)

type artemisConf map[string]interface{}

func newConf(data map[string]interface{}) (artemisConf, error) {
	rawDefaults := GetDefaults()
	defaults := map[string]interface{}{}

	err := json.Unmarshal([]byte(rawDefaults), &defaults)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var val int64
	err = util.GetJSONInt64(data, "validators", &val) //Check provided validators
	if err == nil {
		if val < 4 || val%2 != 0 {
			return nil, fmt.Errorf("invalid number of validators (%d): must be an even number and greater than 3", val)
		}
	}
	out := new(artemisConf)
	*out = artemisConf(util.MergeStringMaps(defaults, data))

	return *out, nil
}

// GetParams fetchs artemis related parameters
func GetParams() string {
	dat, err := helpers.GetStaticBlockchainConfig("artemis", "params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetDefaults fetchs artemis related parameter defaults
func GetDefaults() string {
	dat, err := helpers.GetStaticBlockchainConfig("artemis", "defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetServices returns the services which are used by artemis
func GetServices() []util.Service {
	return []util.Service{
		{
			Name:  "wb_influx_proxy",
			Image: "gcr.io/wb-genesis/bitbucket.org/whiteblockio/influx-proxy:master",
			Env: map[string]string{
				"BASIC_AUTH_BASE64": base64.StdEncoding.EncodeToString([]byte(conf.InfluxUser + ":" + conf.InfluxPassword)),
				"INFLUXDB_URL":      conf.Influx,
				"BIND_PORT":         "8086",
			},
		},
	}
}

func makeNodeConfig(aconf artemisConf, identity string, peers string, node int, details *db.DeploymentDetails, constantsRaw string) (string, error) {

	artConf, err := util.CopyMap(aconf)
	if err != nil {
		log.Println(err)
		return "", err
	}
	artConf["identity"] = identity
	filler := util.ConvertToStringMap(artConf)
	filler["peers"] = peers
	filler["numNodes"] = fmt.Sprintf("%d", details.Nodes)
	filler["constants"] = constantsRaw
	var validators int64
	err = util.GetJSONInt64(details.Params, "validators", &validators)
	if err != nil {
		return "", err
	}

	filler["validators"] = fmt.Sprintf("%d", validators)
	dat, err := helpers.GetBlockchainConfig("artemis", node, "artemis-config.toml.mustache", details)
	if err != nil {
		return "", err
	}
	data, err := mustache.Render(string(dat), filler)
	return data, err
}
