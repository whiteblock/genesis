package eos

import(
    "encoding/json"
    "fmt"
    "time"
    util "../../util"
)

type EosConf struct{
    UserAccounts                    int64       `json:"userAccounts"`
    BlockProducers                  int64       `json:"validators"`

    AccountCpuStake                 int64       `json:"accountCpuStake"`
    AccountRam                      int64       `json:"accountRam"`
    AccountNetStake                 int64       `json:"accountNetStake"`
    AccountFunds                    int64       `json:"accountFunds"`

    BpCpuStake                      int64       `json:"bpCpuStake"`
    BpNetStake                      int64       `json:"bpNetStake"`
    BpRam                           int64       `json:"bpRam"`
    BpFunds                         int64       `json:"bpFunds"`

    MaxBlockNetUsage                int64       `json:"maxBlockNetUsage"`
    TargetBlockNetUsagePct          int64       `json:"targetBlockNetUsagePct"`
    MaxTransactionNetUsage          int64       `json:"maxTransactionNetUsage"`
    BasePerTransactionNetUsage      int64       `json:"basePerTransactionNetUsage"`
    NetUsageLeeway                  int64       `json:"netUsageLeeway"`
    ContextFreeDiscountNetUsageNum  int64       `json:"contextFreeDiscountNetUsageNum"`
    ContextFreeDiscountNetUsageDen  int64       `json:"contextFreeDiscountNetUsageDen"`
    MaxBlockCpuUsage                int64       `json:"maxBlockCpuUsage"`
    TargetBlockCpuUsagePct          int64       `json:"targetBlockCpuUsagePct"`
    MaxTransactionCpuUsage          int64       `json:"maxTransactionCpuUsage"`
    MinTransactionCpuUsage          int64       `json:"minTransactionCpuUsage"`
    MaxTransactionLifetime          int64       `json:"maxTransactionLifetime"`
    DeferredTrxExpirationWindow     int64       `json:"deferredTrxExpirationWindow"`
    MaxTransactionDelay             int64       `json:"maxTransactionDelay"`
    MaxInlineActionSize             int64       `json:"maxInlineActionSize"`
    MaxInlineActionDepth            int64       `json:"maxInlineActionDepth"`
    MaxAuthorityDepth               int64       `json:"maxAuthorityDepth"`
    InitialChainId                  string      `json:"initialChainId"`

    ChainStateDbSizeMb              int64       `json:"chainStateDbSizeMb"`
    ReversibleBlocksDbSizeMb        int64       `json:"reversibleBlocksDbSizeMb"`
    ContractsConsole                bool        `json:"contractsConsole"`
    P2pMaxNodesPerHost              int64       `json:"p2pMaxNodesPerHost"`
    AllowedConnection               string      `json:"allowedConnection"`
    MaxClients                      int64       `json:"maxClients"`
    ConnectionCleanupPeriod         int64       `json:"connectionCleanupPeriod"`
    NetworkVersionMatch             int64       `json:"networkVersionMatch"`
    SyncFetchSpan                   int64       `json:"syncFetchSpan"`
//  MaxImplicitRequest              int64       `json:"maxImplicitRequest"`
    PauseOnStartup                  bool        `json:"pauseOnStartup"`
    MaxTransactionTime              int64       `json:"maxTransactionTime"`
    MaxIrreversibleBlockAge         int64       `json:"maxIrreversibleBlockAge"`
    KeosdProviderTimeout            int64       `json:"keosdProviderTimeout"`
    TxnReferenceBlockLag            int64       `json:"txnReferenceBlockLag"`
    Plugins                         []string    `json:"plugins"`
    ConfigExtras                    []string    `json:"configExtras"`
}

