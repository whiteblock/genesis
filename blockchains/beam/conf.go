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
	return `[
	{"validators":"int"},
	{"txNodes":"int"},
	{"nilNodes":"int"},
]`
}

func GetDefaults() string {
	return `{
		{"validators":10},
		{"txNodes":2},
		{"nilNodes":0},
}`
}

func GetServices() []util.Service {
	return nil
}
/*
	{{{keyOwner}}}
	{{{secretMinerKeys}}}
 */

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