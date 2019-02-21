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
    ResealOnTxs                 string   `json:"resealOnTxs"`
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
    TxQueueBanCount             int64   `json:"txQueueBanCount"`
    TxQueueBanTime              int64   `json:"txQueueBanTime"`
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
    HomesteadTransition         int64   `json:"homesteadTransition"`
    Eip150Transition            int64   `json:"eip150Transition"`
    Eip160Transition            int64   `json:"eip160Transition"`
    Eip161abcTransition         int64   `json:"eip161abcTransition"`
    Eip161dTransition           int64   `json:"eip161dTransition"`
    MaxCodeSize                 int64   `json:"maxCodeSize"`
    MaximumExtraDataSize        int64   `json:"maximumExtraDataSize"`
    MinGasLimit                 int64   `json:"minGasLimit"`
}

/**
 * Fills in the defaults for missing parts,
 */
func NewConf(data map[string]interface{}) (*ParityConf, error) {
    out := new(ParityConf)
    err := json.Unmarshal([]byte(GetDefaults()),out)
    if data == nil {
        return out,err
    }
    rawjson,err := json.Marshal(data)
    err = json.Unmarshal(rawjson,out)
    /*out.ForceSealing, err = util.GetJSONBool(data, "forceSealing")
    out.ResealOnTxs, err = util.GetJSONString(data, "resealOnTxs")
    out.ResealMinPeriod, err = util.GetJSONInt64(data, "resealMinPeriod")
    out.ResealMaxPeriod, err = util.GetJSONInt64(data, "resealMaxPeriod")
    out.WorkQueueSize, err = util.GetJSONInt64(data, "workQueueSize")
    out.RelaySet, err = util.GetJSONString(data, "relaySet")
    out.UsdPerTx, err = util.GetJSONString(data, "usdPerTx")
    out.UsdPerEth, err = util.GetJSONString(data, "usdPerEth")
    out.PriceUpdatePeriod, err = util.GetJSONString(data, "priceUpdatePeriod")
    out.GasFloorTarget, err = util.GetJSONString(data, "gasFloorTarget")
    out.GasCap, err = util.GetJSONString(data, "gasCap")
    out.TxQueueSize, err = util.GetJSONInt64(data, "txQueueSize")
    out.TxQueueGas, err = util.GetJSONString(data, "txQueueGas")
    out.TxQueueStrategy, err = util.GetJSONString(data, "txQueueStrategy")
    out.TxQueueBanCount, err = util.GetJSONInt64(data, "txQueueBanCount")
    out.TxQueueBanTime, err = util.GetJSONInt64(data, "txQueueBanTime")
    out.TxGasLimit, err = util.GetJSONString(data, "txGasLimit")
    out.TxTimeLimit, err = util.GetJSONInt64(data, "txTimeLimit")
    out.RemoveSolved, err = util.GetJSONBool(data, "removeSolved")
    out.RefuseServiceTransactions, err = util.GetJSONBool(data, "refuseServiceTransactions")
    out.EnableIPFS, err = util.GetJSONBool(data, "enableIPFS")
    out.NetworkDiscovery, err = util.GetJSONBool(data, "networkDiscovery")
    out.ExtraAccounts, err = util.GetJSONInt64(data, "extraAccounts")
    out.ChainId, err = util.GetJSONInt64(data, "chainId")
    out.NetworkId, err = util.GetJSONInt64(data, "networkId")
    out.Difficulty, err = util.GetJSONInt64(data, "difficulty")
    out.InitBalance, err = util.GetJSONString(data, "initBalance")
    out.MaxPeers, err = util.GetJSONInt64(data, "maxPeers")
    out.GasLimit, err = util.GetJSONInt64(data, "gasLimit")
    out.HomesteadBlock, err = util.GetJSONInt64(data, "homesteadBlock")
    out.Eip155Block, err = util.GetJSONInt64(data, "eip155Block")
    out.Eip158Block, err = util.GetJSONInt64(data, "eip158Block")
    out.MinimumDifficulty, err = util.GetJSONInt64(data, "minimumDifficulty")
    out.DifficultyBoundDivisor, err = util.GetJSONInt64(data, "difficultyBoundDivisor")
    out.DurationLimit, err = util.GetJSONInt64(data, "durationLimit")
    out.BlockReward, err = util.GetJSONInt64(data, "blockReward")
    out.HomesteadTransition, err = util.GetJSONInt64(data, "homesteadTransition")
    out.Eip150Transition, err = util.GetJSONInt64(data, "eip150Transition")
    out.Eip160Transition, err = util.GetJSONInt64(data, "eip160Transition")
    out.Eip161abcTransition, err = util.GetJSONInt64(data, "eip161abcTransition")
    out.Eip161dTransition, err = util.GetJSONInt64(data, "eip161dTransition")
    out.MaxCodeSize, err = util.GetJSONInt64(data, "maxCodeSize")
    out.MaximumExtraDataSize, err = util.GetJSONInt64(data, "maximumExtraDataSize")
    out.MinGasLimit, err = util.GetJSONInt64(data, "minGasLimit")*/

    return out, nil
}

