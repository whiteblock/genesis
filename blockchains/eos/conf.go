package eos

import(
	"encoding/json"
	"fmt"
	util "../../util"
)

type EosConf struct{
	UserAccounts					int64		`json:"userAccounts"`
	BlockProducers					int			`json:"blockProducers"`
	
	MaxBlockNetUsage				int64		`json:"maxBlockNetUsage"`
	TargetBlockNetUsagePct			int64		`json:"targetBlockNetUsagePct"`
	MaxTransactionNetUsage			int64		`json:"maxTransactionNetUsage"`
	BasePerTransactionNetUsage		int64		`json:"basePerTransactionNetUsage"`
	NetUsageLeeway					int64		`json:"netUsageLeeway"`
	ContextFreeDiscountNetUsageNum	int64		`json:"contextFreeDiscountNetUsageNum"`
	ContextFreeDiscountNetUsageDen	int64		`json:"contextFreeDiscountNetUsageDen"`
	MaxBlockCpuUsage				int64		`json:"maxBlockCpuUsage"`
	TargetBlockCpuUsagePct			int64		`json:"targetBlockCpuUsagePct"`
	MaxTransactionCpuUsage			int64		`json:"maxTransactionCpuUsage"`
	MinTransactionCpuUsage			int64		`json:"minTransactionCpuUsage"`
	MaxTransactionLifetime			int64		`json:"maxTransactionLifetime"`
	DeferredTrxExpirationWindow		int64		`json:"deferredTrxExpirationWindow"`
	MaxTransactionDelay				int64		`json:"maxTransactionDelay"`
	MaxInlineActionSize				int64		`json:"maxInlineActionSize"`
	MaxInlineActionDepth			int64		`json:"maxInlineActionDepth"`
	MaxAuthorityDepth				int64		`json:"maxAuthorityDepth"`
	InitialChainId					string		`json:"initialChainId"`

	ChainStateDbSizeMb				int64		`json:"chainStateDbSizeMb"`
	ReversibleBlocksDbSizeMb		int64		`json:"reversibleBlocksDbSizeMb"`
	ContractsConsole				bool		`json:"contractsConsole"`
	P2pMaxNodesPerHost				int64		`json:"p2pMaxNodesPerHost"`
	AllowedConnection				string		`json:"allowedConnection"`
	MaxClients						int64		`json:"maxClients"`
	ConnectionCleanupPeriod			int64		`json:"connectionCleanupPeriod"`
	NetworkVersionMatch				int64		`json:"networkVersionMatch"`
	SyncFetchSpan					int64		`json:"syncFetchSpan"`
	MaxImplicitRequest				int64		`json:"maxImplicitRequest"`
	PauseOnStartup					bool		`json:"pauseOnStartup"`
	MaxTransactionTime				int64		`json:"maxTransactionTime"`
	MaxIrreversibleBlockAge			int64		`json:"maxIrreversibleBlockAge"`
	KeosdProviderTimeout			int64		`json:"keosdProviderTimeout"`
	TxnReferenceBlockLag			int64		`json:"txnReferenceBlockLag"`
	Plugins							[]string	`json:"plugins"`
	ConfigExtras					[]string	`json:"configExtras"`
}

