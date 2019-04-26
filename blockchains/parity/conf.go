package parity

import (
	"../../db"
	"../../util"
	"../helpers"
	"encoding/json"
	"fmt"
	"github.com/Whiteblock/mustache"
	"log"
)

type parityConf struct {
	BlockReward               int64  `json:"blockReward"`
	ChainID                   int64  `json:"chainId"`
	Consensus                 string `json:"consensus"` //TODO
	Difficulty                int64  `json:"difficulty"`
	DifficultyBoundDivisor    int64  `json:"difficultyBoundDivisor"`
	DontMine                  bool   `json:"dontMine"`
	DurationLimit             int64  `json:"durationLimit"`
	Eip155Block               int64  `json:"eip155Block"`
	Eip158Block               int64  `json:"eip158Block"`
	EIP140Transition          int64  `json:"eip140Transition"`
	EIP155Transition          int64  `json:"eip155Transition"`
	EIP211Transition          int64  `json:"eip211Transition"`
	EIP214Transition          int64  `json:"eip214Transition"`
	EIP658Transition          int64  `json:"eip658Transition"`
	EnableIPFS                bool   `json:"enableIPFS"`
	ExtraAccounts             int64  `json:"extraAccounts"`
	ForceSealing              bool   `json:"forceSealing"`
	GasCap                    string `json:"gasCap"`
	GasFloorTarget            string `json:"gasFloorTarget"`
	GasLimit                  int64  `json:"gasLimit"`
	GasLimitBoundDivisor      int64  `json:"gasLimitBoundDivisor"`
	HomesteadBlock            int64  `json:"homesteadBlock"`
	InitBalance               string `json:"initBalance"`
	MaximumExtraDataSize      int64  `json:"maximumExtraDataSize"`
	MaxPeers                  int64  `json:"maxPeers"`
	MinGasLimit               int64  `json:"minGasLimit"`
	MinimumDifficulty         int64  `json:"minimumDifficulty"`
	NetworkDiscovery          bool   `json:"networkDiscovery"`
	NetworkID                 int64  `json:"networkId"`
	PriceUpdatePeriod         string `json:"priceUpdatePeriod"`
	RefuseServiceTransactions bool   `json:"refuseServiceTransactions"`
	RelaySet                  string `json:"relaySet"`
	RemoveSolved              bool   `json:"removeSolved"`
	ResealMaxPeriod           int64  `json:"resealMaxPeriod"`
	ResealMinPeriod           int64  `json:"resealMinPeriod"`
	ResealOnTxs               string `json:"resealOnTxs"`
	Signature                 string `json:"signature"`    //POA
	Step                      int64  `json:"step"`         //POA
	StepDuration              int64  `json:"stepDuration"` //POA
	TxGasLimit                string `json:"txGasLimit"`
	TxQueueGas                string `json:"txQueueGas"`
	TxQueueSize               int64  `json:"txQueueSize"`
	TxQueueStrategy           string `json:"txQueueStrategy"`
	TxTimeLimit               int64  `json:"txTimeLimit"`
	USDPerEth                 string `json:"usdPerEth"`
	USDPerTX                  string `json:"usdPerTx"`
	ValidateChainIDTransition int64  `json:"validateChainIdTransition"`
	WorkQueueSize             int64  `json:"workQueueSize"`
}

/**
 * Fills in the defaults for missing parts,
 */
func newConf(data map[string]interface{}) (*parityConf, error) {
	out := new(parityConf)
	err := json.Unmarshal([]byte(GetDefaults()), out)
	fmt.Printf("%+v\n", *out)
	if data == nil {
		log.Println(err)
		return out, err
	}
	tmp, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = json.Unmarshal(tmp, out)

	return out, err
}

