/*
	Copyright 2019 whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Genesis is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package parity

import (
	"encoding/json"
	"fmt"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/util"
	"github.com/whiteblock/mustache"
)

type parityConf struct {
	Name                      string `json:"name"`
	DataDir                   string `json:"dataDir"`
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
	EIP150Transition          int64  `json:"eip150Transition"`
	EIP155Transition          int64  `json:"eip155Transition"`
	EIP160Transition          int64  `json:"eip160Transition"`
	EIP161ABCTransition       int64  `json:"eip161abcTransition"`
	EIP161DTransition        int64  `json:"eip161dTransition"`
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
	return out, helpers.HandleBlockchainConfig(blockchain, data, out)
}

func NewParityConf(data map[string]interface{}) (*parityConf, error) {
	out := new(parityConf)
	return out, helpers.HandleBlockchainConfig(blockchain, data, out)
}

// GetServices returns the services which are used by parity
func GetServices() []helpers.Service {
	return []helpers.Service{
		helpers.SimpleService{
			Name:  "Geth",
			Image: "gcr.io/whiteblock/ethereum:latest",
			Env:   nil,
		},
	}
}

func buildConfig(pconf *parityConf, details *db.DeploymentDetails, wallets []string, passwordFile string, node int) (string, error) {

	dat, err := helpers.GetBlockchainConfig("parity", node, "config.toml.template", details)
	if err != nil {
		return "", util.LogError(err)
	}
	var tmp map[string]interface{}

	raw, err := json.Marshal(*pconf)
	if err != nil {
		return "", util.LogError(err)
	}

	err = json.Unmarshal(raw, &tmp)
	if err != nil {
		return "", util.LogError(err)
	}

	mp := util.ConvertToStringMap(tmp)
	raw, err = json.Marshal(wallets)
	if err != nil {
		return "", util.LogError(err)
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
		"eip140Transition":          pconf.EIP140Transition,
		"eip150Transition":          pconf.EIP150Transition,
		"eip155Transition":          pconf.EIP155Transition,
		"eip160Transition":          pconf.EIP160Transition,
		"eip161abcTransition":       pconf.EIP161ABCTransition,
		"eip161dTransition":         pconf.EIP161DTransition,
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
		return "", util.LogError(err)
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
		return "", util.LogError(err)
	}
	data, err := mustache.Render(string(dat), filler)
	return data, util.LogError(err)
}

/*
   passwordFile
   unlock
*/
func buildPoaConfig(pconf *parityConf, details *db.DeploymentDetails, wallets []string, passwordFile string, i int) (string, error) {

	dat, err := helpers.GetBlockchainConfig("parity", i, "config.toml.poa.mustache", details)
	if err != nil {
		return "", util.LogError(err)
	}
	var tmp map[string]interface{}

	raw, err := json.Marshal(*pconf)
	if err != nil {
		return "", util.LogError(err)
	}

	err = json.Unmarshal(raw, &tmp)
	if err != nil {
		return "", util.LogError(err)
	}

	mp := util.ConvertToStringMap(tmp)
	raw, err = json.Marshal(wallets)
	if err != nil {
		return "", util.LogError(err)
	}
	mp["unlock"] = string(raw)
	mp["passwordFile"] = fmt.Sprintf("[\"%s\"]", passwordFile)
	mp["networkId"] = fmt.Sprintf("%d", pconf.NetworkID)
	mp["signer"] = fmt.Sprintf("\"%s\"", wallets[i])
	return mustache.Render(string(dat), mp)
}
