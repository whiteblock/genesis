package paritypoa

import (
	util "../../util"
	"encoding/json"
	"fmt"
	"github.com/Whiteblock/mustache"
	"io/ioutil"
	"log"
	//"strconv"
)

type ParityPOAConf struct {
	// ForceSealing              bool   `json:"forceSealing"`
	// ResealOnTxs               string `json:"resealOnTxs"`
	// ResealMinPeriod           int64  `json:"resealMinPeriod"`
	// ResealMaxPeriod           int64  `json:"resealMaxPeriod"`
	// WorkQueueSize             int64  `json:"workQueueSize"`
	// RelaySet                  string `json:"relaySet"`
	// UsdPerTx                  string `json:"usdPerTx"`
	// UsdPerEth                 string `json:"usdPerEth"`
	// PriceUpdatePeriod         string `json:"priceUpdatePeriod"`
	// GasFloorTarget            string `json:"gasFloorTarget"`
	// GasCap                    string `json:"gasCap"`
	// TxQueueSize               int64  `json:"txQueueSize"`
	// TxQueueGas                string `json:"txQueueGas"`
	// TxQueueStrategy           string `json:"txQueueStrategy"`
	// TxGasLimit                string `json:"txGasLimit"`
	// TxTimeLimit               int64  `json:"txTimeLimit"`
	// RemoveSolved              bool   `json:"removeSolved"`
	// RefuseServiceTransactions bool   `json:"refuseServiceTransactions"`
	// EnableIPFS                bool   `json:"enableIPFS"`
	// NetworkDiscovery          bool   `json:"networkDiscovery"`
	// ExtraAccounts             int64  `json:"extraAccounts"`
	ChainId                   int64  `json:"chainId"`
	// MaxPeers                  int64  `json:"maxPeers"`
	// HomesteadBlock            int64  `json:"homesteadBlock"`
	// Eip155Block               int64  `json:"eip155Block"`
	// Eip158Block               int64  `json:"eip158Block"`
	// MinimumDifficulty         int64  `json:"minimumDifficulty"`
	// DifficultyBoundDivisor    int64  `json:"difficultyBoundDivisor"`
	// DurationLimit             int64  `json:"durationLimit"`
	// BlockReward               int64  `json:"blockReward"`

	//engine
	StepDuration              int64 `json:"stepDuration"`
	//params
	GasLimitBoundDivisor      int64  `json:"gasLimitBoundDivisor"`
	MaximumExtraDataSize      int64  `json:"maximumExtraDataSize"`
	MinGasLimit               int64  `json:"minGasLimit"`
	NetworkId                 int64  `json:"networkId"`
	ValidateChainIdTransition int64  `json:"validateChainIdTransition"`
	EIP155Transition          int64  `json:"eip155Transition"`
	EIP140Transition          int64  `json:"eip140Transition"`
	EIP211Transition          int64  `json:"eip211Transition"`
	EIP214Transition          int64  `json:"eip214Transition"`
	EIP658Transition          int64  `json:"eip658Transition"`
	//genesis
	Step                      int64  `json:"step"`
	Signature                 string  `json:"signature"`
	Difficulty                int64  `json:"difficulty"`
	GasLimit                  int64  `json:"gasLimit"`
	InitBalance               string `json:"initBalance"`

}

/**
 * Fills in the defaults for missing parts,
 */
