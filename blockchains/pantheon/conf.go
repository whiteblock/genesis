package pantheon

import (
	"../../util"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type PanConf struct {
	NetworkId             int64  `json:"networkId"`
	Difficulty            int64  `json:"difficulty"`
	InitBalance           string `json:"initBalance"`
	MaxPeers              int64  `json:"maxPeers"`
	GasLimit              int64  `json:"gasLimit"`
	Consensus             string `json:"consensus"`
	EthashDifficulty      int64  `json:"fixeddifficulty`
	BlockPeriodSeconds    int64  `json:"blockPeriodSeconds"`
	Epoch                 int64  `json:"epoch"`
	RequestTimeoutSeconds int64  `json:"requesttimeoutseconds"`
	Accounts              int64  `json:"accounts"`
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

	err = util.GetJSONString(data, "consensus", &out.Consensus)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "fixeddifficulty", &out.EthashDifficulty)
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

	err = util.GetJSONInt64(data, "requesttimeoutseconds", &out.RequestTimeoutSeconds)
	if err != nil {
		return nil, err
	}

	initBalance, exists := data["initBalance"]
	if exists && initBalance != nil {
		switch initBalance.(type) {
		case json.Number:
			out.InitBalance = initBalance.(json.Number).String()
		case string:
			out.InitBalance = initBalance.(string)
		default:
			return nil, fmt.Errorf("incorrect type for initBalance given")
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
	return []util.Service{
		{ //Include a geth node for transaction signing
			Name:  "geth",
			Image: "gcr.io/whiteblock/geth:master",
			Env:   nil,
		},
	}
}
