package pantheon

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	util "../../util"
)

type PanConf struct {
	NetworkId      		int64  `json:"networkId"`
	Difficulty     		int64  `json:"difficulty"`
	InitBalance    		string `json:"initBalance"`
	MaxPeers       		int64  `json:"maxPeers"`
	GasLimit       		int64  `json:"gasLimit"`
	BlockPeriodSeconds 	int64  `json:'blockPeriodSeconds'`
	Epoch          		int64  `json:'epoch'`
	// ExtraData	   string `json:"extraData"` //for IBFT2
}

/**
 * Fills in the defaults for missing parts,
 */
func NewConf(data map[string]interface{}) (*PanConf, error) {
	out := new(PanConf)
	err := json.Unmarshal([]byte(GetDefaults()), out)

	if data == nil {
		return out, nil
	}

	err = util.GetJSONInt64(data, "networkId", &out.NetworkId)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "difficulty", &out.Difficulty)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "maxPeers", &out.MaxPeers)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "gasLimit", &out.GasLimit)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "blockPeriodSeconds", &out.BlockPeriodSeconds)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "epoch", &out.Epoch)
	if err != nil {
		return nil, err
	}

	// err = util.GetJSONString(data, "extraData", &out.ExtraData)
	// if err != nil {
	// 	return nil, err
	// }

	initBalance, exists := data["initBalance"]
	if exists && initBalance != nil {
		switch initBalance.(type) {
		case json.Number:
			out.InitBalance = initBalance.(json.Number).String()
		case string:
			out.InitBalance = initBalance.(string)
		default:
			return nil, errors.New("Incorrect type for initBalance given")
		}
	}

	return out, nil
}

func GetParams() string {
	dat, err := ioutil.ReadFile("./resources/pantheon/params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

func GetDefaults() string {
	dat, err := ioutil.ReadFile("./resources/pantheon/defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

func GetServices() []util.Service {
	return nil
}