func NewConf(data map[string]interface{}) (*ParityPOAConf, error) {
	out := new(ParityPOAConf)
	err := json.Unmarshal([]byte(GetDefaults()), out)
	fmt.Printf("%+v\n", *out)
	if data == nil {
		log.Println(err)
		return out, err
	}

	err = util.GetJSONInt64(data, "stepDuration", &out.StepDuration)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "gasLimitBoundDivisor", &out.GasLimitBoundDivisor)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "maximumExtraDataSize", &out.MaximumExtraDataSize)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "minGasLimit", &out.MinGasLimit)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "minGasLimit", &out.MinGasLimit)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "validateChainIdTransition", &out.ValidateChainIdTransition)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "eip155Transition", &out.EIP155Transition)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "eip140Transition", &out.EIP140Transition)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "eip211Transition", &out.EIP211Transition)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "eip214Transition", &out.EIP214Transition)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "eip658Transition", &out.EIP658Transition)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "networkId", &out.NetworkId)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "difficulty", &out.Difficulty)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "gasLimit", &out.GasLimit)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "maximumExtraDataSize", &out.MaximumExtraDataSize)
	if err != nil {
		return nil, err
	}

	// err = util.GetJSONBool(data, "forceSealing", &out.ForceSealing)

	return out, nil
}

func GetParams() string {
	dat, err := ioutil.ReadFile("./resources/parity-poa/params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}
func GetDefaults() string {
	dat, err := ioutil.ReadFile("./resources/parity-poa/defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

func GetServices() []util.Service {
	return []util.Service{
		util.Service{
			Name:  "Geth",
			Image: "gcr.io/whiteblock/ethereum:latest",
			Env:   nil,
		},
	}
}

/*
   passwordFile
   unlock
*/
func BuildConfig(pconf *ParityPOAConf, files map[string]string, wallets []string, passwordFile string) (string, error) {

	dat, err := util.GetBlockchainConfig("parity", "config.toml.template", files)
	if err != nil {
		log.Println(err)
		return "", err
	}
	var tmp interface{}

	raw, err := json.Marshal(*pconf)
	if err != nil {
		log.Println(err)
		return "", err
	}

	err = json.Unmarshal(raw, &tmp)
	if err != nil {
		log.Println(err)
		return "", err
	}

	mp := util.ConvertToStringMap(tmp)
	raw, err = json.Marshal(wallets)
	if err != nil {
		log.Println(err)
		return "", err
	}
	mp["unlock"] = string(raw)
	mp["passwordFile"] = fmt.Sprintf("[\"%s\"]", passwordFile)
	mp["networkId"] = fmt.Sprintf("%d", pconf.NetworkId)
	return mustache.Render(string(dat), mp)
}

func BuildSpec(pconf *ParityPOAConf, files map[string]string, wallets []string) (string, error) {

	accounts := make(map[string]interface{})
	for _, wallet := range wallets {
		accounts[wallet] = map[string]interface{}{
			"balance": pconf.InitBalance,
		}
	}

	var validators []string
	for _, wallet := range wallets {
		validators = append(validators, wallet)
	}

	tmp := map[string]interface{}{
		"stepDuration":           pconf.StepDuration,
		"validators":             validators,
		"difficulty":             fmt.Sprintf("0x%x", pconf.Difficulty),
		"gasLimit":               fmt.Sprintf("0x%x", pconf.GasLimit),
		"networkId":              fmt.Sprintf("0x%x", pconf.NetworkId),
		"maximumExtraDataSize":   fmt.Sprintf("0x%x", pconf.MaximumExtraDataSize),
		"minGasLimit":            fmt.Sprintf("0x%x", pconf.MinGasLimit),
		"gasLimitBoundDivisor":   fmt.Sprintf("0x%x", pconf.GasLimitBoundDivisor),
		"validateChainIdTransition": pconf.ValidateChainIdTransition,
		"eip155Transition":       pconf.EIP155Transition,
		"eip140Transition":       pconf.EIP140Transition,
        "eip211Transition":       pconf.EIP211Transition,
        "eip214Transition":       pconf.EIP214Transition,
        "eip658Transition":       pconf.EIP658Transition,
		"accounts":               accounts,
	}
	filler := util.ConvertToStringMap(tmp)
	dat, err := util.GetBlockchainConfig("parity", "spec.json.mustache", files)
	if err != nil {
		return "", err
	}
	return mustache.Render(string(dat), filler)
}