func GetParams() string {
    return `[
    ["forceSealing","bool"],
    ["resealOnTxs","string"],
    ["resealMinPeriod","int"],
    ["resealMaxPeriod","int"],
    ["workQueueSize","int"],
    ["relaySet","string"],
    ["usdPerTx","string"],
    ["usdPerEth","string"],
    ["priceUpdatePeriod","string"],
    ["gasFloorTarget","string"],
    ["gasCap","string"],
    ["txQueueSize","int"],
    ["txQueueGas","string"],
    ["txQueueStrategy","string"],
    ["txQueueBanCount","int"],
    ["txQueueBanTime","int"],
    ["txGasLimit","string"],
    ["txTimeLimit","int"],
    ["removeSolved","bool"],
    ["refuseServiceTransactions","bool"],
    ["enableIPFS","bool"],
    ["networkDiscovery","bool"],
    ["extraAccounts","int"],
    ["chainId","int"],
    ["networkId","int"],
    ["difficulty","int"],
    ["initBalance","string"],
    ["maxPeers","int"],
    ["gasLimit","int"],
    ["homesteadBlock","int"],
    ["eip155Block","int"],
    ["eip158Block","int"],

    ["minimumDifficulty","int"],
    ["difficultyBoundDivisor","int"],
    ["durationLimit","int"],
    ["blockReward","int"],
    ["homesteadTransition","int"],
    ["eip150Transition","int"],
    ["eip160Transition","int"],
    ["eip161abcTransition","int"],
    ["eip161dTransition","int"],
    ["maxCodeSize","int"],
    ["maximumExtraDataSize","int"],
    ["minGasLimit","int"],
]`
}

func GetDefaults() string {
    return `{
    "forceSealing":true,
    "resealOnTxs":"all",
    "resealMinPeriod":4000,
    "resealMaxPeriod":60000,
    "workQueueSize":20,
    "relaySet":"cheap",
    "usdPerTx":"0.0025",
    "usdPerEth":"auto",
    "priceUpdatePeriod":"hourly",
    "gasFloorTarget":"4700000",
    "gasCap":"6283184",
    "txQueueSize":8192,
    "txQueueGas":"off",
    "txQueueStrategy":"gas_factor",
    "txQueueBanCount":1,
    "txQueueBanTime":180,
    "txGasLimit":"6283184",
    "txTimeLimit":100,
    "removeSolved":false,
    "refuseServiceTransactions":false,
    "enableIPFS":false,
    "networkDiscovery":true,
    "extraAccounts":0,
    "chainId":15468,
    "networkId":15468,
    "difficulty":100000,
    "initBalance":100000000000000000000,
    "maxPeers":1000,
    "gasLimit":4000000,
    "homesteadBlock":0,
    "eip155Block":0,
    "eip158Block":0,
    "minimumDifficulty":131072,
    "difficultyBoundDivisor":2048,
    "durationLimit":13,
    "blockReward":5000000000000000000,
    "homesteadTransition":0,
    "eip150Transition":0,
    "eip160Transition":10,
    "eip161abcTransition":10,
    "eip161dTransition":10,
    "maxCodeSize":24576,
    "maximumExtraDataSize":32,
    "minGasLimit":5000
}`
}

func GetServices() []util.Service {
    return nil
}

/*
    passwordFile
    unlock
 */
func BuildConfig(pconf *ParityConf,wallets []string,passwordFile string) (string,error) {

    dat, err := ioutil.ReadFile("./blockchains/parity/config.toml.template")
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
    mp["passwordFile"] = fmt.Sprintf("[\"%s\"]",passwordFile);;
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
        "homesteadTransition":pconf.HomesteadTransition,
        "eip150Transition":pconf.Eip150Transition,
        "eip160Transition":pconf.Eip160Transition,
        "eip161abcTransition":pconf.Eip161abcTransition,
        "eip161dTransition":pconf.Eip161dTransition,
        "maxCodeSize":pconf.MaxCodeSize,
        "difficulty":fmt.Sprintf("0x%x",pconf.Difficulty),
        "gasLimit":fmt.Sprintf("0x%x",pconf.GasLimit),
        "networkID":fmt.Sprintf("0x%x",pconf.NetworkId),
        "maximumExtraDataSize":fmt.Sprintf("0x%x",pconf.MaximumExtraDataSize),
        "minGasLimit":fmt.Sprintf("0x%x",pconf.MinGasLimit),
        "accounts":accounts,
    }
    filler := util.ConvertToStringMap(tmp)
    dat, err := ioutil.ReadFile("./blockchains/parity/spec.json.mustache")
    if err != nil {
        return "",err
    }
    data, err := mustache.Render(string(dat), filler)
    return data,err
}