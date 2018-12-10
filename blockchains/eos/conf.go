package eos

import(
	"encoding/json"
)

type EosConf struct{
	UserAccounts					int64	`json:"userAccounts"`
	BlockProducers					int64	`json:"blockProducers"`
	
	MaxBlockNetUsage				int64	`json:"maxBlockNetUsage"`
	TargetBlockNetUsagePct			int64	`json:"targetBlockNetUsagePct"`
	MaxTransactionNetUsage			int64	`json:"maxTransactionNetUsage"`
	BasePerTransactionNetUsage		int64	`json:"basePerTransactionNetUsage"`
	NetUsageLeeway					int64	`json:"netUsageLeeway"`
	ContextFreeDiscountNetUsageNum	int64	`json:"contextFreeDiscountNetUsageNum"`
	ContextFreeDiscountNetUsageDen	int64	`json:"contextFreeDiscountNetUsageDen"`
	MaxBlockCpuUsage				int64	`json:"maxBlockCpuUsage"`
	TargetBlockCpuUsagePct			int64	`json:"targetBlockCpuUsagePct"`
	MaxTransactionCpuUsage			int64	`json:"maxTransactionCpuUsage"`
	MinTransactionCpuUsage			int64	`json:"minTransactionCpuUsage"`
	MaxTransactionLifetime			int64	`json:"maxTransactionLifetime"`
	DeferredTrxExpirationWindow		int64	`json:"deferredTrxExpirationWindow"`
	MaxTransactionDelay				int64	`json:"maxTransactionDelay"`
	MaxInlineActionSize				int64	`json:"maxInlineActionSize"`
	MaxInlineActionDepth			int64	`json:"maxInlineActionDepth"`
	MaxAuthorityDepth				int64	`json:"maxAuthorityDepth"`
	InitialChainId					string	`json:"initialChainId"`

	ChainStateDbSizeMb				int64	`json:"chainStateDbSizeMb"`
	ReversibleBlocksDbSizeMb		int64	`json:"reversibleBlocksDbSizeMb"`
	ContractsConsole				bool	`json:"contractsConsole"`
	P2pMaxNodesPerHost				int64	`json:"p2pMaxNodesPerHost"`
	AllowedConnection				string	`json:"allowedConnection"`
	MaxClients						int64	`json:"maxClients"`
	ConnectionCleanupPeriod			int64	`json:"connectionCleanupPeriod"`
	NetworkVersionMatch				int64	`json:"networkVersionMatch"`
	SyncFetchSpan					int64	`json:"syncFetchSpan"`
	MaxImplicitRequest				int64	`json:"maxImplicitRequest"`
	PauseOnStartup					bool	`json:"pauseOnStartup"`
	MaxTransactionTime				int64	`json:"maxTransactionTime"`
	MaxIrreversibleBlockAge			int64	`json:"maxIrreversibleBlockAge"`
	KeosdProviderTimeout			int64	`json:"keosdProviderTimeout"`
	TxnReferenceBlockLag			int64	`json:"txnReferenceBlockLag"`
}

func NewConf(data map[string]interface{}) (*EosConf,error){
	out := new(EosConf)
	json.Unmarshal([]byte(GetDefaults()),out)
	if data == nil {
		return out,nil
	}
	return out,nil
}

func GetDefaults() string{
	return `{
	"userAccounts":100,
	"blockProducers":21,
	"maxBlockNetUsage":1048576,
	"targetBlockNetUsagePct":1000,
	"maxTransactionNetUsage":524288,
	"basePerTransactionNetUsage":12,
	"netUsageLeeway":500,
	"contextFreeDiscountNetUsageNum":20,
	"contextFreeDiscountNetUsageDen":100,
	"maxBlockCpuUsage":100000,
	"targetBlockCpuUsagePct":500,
	"maxTransactionCpuUsage":50000,
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
	"keosdProviderTimeout":5,
	"txnReferenceBlockLag":0

}`
}

func GetParams() string{
	return `[
		{"userAccounts":"int"},

	]`
}