// GetParams fetchs parity related parameters
func GetParams() string {
	dat, err := helpers.GetStaticBlockchainConfig("parity", "params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetDefaults fetchs parity related parameter defaults
func GetDefaults() string {
	dat, err := helpers.GetStaticBlockchainConfig("parity", "defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetServices returns the services which are used by artemis
func GetServices() []util.Service {
	return []util.Service{
		{
			Name:  "Geth",
			Image: "gcr.io/whiteblock/ethereum:latest",
			Env:   nil,
		},
	}
}


func buildConfig(pconf *parityConf, details *db.DeploymentDetails, wallets []string, passwordFile string, node int) (string, error) {

	dat, err := helpers.GetBlockchainConfig("parity", node, "config.toml.template", details)
	if err != nil {
		log.Println(err)
		return "", err
	}
	var tmp map[string]interface{}

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
	mp["networkId"] = fmt.Sprintf("%d", pconf.NetworkID)
	return mustache.Render(string(dat), mp)
}

func buildPoaSpec(pconf *parityConf, details *db.DeploymentDetails, wallets []string) (string, error) {

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
		"stepDuration":              pconf.StepDuration,
		"validators":                validators,
		"difficulty":                fmt.Sprintf("0x%x", pconf.Difficulty),
		"gasLimit":                  fmt.Sprintf("0x%x", pconf.GasLimit),
		"networkId":                 fmt.Sprintf("0x%x", pconf.NetworkID),
		"maximumExtraDataSize":      fmt.Sprintf("0x%x", pconf.MaximumExtraDataSize),
		"minGasLimit":               fmt.Sprintf("0x%x", pconf.MinGasLimit),
		"gasLimitBoundDivisor":      fmt.Sprintf("0x%x", pconf.GasLimitBoundDivisor),
		"validateChainIdTransition": pconf.ValidateChainIDTransition,
		"eip155Transition":          pconf.EIP155Transition,
		"eip140Transition":          pconf.EIP140Transition,
		"eip211Transition":          pconf.EIP211Transition,
		"eip214Transition":          pconf.EIP214Transition,
		"eip658Transition":          pconf.EIP658Transition,
		"accounts":                  accounts,
	}
	filler := util.ConvertToStringMap(tmp)
	dat, err := helpers.GetBlockchainConfig("parity", 0, "spec.json.poa.mustache", details)
	if err != nil {
		return "", err
	}
	return mustache.Render(string(dat), filler)
}

func buildSpec(pconf *parityConf, details *db.DeploymentDetails, wallets []string) (string, error) {

	accounts := make(map[string]interface{})
	for _, wallet := range wallets {
		accounts[wallet] = map[string]interface{}{
			"balance": pconf.InitBalance,
		}
	}

	tmp := map[string]interface{}{
		"minimumDifficulty":      fmt.Sprintf("0x%x", pconf.MinimumDifficulty),
		"difficultyBoundDivisor": fmt.Sprintf("0x%x", pconf.DifficultyBoundDivisor),
		"durationLimit":          fmt.Sprintf("0x%x", pconf.DurationLimit),
		"blockReward":            fmt.Sprintf("0x%x", pconf.BlockReward),
		"difficulty":             fmt.Sprintf("0x%x", pconf.Difficulty),
		"gasLimit":               fmt.Sprintf("0x%x", pconf.GasLimit),
		"networkId":              fmt.Sprintf("0x%x", pconf.NetworkID),
		"maximumExtraDataSize":   fmt.Sprintf("0x%x", pconf.MaximumExtraDataSize),
		"minGasLimit":            fmt.Sprintf("0x%x", pconf.MinGasLimit),
		"gasLimitBoundDivisor":   fmt.Sprintf("0x%x", pconf.GasLimitBoundDivisor),
		"accounts":               accounts,
	}
	filler := util.ConvertToStringMap(tmp)
	dat, err := helpers.GetBlockchainConfig("parity", 0, "spec.json.mustache", details)
	if err != nil {
		return "", err
	}
	return mustache.Render(string(dat), filler)
}

func gethSpec(pconf *parityConf, wallets []string) (string, error) {
	accounts := make(map[string]interface{})
	for _, wallet := range wallets {
		accounts[wallet] = map[string]interface{}{
			"balance": pconf.InitBalance,
		}
	}

	tmp := map[string]interface{}{
		"chainId":        pconf.NetworkID,
		"difficulty":     fmt.Sprintf("0x%x", pconf.Difficulty),
		"gasLimit":       fmt.Sprintf("0x%x", pconf.GasLimit),
		"homesteadBlock": 0,
		"eip155Block":    10,
		"eip158Block":    10,
		"alloc":          accounts,
	}
	filler := util.ConvertToStringMap(tmp)
	dat, err := helpers.GetStaticBlockchainConfig("geth", "genesis.json")
	if err != nil {
		return "", err
	}
	data, err := mustache.Render(string(dat), filler)
	return data, err
}

/*
   passwordFile
   unlock
*/
func buildPoaConfig(pconf *parityConf, details *db.DeploymentDetails, wallets []string, passwordFile string, i int) (string, error) {

	dat, err := helpers.GetBlockchainConfig("parity", i, "config.toml.poa.mustache", details)
	if err != nil {
		log.Println(err)
		return "", err
	}
	var tmp map[string]interface{}

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
	mp["networkId"] = fmt.Sprintf("%d", pconf.NetworkID)
	mp["signer"] = fmt.Sprintf("\"%s\"", wallets[i])
	return mustache.Render(string(dat), mp)
}
