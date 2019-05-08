/*
	Copyright 2019 Whiteblock Inc.
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

package eos

import (
	"../../db"
	"../../util"
	"../helpers"
	"encoding/json"
	"fmt"
	"github.com/Whiteblock/mustache"
	"time"
)

type eosConf struct {
	UserAccounts                   int64    `json:"userAccounts"`
	BlockProducers                 int64    `json:"validators"`
	AccountCPUStake                int64    `json:"accountCpuStake"`
	AccountRAM                     int64    `json:"accountRam"`
	AccountNetStake                int64    `json:"accountNetStake"`
	AccountFunds                   int64    `json:"accountFunds"`
	BpCPUStake                     int64    `json:"bpCpuStake"`
	BpNetStake                     int64    `json:"bpNetStake"`
	BpRAM                          int64    `json:"bpRam"`
	BpFunds                        int64    `json:"bpFunds"`
	MaxBlockNetUsage               int64    `json:"maxBlockNetUsage"`
	TargetBlockNetUsagePct         int64    `json:"targetBlockNetUsagePct"`
	MaxTransactionNetUsage         int64    `json:"maxTransactionNetUsage"`
	BasePerTransactionNetUsage     int64    `json:"basePerTransactionNetUsage"`
	NetUsageLeeway                 int64    `json:"netUsageLeeway"`
	ContextFreeDiscountNetUsageNum int64    `json:"contextFreeDiscountNetUsageNum"`
	ContextFreeDiscountNetUsageDen int64    `json:"contextFreeDiscountNetUsageDen"`
	MaxBlockCPUUsage               int64    `json:"maxBlockCpuUsage"`
	TargetBlockCPUUsagePct         int64    `json:"targetBlockCpuUsagePct"`
	MaxTransactionCPUUsage         int64    `json:"maxTransactionCpuUsage"`
	MinTransactionCPUUsage         int64    `json:"minTransactionCpuUsage"`
	MaxTransactionLifetime         int64    `json:"maxTransactionLifetime"`
	DeferredTrxExpirationWindow    int64    `json:"deferredTrxExpirationWindow"`
	MaxTransactionDelay            int64    `json:"maxTransactionDelay"`
	MaxInlineActionSize            int64    `json:"maxInlineActionSize"`
	MaxInlineActionDepth           int64    `json:"maxInlineActionDepth"`
	MaxAuthorityDepth              int64    `json:"maxAuthorityDepth"`
	InitialChainID                 string   `json:"initialChainId"`
	ChainStateDbSizeMb             int64    `json:"chainStateDbSizeMb"`
	ReversibleBlocksDbSizeMb       int64    `json:"reversibleBlocksDbSizeMb"`
	ContractsConsole               bool     `json:"contractsConsole"`
	P2pMaxNodesPerHost             int64    `json:"p2pMaxNodesPerHost"`
	AllowedConnection              string   `json:"allowedConnection"`
	MaxClients                     int64    `json:"maxClients"`
	ConnectionCleanupPeriod        int64    `json:"connectionCleanupPeriod"`
	NetworkVersionMatch            int64    `json:"networkVersionMatch"`
	SyncFetchSpan                  int64    `json:"syncFetchSpan"`
	PauseOnStartup                 bool     `json:"pauseOnStartup"`
	MaxTransactionTime             int64    `json:"maxTransactionTime"`
	MaxIrreversibleBlockAge        int64    `json:"maxIrreversibleBlockAge"`
	KeosdProviderTimeout           int64    `json:"keosdProviderTimeout"`
	TxnReferenceBlockLag           int64    `json:"txnReferenceBlockLag"`
	Plugins                        []string `json:"plugins"`
	ConfigExtras                   []string `json:"configExtras"`
}

func newConf(data map[string]interface{}) (*eosConf, error) {
	out := new(eosConf)
	json.Unmarshal([]byte(GetDefaults()), out)
	if data == nil {
		return out, nil
	}

	err := util.GetJSONInt64(data, "userAccounts", &out.UserAccounts)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "validators", &out.BlockProducers)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "maxBlockNetUsage", &out.MaxBlockNetUsage)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "targetBlockNetUsagePct", &out.TargetBlockNetUsagePct)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "maxTransactionNetUsage", &out.MaxTransactionNetUsage)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "basePerTransactionNetUsage", &out.BasePerTransactionNetUsage)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "netUsageLeeway", &out.NetUsageLeeway)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "contextFreeDiscountNetUsageNum", &out.ContextFreeDiscountNetUsageNum)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "contextFreeDiscountNetUsageDen", &out.ContextFreeDiscountNetUsageDen)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "maxBlockCpuUsage", &out.MaxBlockCPUUsage)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "targetBlockCpuUsagePct", &out.TargetBlockCPUUsagePct)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "maxTransactionCpuUsage", &out.MaxTransactionCPUUsage)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "minTransactionCpuUsage", &out.MinTransactionCPUUsage)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "maxTransactionLifetime", &out.MaxTransactionLifetime)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "deferredTrxExpirationWindow", &out.DeferredTrxExpirationWindow)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "maxTransactionDelay", &out.MaxTransactionDelay)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "maxInlineActionSize", &out.MaxInlineActionSize)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "maxInlineActionDepth", &out.MaxInlineActionDepth)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "maxAuthorityDepth", &out.MaxAuthorityDepth)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONString(data, "initialChainId", &out.InitialChainID)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "chainStateDbSizeMb", &out.ChainStateDbSizeMb)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "reversibleBlocksDbSizeMb", &out.ReversibleBlocksDbSizeMb)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONBool(data, "contractsConsole", &out.ContractsConsole)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "p2pMaxNodesPerHost", &out.P2pMaxNodesPerHost)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONString(data, "allowedConnection", &out.AllowedConnection)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "maxClients", &out.MaxClients)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "connectionCleanupPeriod", &out.ConnectionCleanupPeriod)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "networkVersionMatch", &out.NetworkVersionMatch)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "syncFetchSpan", &out.SyncFetchSpan)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONBool(data, "pauseOnStartup", &out.PauseOnStartup)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "maxTransactionTime", &out.MaxTransactionTime)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "maxIrreversibleBlockAge", &out.MaxIrreversibleBlockAge)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "keosdProviderTimeout", &out.KeosdProviderTimeout)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "txnReferenceBlockLag", &out.TxnReferenceBlockLag)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONStringArr(data, "plugins", &out.Plugins)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONStringArr(data, "configExtras", &out.ConfigExtras)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "accountCpuStake", &out.AccountCPUStake)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "accountRam", &out.AccountRAM)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "accountNetStake", &out.AccountNetStake)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "accountFunds", &out.AccountFunds)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "bpCpuStake", &out.BpCPUStake)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "bpRam", &out.BpRAM)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "bpNetStake", &out.BpNetStake)
	if err != nil {
		return nil, err
	}

	err = util.GetJSONInt64(data, "bpFunds", &out.BpFunds)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (econf *eosConf) GenerateGenesis(masterPublicKey string, details *db.DeploymentDetails) (string, error) {

	filler := util.ConvertToStringMap(map[string]interface{}{
		"initialTimestamp":               time.Now().Format("2006-01-02T15-04-05.000"),
		"initialKey":                     masterPublicKey,
		"maxBlockNetUsage":               econf.MaxBlockNetUsage,
		"targetBlockNetUsagePct":         econf.TargetBlockNetUsagePct,
		"maxTransactionNetUsage":         econf.MaxTransactionNetUsage,
		"basePerTransactionNetUsage":     econf.BasePerTransactionNetUsage,
		"netUsageLeeway":                 econf.NetUsageLeeway,
		"contextFreeDiscountNetUsageNum": econf.ContextFreeDiscountNetUsageNum,
		"contextFreeDiscountNetUsageDen": econf.ContextFreeDiscountNetUsageDen,
		"maxBlockCpuUsage":               econf.MaxBlockCPUUsage,
		"targetBlockCpuUsagePct":         econf.TargetBlockCPUUsagePct,
		"maxTransactionCpuUsage":         econf.MaxTransactionCPUUsage,
		"minTransactionCpuUsage":         econf.MinTransactionCPUUsage,
		"maxTransactionLifetime":         econf.MaxTransactionLifetime,
		"deferredTrxExpirationWindow":    econf.DeferredTrxExpirationWindow,
		"maxTransactionDelay":            econf.MaxTransactionDelay,
		"maxInlineActionSize":            econf.MaxInlineActionSize,
		"maxInlineActionDepth":           econf.MaxInlineActionDepth,
		"maxAuthorityDepth":              econf.MaxAuthorityDepth,
		"initialChainId":                 econf.InitialChainID,
	})

	dat, err := helpers.GetBlockchainConfig("eos", 0, "genesis.json.mustache", details)
	if err != nil {
		return "", util.LogError(err)
	}
	return mustache.Render(string(dat), filler)
}

func (econf *eosConf) GenerateConfig() string {

	out := []string{
		"bnet-endpoint = 0.0.0.0:4321",
		"bnet-no-trx = false",
		"blocks-dir = /datadir/blocks/",
		fmt.Sprintf("chain-state-db-size-mb = %d", econf.ChainStateDbSizeMb),
		fmt.Sprintf("reversible-blocks-db-size-mb = %d", econf.ReversibleBlocksDbSizeMb),
		fmt.Sprintf("contracts-console = %v", econf.ContractsConsole),
		"https-client-validate-peers = 0",
		fmt.Sprintf("p2p-max-nodes-per-host = %d", econf.P2pMaxNodesPerHost),
		fmt.Sprintf("allowed-connection = %s", econf.AllowedConnection),
		fmt.Sprintf("max-clients = %d", econf.MaxClients),
		fmt.Sprintf("connection-cleanup-period = %d", econf.ConnectionCleanupPeriod),
		fmt.Sprintf("network-version-match = %d", econf.NetworkVersionMatch),
		fmt.Sprintf("sync-fetch-span = %d", econf.SyncFetchSpan),
		fmt.Sprintf("pause-on-startup = %v", econf.PauseOnStartup),
		fmt.Sprintf("max-transaction-time = %d", econf.MaxTransactionTime),
		fmt.Sprintf("max-irreversible-block-age = %d", econf.MaxIrreversibleBlockAge),
		fmt.Sprintf("keosd-provider-timeout = %d", econf.KeosdProviderTimeout),
		fmt.Sprintf("txn-reference-block-lag = %d", econf.TxnReferenceBlockLag),

		"access-control-allow-credentials = false",
		"http-server-address = 0.0.0.0:8889",
		"p2p-listen-endpoint = 0.0.0.0:8999",
	}
	for _, plugin := range econf.Plugins {
		out = append(out, "plugin = "+plugin)
	}
	for _, extra := range econf.ConfigExtras {
		out = append(out, extra)
	}
	return util.CombineConfig(out)
}

// GetDefaults fetchs eos related parameter defaults
func GetDefaults() string {
	dat, err := helpers.GetStaticBlockchainConfig("eos", "defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetParams fetchs eos related parameters
func GetParams() string {
	dat, err := helpers.GetStaticBlockchainConfig("eos", "params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetServices returns the services which are used by eos
func GetServices() []util.Service {
	return nil
}
