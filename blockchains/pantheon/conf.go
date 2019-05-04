package pantheon

import (
	"../../util"
	"../helpers"
	"encoding/json"
)

type panConf struct {
	NetworkID             int64  `json:"networkId"`
	Difficulty            int64  `json:"difficulty"`
	InitBalance           string `json:"initBalance"`
	MaxPeers              int64  `json:"maxPeers"`
	GasLimit              int64  `json:"gasLimit"`
	Consensus             string `json:"consensus"`
	FixedDifficulty       int64  `json:"fixedDifficulty"`
	BlockPeriodSeconds    int64  `json:"blockPeriodSeconds"`
	Epoch                 int64  `json:"epoch"`
	RequestTimeoutSeconds int64  `json:"requesttimeoutseconds"`
	Accounts              int64  `json:"accounts"`
	Orion                 bool   `json:"orion"`
}

/**
 * Fills in the defaults for missing parts,
 */
func newConf(data map[string]interface{}) (*panConf, error) {

	out := new(panConf)
	err := json.Unmarshal([]byte(GetDefaults()), out)
	if data == nil {
		return out, util.LogError(err)
	}
	tmp, err := json.Marshal(data)
	if err != nil {
		return nil, util.LogError(err)
	}
	err = json.Unmarshal(tmp, out)

	return out, err
}

// GetParams fetchs pantheon related parameters
func GetParams() string {
	dat, err := helpers.GetStaticBlockchainConfig(blockchain, "params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetDefaults fetchs pantheon related parameter defaults
func GetDefaults() string {
	dat, err := helpers.GetStaticBlockchainConfig(blockchain, "defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetServices returns the services which are used by artemis
func GetServices() []util.Service {
	return []util.Service{
		{ //Include a geth node for transaction signing
			Name:  "geth",
			Image: "gcr.io/whiteblock/geth:master",
			Env:   nil,
		},
	}
}
