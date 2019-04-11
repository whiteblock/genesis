package artemis

import (
	util "../../util"
	"encoding/json"
	"fmt"
	"github.com/Whiteblock/mustache"
	"io/ioutil"
	"log"
)

type ArtemisConf map[string]interface{}

func NewConf(data map[string]interface{}) (ArtemisConf, error) {
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
			return nil, fmt.Errorf("Invalid number of validators (%d). Validators must be an even number and greater than 3.", val)
		}
	}
	out := new(ArtemisConf)
	*out = ArtemisConf(util.MergeStringMaps(defaults, data))

	return *out, nil
}

func GetParams() string {
	dat, err := ioutil.ReadFile("./resources/artemis/params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

func GetDefaults() string {
	dat, err := ioutil.ReadFile("./resources/artemis/defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

func GetServices() []util.Service {
	return nil
}

func makeNodeConfig(artemisConf ArtemisConf, identity string, peers string, numNodes int, params map[string]interface{}) (string, error) {

	artConf := map[string]interface{}(artemisConf)
	artConf["identity"] = identity
	filler := util.ConvertToStringMap(artConf)
	filler["peers"] = peers
	filler["numNodes"] = fmt.Sprintf("%d", numNodes)

	var validators int64
	err := util.GetJSONInt64(params, "validators", &validators)
	if err != nil {
		return "", err
	}

	filler["validators"] = fmt.Sprintf("%d", validators)
	dat, err := ioutil.ReadFile("./resources/artemis/artemis-config.toml.mustache")
	if err != nil {
		return "", err
	}
	data, err := mustache.Render(string(dat), filler)
	return data, err
}
