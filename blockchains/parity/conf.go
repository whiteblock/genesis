package parity

import (
    "encoding/json"
    "github.com/Whiteblock/mustache"
    "io/ioutil"
    util "../../util"
    //"strconv"
)

type ParityConf struct {
    ForceSealing                bool    `json:"forceSealing"`
    ResealOnTxs                 int64   `json:"resealOnTxs"`
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

    TxQueueStrategy             string  `json:"txQueueStrategy"`
    TxQueueBanCount             int64   `json:"txQueueBanCount"`
    TxQueueBanTime              int64   `json:"txQueueBanTime"`
    TxGasLimit                  string  `json:"txGasLimit"`
    TxTimeLimit                 int64   `json:"txTimeLimit"`
    RemoveSolved                bool    `json:"removeSolved"`
    RefuseServiceTransactions   bool    `json:"refuseServiceTransactions"`
    EnableIPFS                  bool    `json:"enableIPFS"`
    NetworkDiscovery            bool    `json:"networkDiscovery"`

    out.ForceSealing, err = util.GetJSONBool(data, "forceSealing")
    out.ResealOnTxs, err = util.GetJSONInt64(data, "resealOnTxs")
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


    return out, nil
}

func GetParams() string {
    return `[
    ["forceSealing","bool"],
    ["resealOnTxs","int"],
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
    ["networkDiscovery","bool"]
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
    "networkDiscovery":true
}`
}

func GetServices() []util.Service {
    return nil
}


func BuildConfig(pconf *ParityConf) (string,error) {
    //unlock
    //passwordFile

    dat, err := ioutil.ReadFile("./blockchains/parity/config.toml.template")
    if err != nil {
        return "",err
    }
    data, err := mustache.Render(dat, map[string]string{"c": "world"})
    return data,err
}