func NewConf(data map[string]interface{}) (*EosConf,error){
    out := new(EosConf)
    json.Unmarshal([]byte(GetDefaults()),out)
    if data == nil {
        return out,nil
    }

    err := util.GetJSONInt64(data,"userAccounts",&out.UserAccounts)
    if err != nil {
        return nil,err
    }
    
    err = util.GetJSONInt64(data,"validators",&out.BlockProducers)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data,"maxBlockNetUsage",&out.MaxBlockNetUsage)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data,"targetBlockNetUsagePct",&out.TargetBlockNetUsagePct)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"maxTransactionNetUsage",&out.MaxTransactionNetUsage)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"basePerTransactionNetUsage",&out.BasePerTransactionNetUsage)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"netUsageLeeway",&out.NetUsageLeeway)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"contextFreeDiscountNetUsageNum",&out.ContextFreeDiscountNetUsageNum)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"contextFreeDiscountNetUsageDen",&out.ContextFreeDiscountNetUsageDen)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"maxBlockCpuUsage",&out.MaxBlockCpuUsage)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"targetBlockCpuUsagePct",&out.TargetBlockCpuUsagePct)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"maxTransactionCpuUsage",&out.MaxTransactionCpuUsage)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"minTransactionCpuUsage",&out.MinTransactionCpuUsage)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"maxTransactionLifetime",&out.MaxTransactionLifetime)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"deferredTrxExpirationWindow",&out.DeferredTrxExpirationWindow)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"maxTransactionDelay",&out.MaxTransactionDelay)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"maxInlineActionSize",&out.MaxInlineActionSize)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"maxInlineActionDepth",&out.MaxInlineActionDepth)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"maxAuthorityDepth",&out.MaxAuthorityDepth)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONString(data,"initialChainId",&out.InitialChainId)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"chainStateDbSizeMb",&out.ChainStateDbSizeMb)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"reversibleBlocksDbSizeMb",&out.ReversibleBlocksDbSizeMb)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONBool(data,"contractsConsole",&out.ContractsConsole)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"p2pMaxNodesPerHost",&out.P2pMaxNodesPerHost)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONString(data,"allowedConnection",&out.AllowedConnection)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"maxClients",&out.MaxClients)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"connectionCleanupPeriod",&out.ConnectionCleanupPeriod)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"networkVersionMatch",&out.NetworkVersionMatch)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"syncFetchSpan",&out.SyncFetchSpan)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONBool(data,"pauseOnStartup",&out.PauseOnStartup)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"maxTransactionTime",&out.MaxTransactionTime)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"maxIrreversibleBlockAge",&out.MaxIrreversibleBlockAge)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"keosdProviderTimeout",&out.KeosdProviderTimeout)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"txnReferenceBlockLag",&out.TxnReferenceBlockLag)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONStringArr(data,"plugins",&out.Plugins)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONStringArr(data,"configExtras",&out.ConfigExtras)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"accountCpuStake",&out.AccountCpuStake)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"accountRam",&out.AccountRam)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"accountNetStake",&out.AccountNetStake)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"accountFunds",&out.AccountFunds)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"bpCpuStake",&out.BpCpuStake)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"bpRam",&out.BpRam)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"bpNetStake",&out.BpNetStake)
    if err != nil {
        return nil, err
    }

    err = util.GetJSONInt64(data,"bpFunds",&out.BpFunds)
    if err != nil {
        return nil, err
    }

/*  if _,ok := data["maxImplicitRequest"]; ok {
        out.MaxImplicitRequest,err = util.GetJSONInt64(data,"maxImplicitRequest")
        if err != nil {
            return nil,err
        }
    }*/


    return out,nil
}

func (this *EosConf) GenerateGenesis(masterPublicKey string) string {
    return fmt.Sprintf (
`{
    "initial_timestamp": "%s",
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
    time.Now().Format("2006-01-02T15-04-05.000"),
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
//      fmt.Sprintf("max-implicit-request = %d",this.MaxImplicitRequest),
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
    "validators":21,
    "accountCpuStake":2000000,
    "accountRam":32768,
    "accountNetStake":500000,
    "accountFunds":100000,
    "bpCpuStake":1000000,
    "bpNetStake":1000000,
    "bpRam":32768,
    "bpFunds":100000,
    "maxBlockNetUsage":1048576,
    "targetBlockNetUsagePct":1000,
    "maxTransactionNetUsage":524288,
    "basePerTransactionNetUsage":12,
    "netUsageLeeway":500,
    "contextFreeDiscountNetUsageNum":20,
    "contextFreeDiscountNetUsageDen":100,
    "maxBlockCpuUsage":10000000,
    "targetBlockCpuUsagePct":500,
    "maxTransactionCpuUsage":5000000,
    "minTransactionCpuUsage":100,
    "maxTransactionLifetime":36000,
    "deferredTrxExpirationWindow":1000,
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
    {"accountCPUStake":"int"},
    {"accountRAMStake":"int"},
    {"accountNetStake":"int"},
    {"accountFunds":"int"},
    {"bpCpuStake":"int"},
    {"bpNetStake":"int"},
    {"bpRamStake":"int"},
    {"bpFunds":"int"},
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


func GetServices() []util.Service {
    return nil
}
