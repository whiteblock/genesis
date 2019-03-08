package eos

import(
    "encoding/json"
    "fmt"
    "time"
    "io/ioutil"
    "github.com/Whiteblock/mustache"
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

func (this *EosConf) GenerateGenesis(masterPublicKey string) (string,error) {

    filler := util.ConvertToStringMap(map[string]interface{}{
        "initialTimestamp": time.Now().Format("2006-01-02T15-04-05.000"),
        "initialKey":masterPublicKey,
        "maxBlockNetUsage":this.MaxBlockNetUsage,
        "targetBlockNetUsagePct":this.TargetBlockNetUsagePct,
        "maxTransactionNetUsage":this.MaxTransactionNetUsage,
        "basePerTransactionNetUsage":this.BasePerTransactionNetUsage,
        "netUsageLeeway":this.NetUsageLeeway,
        "contextFreeDiscountNetUsageNum":this.ContextFreeDiscountNetUsageNum,
        "contextFreeDiscountNetUsageDen":this.ContextFreeDiscountNetUsageDen,
        "maxBlockCpuUsage":this.MaxBlockCpuUsage,
        "targetBlockCpuUsagePct":this.TargetBlockCpuUsagePct,
        "maxTransactionCpuUsage":this.MaxTransactionCpuUsage,
        "minTransactionCpuUsage":this.MinTransactionCpuUsage,
        "maxTransactionLifetime":this.MaxTransactionLifetime,
        "deferredTrxExpirationWindow":this.DeferredTrxExpirationWindow,
        "maxTransactionDelay":this.MaxTransactionDelay,
        "maxInlineActionSize":this.MaxInlineActionSize,
        "maxInlineActionDepth":this.MaxInlineActionDepth,
        "maxAuthorityDepth":this.MaxAuthorityDepth,
        "initialChainId":this.InitialChainId,
    })
    dat, err := ioutil.ReadFile("./resources/eos/genesis.json.mustache")
    if err != nil {
        return "",err
    }
    data, err := mustache.Render(string(dat), filler)
    return data,err
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
    for _,plugin := range this.Plugins {
        out = append(out,"plugin = " + plugin)
    }
    for _,extra := range this.ConfigExtras {
        out = append(out,extra)
    }
    return util.CombineConfig(out)
}

func GetDefaults() string {

    dat, err := ioutil.ReadFile("./resources/eos/defaults.json")
    if err != nil {
        panic(err)//Missing required files is a fatal error
    }
    return string(dat)
}

func GetParams() string {
    dat, err := ioutil.ReadFile("./resources/eos/params.json")
    if err != nil {
        panic(err)//Missing required files is a fatal error
    }
    return string(dat)
}


func GetServices() []util.Service {
    return nil
}