func NewConf(data map[string]interface{}) (*EosConf,error){
	out := new(EosConf)
	json.Unmarshal([]byte(GetDefaults()),out)
	if data == nil {
		return out,nil
	}
	var err error
	if _,ok := data["userAccounts"]; ok {
		out.UserAccounts,err = util.GetJSONInt64(data,"userAccounts")
		if err != nil {
			return nil,err
		}
	}
	
	if _,ok := data["blockProducers"]; ok {
		num,err := util.GetJSONInt64(data,"blockProducers")
		if err != nil {
			return nil,err
		}
		out.BlockProducers = int(num)
	}

	if _,ok := data["maxBlockNetUsage"]; ok {
		out.MaxBlockNetUsage,err = util.GetJSONInt64(data,"maxBlockNetUsage")
		if err != nil {
			return nil,err
		}
	}
	if _,ok := data["targetBlockNetUsagePct"]; ok {
		out.TargetBlockNetUsagePct,err = util.GetJSONInt64(data,"targetBlockNetUsagePct")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["maxTransactionNetUsage"]; ok {
		out.MaxTransactionNetUsage,err = util.GetJSONInt64(data,"maxTransactionNetUsage")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["basePerTransactionNetUsage"]; ok {
		out.BasePerTransactionNetUsage,err = util.GetJSONInt64(data,"basePerTransactionNetUsage")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["netUsageLeeway"]; ok {
		out.NetUsageLeeway,err = util.GetJSONInt64(data,"netUsageLeeway")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["contextFreeDiscountNetUsageNum"]; ok {
		out.ContextFreeDiscountNetUsageNum,err = util.GetJSONInt64(data,"contextFreeDiscountNetUsageNum")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["contextFreeDiscountNetUsageDen"]; ok {
		out.ContextFreeDiscountNetUsageDen,err = util.GetJSONInt64(data,"contextFreeDiscountNetUsageDen")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["maxBlockCpuUsage"]; ok {
		out.MaxBlockCpuUsage,err = util.GetJSONInt64(data,"maxBlockCpuUsage")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["targetBlockCpuUsagePct"]; ok {
		out.TargetBlockCpuUsagePct,err = util.GetJSONInt64(data,"targetBlockCpuUsagePct")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["maxTransactionCpuUsage"]; ok {
		out.MaxTransactionCpuUsage,err = util.GetJSONInt64(data,"maxTransactionCpuUsage")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["minTransactionCpuUsage"]; ok {
		out.MinTransactionCpuUsage,err = util.GetJSONInt64(data,"minTransactionCpuUsage")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["maxTransactionLifetime"]; ok {
		out.MaxTransactionLifetime,err = util.GetJSONInt64(data,"maxTransactionLifetime")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["deferredTrxExpirationWindow"]; ok {
		out.DeferredTrxExpirationWindow,err = util.GetJSONInt64(data,"deferredTrxExpirationWindow")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["maxTransactionDelay"]; ok {
		out.MaxTransactionDelay,err = util.GetJSONInt64(data,"maxTransactionDelay")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["maxInlineActionSize"]; ok {
		out.MaxInlineActionSize,err = util.GetJSONInt64(data,"maxInlineActionSize")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["maxInlineActionDepth"]; ok {
		out.MaxInlineActionDepth,err = util.GetJSONInt64(data,"maxInlineActionDepth")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["maxAuthorityDepth"]; ok {
		out.MaxAuthorityDepth,err = util.GetJSONInt64(data,"maxAuthorityDepth")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["initialChainId"]; ok {
		out.InitialChainId,err = util.GetJSONString(data,"initialChainId")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["chainStateDbSizeMb"]; ok {
		out.ChainStateDbSizeMb,err = util.GetJSONInt64(data,"chainStateDbSizeMb")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["reversibleBlocksDbSizeMb"]; ok {
		out.ReversibleBlocksDbSizeMb,err = util.GetJSONInt64(data,"reversibleBlocksDbSizeMb")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["contractsConsole"]; ok {
		out.ContractsConsole,err = util.GetJSONBool(data,"contractsConsole")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["p2pMaxNodesPerHost"]; ok {
		out.P2pMaxNodesPerHost,err = util.GetJSONInt64(data,"p2pMaxNodesPerHost")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["allowedConnection"]; ok {
		out.AllowedConnection,err = util.GetJSONString(data,"allowedConnection")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["maxClients"]; ok {
		out.MaxClients,err = util.GetJSONInt64(data,"maxClients")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["connectionCleanupPeriod"]; ok {
		out.ConnectionCleanupPeriod,err = util.GetJSONInt64(data,"connectionCleanupPeriod")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["networkVersionMatch"]; ok {
		out.NetworkVersionMatch,err = util.GetJSONInt64(data,"networkVersionMatch")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["syncFetchSpan"]; ok {
		out.SyncFetchSpan,err = util.GetJSONInt64(data,"syncFetchSpan")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["maxImplicitRequest"]; ok {
		out.MaxImplicitRequest,err = util.GetJSONInt64(data,"maxImplicitRequest")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["pauseOnStartup"]; ok {
		out.PauseOnStartup,err = util.GetJSONBool(data,"pauseOnStartup")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["maxTransactionTime"]; ok {
		out.MaxTransactionTime,err = util.GetJSONInt64(data,"maxTransactionTime")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["maxIrreversibleBlockAge"]; ok {
		out.MaxIrreversibleBlockAge,err = util.GetJSONInt64(data,"maxIrreversibleBlockAge")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["keosdProviderTimeout"]; ok {
		out.KeosdProviderTimeout,err = util.GetJSONInt64(data,"keosdProviderTimeout")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["txnReferenceBlockLag"]; ok {
		out.TxnReferenceBlockLag,err = util.GetJSONInt64(data,"txnReferenceBlockLag")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["plugins"]; ok {
		out.Plugins,err = util.GetJSONStringArr(data,"plugins")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["configExtras"]; ok {
		out.ConfigExtras,err = util.GetJSONStringArr(data,"configExtras")
		if err != nil {
			return nil,err
		}
	}

	return out,nil
}

func (this *EosConf) GenerateGenesis(masterPublicKey string) string {
	return fmt.Sprintf (
`{
	"initial_timestamp": "2018-12-07T12:11:00.000",
	"initial_key": "%s",
	"initial_configuration": {
		"max_block_net_usage": %d,
		"target_block_net_usage_pct": %d,
		"max_transaction_net_usage": %d,
		"base_per_transaction_net_usage": %d,
		"net_usage_leeway": %d,
		"context_free_discount_net_usage_num": %d,
		"context_free_discount_net_usage_den": %d,
		"max_block_cpu_usage": %d,
		"target_block_cpu_usage_pct": %d,
		"max_transaction_cpu_usage": %d,
		"min_transaction_cpu_usage": %d,
		"max_transaction_lifetime": %d,
		"deferred_trx_expiration_window": %d,
		"max_transaction_delay": %d,
		"max_inline_action_size": %d,
		"max_inline_action_depth": %d,
		"max_authority_depth": %d
	},
	"initial_chain_id": "%s"
}`,
	masterPublicKey,
	this.MaxBlockNetUsage,
	this.TargetBlockNetUsagePct,
	this.MaxTransactionNetUsage,
	this.BasePerTransactionNetUsage,
	this.NetUsageLeeway,
	this.ContextFreeDiscountNetUsageNum,
	this.ContextFreeDiscountNetUsageDen,
	this.MaxBlockCpuUsage,
	this.TargetBlockCpuUsagePct,
	this.MaxTransactionCpuUsage,
	this.MinTransactionCpuUsage,
	this.MaxTransactionLifetime,
	this.DeferredTrxExpirationWindow,
	this.MaxTransactionDelay,
	this.MaxInlineActionSize,
	this.MaxInlineActionDepth,
	this.MaxAuthorityDepth,
	this.InitialChainId)
}

func (this *EosConf) GenerateConfig() string {
	out := []string{
		"bnet-endpoint = 0.0.0.0:4321",
		"bnet-no-trx = false",
		"blocks-dir = /datadir/blocks/",
		fmt.Sprintf("chain-state-db-size-mb = %d",this.ChainStateDbSizeMb),
		fmt.Sprintf("reversible-blocks-db-size-mb = %d",this.ReversibleBlocksDbSizeMb),
		fmt.Sprintf("contracts-console = %v",this.ContractsConsole),
		"https-client-validate-peers = 0",
		fmt.Sprintf("p2p-max-nodes-per-host = %d",this.P2pMaxNodesPerHost),
		fmt.Sprintf("allowed-connection = %s",this.AllowedConnection),
		fmt.Sprintf("max-clients = %d",this.MaxClients),
		fmt.Sprintf("connection-cleanup-period = %d",this.ConnectionCleanupPeriod),
		fmt.Sprintf("network-version-match = %d",this.NetworkVersionMatch),
		fmt.Sprintf("sync-fetch-span = %d",this.SyncFetchSpan),
		fmt.Sprintf("max-implicit-request = %d",this.MaxImplicitRequest),
		fmt.Sprintf("pause-on-startup = %v",this.PauseOnStartup),
		fmt.Sprintf("max-transaction-time = %d",this.MaxTransactionTime),
		fmt.Sprintf("max-irreversible-block-age = %d",this.MaxIrreversibleBlockAge),
		fmt.Sprintf("keosd-provider-timeout = %d",this.KeosdProviderTimeout),
		fmt.Sprintf("txn-reference-block-lag = %d",this.TxnReferenceBlockLag),
		
		"access-control-allow-credentials = false",
		"http-server-address = 0.0.0.0:8889",
		"p2p-listen-endpoint = 0.0.0.0:8999",
	}
	for _,plugin := range this.Plugins{
		out = append(out,"plugin = " + plugin)
	}
	for _,extra := range this.ConfigExtras{
		out = append(out,extra)
	}
	return util.CombineConfig(out)
}

func GetDefaults() string{
	return `{
	"userAccounts":200,
	"blockProducers":21,
	"maxBlockNetUsage":1048576,
	"targetBlockNetUsagePct":1000,
	"maxTransactionNetUsage":524288,
	"basePerTransactionNetUsage":12,
	"netUsageLeeway":500,
	"contextFreeDiscountNetUsageNum":20,
	"contextFreeDiscountNetUsageDen":100,
	"maxBlockCpuUsage":1000000,
	"targetBlockCpuUsagePct":500,
	"maxTransactionCpuUsage":500000,
	"minTransactionCpuUsage":100,
	"maxTransactionLifetime":3600,
	"deferredTrxExpirationWindow":600,
	"maxTransactionDelay":3888000,
	"maxInlineActionSize":4096,
	"maxInlineActionDepth":4,
	"maxAuthorityDepth":6,
	"initialChainId":"6469636b627574740a",
	"chainStateDbSizeMb":8192,
	"reversibleBlocksDbSizeMb":340,
	"contractsConsole":false,
	"p2pMaxNodesPerHost":4,
	"allowedConnection":"any",
	"maxClients":0,
	"connectionCleanupPeriod":30,
	"syncFetchSpan":100,
	"maxImplicitRequest":1500,
	"pauseOnStartup":false,
	"maxTransactionTime":100,
	"maxIrreversibleBlockAge":1000000,
	"keosdProviderTimeout":5,
	"txnReferenceBlockLag":0,
	"plugins":[
		"eosio::chain_plugin",
		"eosio::chain_api_plugin",
		"eosio::producer_plugin",
		"eosio::http_plugin",
		"eosio::history_api_plugin",
		"eosio::net_plugin",
		"eosio::net_api_plugin"
	]
}`
}


func GetParams() string{
	return `[
	{"userAccounts":"int"},
	{"blockProducers":"int"},
	{"maxBlockNetUsage":"int"},
	{"targetBlockNetUsagePct":"int"},
	{"maxTransactionNetUsage":"int"},
	{"basePerTransactionNetUsage":"int"},
	{"netUsageLeeway":"int"},
	{"contextFreeDiscountNetUsageNum":"int"},
	{"contextFreeDiscountNetUsageDen":"int"},
	{"maxBlockCpuUsage":"int"},
	{"targetBlockCpuUsagePct":"int"},
	{"maxTransactionCpuUsage":"int"},
	{"minTransactionCpuUsage":"int"},
	{"maxTransactionLifetime":"int"},
	{"deferredTrxExpirationWindow":"int"},
	{"maxTransactionDelay":"int"},
	{"maxInlineActionSize":"int"},
	{"maxInlineActionDepth":"int"},
	{"maxAuthorityDepth":"int"},
	{"initialChainId":"string"},
	{"chainStateDbSizeMb":"int"},
	{"reversibleBlocksDbSizeMb":"int"},
	{"contractsConsole":"bool"},
	{"p2pMaxNodesPerHost":"int"},
	{"allowedConnection":"string"},
	{"maxClients":"int"},
	{"connectionCleanupPeriod":"int"},
	{"syncFetchSpan":"int"},
	{"maxImplicitRequest":"int"},
	{"pauseOnStartup":"bool"},
	{"maxTransactionTime":"int"},
	{"maxIrreversibleBlockAge":"int"},
	{"keosdProviderTimeout":"int"},
	{"txnReferenceBlockLag":"int"},
	{"plugins":"[]string"}
]`
}