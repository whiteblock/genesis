package beam

import (
	util "../../util"
)

type BeamConf struct {
	Miners   int64 `json:"miners"`
	TxNodes  int64 `json:"txNodes"`
	NilNodes int64 `json:"nilNodes"`
}

func NewConf(data map[string]interface{}) (*BeamConf, error) {
	out := new(BeamConf)

	var err error

	if _, ok := data["miners"]; ok {
		out.Miners, err = util.GetJSONInt64(data, "miners")
		if err != nil {
			return nil, err
		}
	}

	if _, ok := data["txNodes"]; ok {
		out.TxNodes, err = util.GetJSONInt64(data, "txNodes")
		if err != nil {
			return nil, err
		}
	}
	if _, ok := data["nilNodes"]; ok {
		out.NilNodes, err = util.GetJSONInt64(data, "nilNodes")
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func GetParams() string {
	return `[
	{"miners":"int"},
	{"txNodes":"int"},
	{"nilNodes":"int"},
]`
}

func GetDefaults() string {
	return `{
		{"miners":10},
		{"txNodes":2},
		{"nilNodes":0},
			}`
}

func GetServices() []util.Service {
	return nil
}
