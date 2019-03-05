package parity

import (
    "encoding/json"
    "github.com/Whiteblock/mustache"
    "io/ioutil"
    util "../../util"
    "fmt"
    "log"
    //"strconv"
)

type ParityConf struct {
    ForceSealing                bool    `json:"forceSealing"`
    ResealOnTxs                 string  `json:"resealOnTxs"`
    ResealMinPeriod             int64   `json:"resealMinPeriod"`
    ResealMaxPeriod             int64   `json:"resealMaxPeriod"`
    WorkQueueSize               int64   `json:"workQueueSize"`
    RelaySet                    string  `json:"relaySet"`
    UsdPerTx                    string  `json:"usdPerTx"`
    UsdPerEth                   string  `json:"usdPerEth"`
    PriceUpdatePeriod           string  `json:"priceUpdatePeriod"`
    GasFloorTarget              string  `json:"gasFloorTarget"`
    GasCap                      string  `json:"gasCap"`
    TxQueueSize                 int64   `json:"txQueueSize"`
    TxQueueGas                  string  `json:"txQueueGas"`
    TxQueueStrategy             string  `json:"txQueueStrategy"`
    TxGasLimit                  string  `json:"txGasLimit"`
    TxTimeLimit                 int64   `json:"txTimeLimit"`
    RemoveSolved                bool    `json:"removeSolved"`
    RefuseServiceTransactions   bool    `json:"refuseServiceTransactions"`
    EnableIPFS                  bool    `json:"enableIPFS"`
    NetworkDiscovery            bool    `json:"networkDiscovery"`
    ExtraAccounts               int64   `json:"extraAccounts"`
    ChainId                     int64   `json:"chainId"`
    NetworkId                   int64   `json:"networkId"`
    Difficulty                  int64   `json:"difficulty"`
    InitBalance                 string  `json:"initBalance"`
    MaxPeers                    int64   `json:"maxPeers"`
    GasLimit                    int64   `json:"gasLimit"`
    HomesteadBlock              int64   `json:"homesteadBlock"`
    Eip155Block                 int64   `json:"eip155Block"`
    Eip158Block                 int64   `json:"eip158Block"`
    MinimumDifficulty           int64   `json:"minimumDifficulty"`
    DifficultyBoundDivisor      int64   `json:"difficultyBoundDivisor"`
    DurationLimit               int64   `json:"durationLimit"`
    BlockReward                 int64   `json:"blockReward"`
    MaximumExtraDataSize        int64   `json:"maximumExtraDataSize"`
    MinGasLimit                 int64   `json:"minGasLimit"`
    GasLimitBoundDivisor        int64   `json:"gasLimitBoundDivisor"`
}

/**
 * Fills in the defaults for missing parts,
 */
func NewConf(data map[string]interface{}) (*ParityConf, error) {
    out := new(ParityConf)
    err := json.Unmarshal([]byte(GetDefaults()),out)
    fmt.Printf("%+v\n",*out)
    if data == nil {
        log.Println(err)
        return out,err
    }
    err = util.GetJSONBool(data, "forceSealing", &out.ForceSealing)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONString(data, "resealOnTxs",&out.ResealOnTxs)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "resealMinPeriod",&out.ResealMinPeriod)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "resealMaxPeriod",&out.ResealMaxPeriod)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "workQueueSize",&out.WorkQueueSize)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONString(data, "relaySet",&out.RelaySet)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONString(data, "usdPerTx",&out.UsdPerTx)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONString(data, "usdPerEth",&out.UsdPerEth)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONString(data, "priceUpdatePeriod",&out.PriceUpdatePeriod)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONString(data, "gasFloorTarget",&out.GasFloorTarget)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONString(data, "gasCap",&out.GasCap)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "txQueueSize",&out.TxQueueSize)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONString(data, "txQueueGas",&out.TxQueueGas)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONString(data, "txQueueStrategy",&out.TxQueueStrategy)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONString(data, "txGasLimit",&out.TxGasLimit)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "txTimeLimit",&out.TxTimeLimit)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONBool(data, "removeSolved",&out.RemoveSolved)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONBool(data, "refuseServiceTransactions",&out.RefuseServiceTransactions)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONBool(data, "enableIPFS",&out.EnableIPFS)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONBool(data, "networkDiscovery",&out.NetworkDiscovery)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "extraAccounts",&out.ExtraAccounts)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "chainId",&out.ChainId)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "networkId",&out.NetworkId)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "difficulty",&out.Difficulty)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONString(data, "initBalance",&out.InitBalance)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "maxPeers",&out.MaxPeers)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "gasLimit",&out.GasLimit)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "homesteadBlock",&out.HomesteadBlock)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "eip155Block",&out.Eip155Block)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "eip158Block",&out.Eip158Block)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "minimumDifficulty",&out.MinimumDifficulty)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "difficultyBoundDivisor",&out.DifficultyBoundDivisor)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "durationLimit",&out.DurationLimit)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "blockReward",&out.BlockReward)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "maximumExtraDataSize",&out.MaximumExtraDataSize)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "minGasLimit",&out.MinGasLimit)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data, "gasLimitBoundDivisor",&out.GasLimitBoundDivisor)
    if err != nil {
        return nil,err
    }
    
    return out, nil
}

