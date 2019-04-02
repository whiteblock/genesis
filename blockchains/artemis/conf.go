package artemis

import (
	"log"
	"encoding/json"
	"io/ioutil"
	"github.com/Whiteblock/mustache"
	util "../../util"
)

type ArtemisConf map[string]interface{}

func NewConf(data map[string]interface{}) (*ArtemisConf, error) {
	rawDefaults := GetDefaults()
	defaults := map[string]interface{}{}

	err := json.Unmarshal([]byte(rawDefaults), &defaults)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	out := new(ArtemisConf)
	*out = ArtemisConf(util.MergeStringMaps(defaults, data))

	return out, nil
}

func GetParams() string {
    dat, err := ioutil.ReadFile("./resources/artemis/params.json")
    if err != nil {
        panic(err)//Missing required files is a fatal error
    }
    return string(dat)
}

func GetDefaults() string {
    dat, err := ioutil.ReadFile("./resources/artemis/defaults.json")
    if err != nil {
        panic(err)//Missing required files is a fatal error
    }
    return string(dat)
}

func GetServices() []util.Service {
	return nil
}

func makeNodeConfig(artemisConf *ArtemisConf, identity string, peers string) (string,error){
	artConf := map[string]interface{}(*artemisConf)
	artConf["identity"] = identity
	filler := util.ConvertToStringMap(artConf)
	filler["peers"] = peers
    dat, err := ioutil.ReadFile("./resources/artemis/artemis-config.toml.mustache")
    if err != nil {
        return "",err
    }
    data, err := mustache.Render(string(dat), filler)
    return data,err
}
