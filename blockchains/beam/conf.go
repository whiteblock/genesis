package beam

import (
	"io/ioutil"
	"github.com/Whiteblock/mustache"
	util "../../util"
)

type BeamConf struct {
	Validators int64 `json:"validators"`
	TxNodes    int64 `json:"txNodes"`
	NilNodes   int64 `json:"nilNodes"`
}

func NewConf(data map[string]interface{}) (*BeamConf, error) {
	out := new(BeamConf)

	err := util.GetJSONInt64(data, "validators",&out.Validators)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "txNodes",&out.TxNodes)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "nilNodes",&out.NilNodes)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func GetParams() string {
    dat, err := ioutil.ReadFile("./resources/beam/params.json")
    if err != nil {
        panic(err)//Missing required files is a fatal error
    }
    return string(dat)
}

func GetDefaults() string {
    dat, err := ioutil.ReadFile("./resources/beam/defaults.json")
    if err != nil {
        panic(err)//Missing required files is a fatal error
    }
    return string(dat)
}

func GetServices() []util.Service {
	return nil
}

func makeNodeConfig(bconf *BeamConf,keyOwner string,keyMine string) (string,error){

    filler := util.ConvertToStringMap(map[string]interface{}{
        "keyOwner":keyOwner,
        "keyMine":keyMine,
        
    })
    dat, err := ioutil.ReadFile("./resources/beam/beam-node.cfg.mustache")
    if err != nil {
        return "",err
    }
    data, err := mustache.Render(string(dat), filler)
    return data,err
}