func GetParams() string {
    dat, err := ioutil.ReadFile("./resources/parity/params.json")
    if err != nil {
        panic(err)//Missing required files is a fatal error
    }
    return string(dat)
}
func GetDefaults() string {
    dat, err := ioutil.ReadFile("./resources/parity/defaults.json")
    if err != nil {
        panic(err)//Missing required files is a fatal error
    }
    return string(dat)
}

func GetServices() []util.Service {
    return []util.Service{
        util.Service{
            Name: "Geth",
            Image:"gcr.io/whiteblock/ethereum:latest",
            Env:nil,
        },
    }
}

/*
    passwordFile
    unlock
 */
func BuildConfig(pconf *ParityConf,wallets []string,passwordFile string) (string,error) {

    dat, err := ioutil.ReadFile("./resources/parity/config.toml.template")
    if err != nil {
        log.Println(err)
        return "",err
    }
    var tmp interface{}

    raw,err := json.Marshal(*pconf)
    if err != nil {
        log.Println(err)
        return "",err
    }

    err = json.Unmarshal(raw,&tmp)
    if err != nil {
        log.Println(err)
        return "",err
    }

    mp := util.ConvertToStringMap(tmp)
    raw,err = json.Marshal(wallets);
    if err != nil {
        log.Println(err)
        return "",err
    }
    mp["unlock"] = string(raw)
    mp["passwordFile"] = fmt.Sprintf("[\"%s\"]",passwordFile)
    mp["networkId"] = fmt.Sprintf("%d",pconf.NetworkId)
    data, err := mustache.Render(string(dat),mp)
    return data,err
}

func BuildSpec(pconf *ParityConf, wallets []string) (string,error) {

    accounts := make(map[string]interface{})
    for _,wallet := range wallets {
        accounts[wallet] = map[string]interface{}{
            "balance": pconf.InitBalance,
        }
    }

    tmp := map[string]interface{}{
        "minimumDifficulty":fmt.Sprintf("0x%x",pconf.MinimumDifficulty),
        "difficultyBoundDivisor":fmt.Sprintf("0x%x",pconf.DifficultyBoundDivisor),
        "durationLimit":fmt.Sprintf("0x%x",pconf.DurationLimit),
        "blockReward":fmt.Sprintf("0x%x",pconf.BlockReward),
        "difficulty":fmt.Sprintf("0x%x",pconf.Difficulty),
        "gasLimit":fmt.Sprintf("0x%x",pconf.GasLimit),
        "networkId":fmt.Sprintf("0x%x",pconf.NetworkId),
        "maximumExtraDataSize":fmt.Sprintf("0x%x",pconf.MaximumExtraDataSize),
        "minGasLimit":fmt.Sprintf("0x%x",pconf.MinGasLimit),
        "gasLimitBoundDivisor":fmt.Sprintf("0x%x",pconf.GasLimitBoundDivisor),
        "accounts":accounts,
    }
    filler := util.ConvertToStringMap(tmp)
    dat, err := ioutil.ReadFile("./resources/parity/spec.json.mustache")
    if err != nil {
        return "",err
    }
    data, err := mustache.Render(string(dat), filler)
    return data,err
}

func GethSpec(pconf *ParityConf,wallets []string) (string,error) {
    accounts := make(map[string]interface{})
    for _,wallet := range wallets {
        accounts[wallet] = map[string]interface{}{
            "balance": pconf.InitBalance,
        }
    }

    tmp := map[string]interface{}{
        "chainId":pconf.NetworkId,
        "difficulty":fmt.Sprintf("0x%x",pconf.Difficulty),
        "gasLimit":fmt.Sprintf("0x%x",pconf.GasLimit),
        "alloc":accounts,       
    }
    filler := util.ConvertToStringMap(tmp)
    dat, err := ioutil.ReadFile("./resources/geth/genesis.json")
    if err != nil {
        return "",err
    }
    data, err := mustache.Render(string(dat), filler)
    return data,err
}