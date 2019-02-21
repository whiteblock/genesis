package rchain

import (
    "encoding/json"
    "encoding/base64"
    util "../../util"
)

type RChainConf struct {
    NoUpnp                  bool    `json:"noUpnp"`
    DefaultTimeout          int64   `json:"defaultTimeout"`
    MapSize                 int64   `json:"mapSize"`
    CasperBlockStoreSize    int64   `json:"casperBlockStoreSize"`
    InMemoryStore           bool    `json:"inMemoryStore"`
    MaxNumOfConnections     int64   `json:"maxNumOfConnections"`
    Validators              int64   `json:"validators"`
    ValidatorCount          int64   `json:"validatorCount"`
    SigAlgorithm            string  `json:"sigAlgorithm"`
    Command                 string  `json:"command"`
}

func NewRChainConf(data map[string]interface{}) (*RChainConf,error) {
    out := new(RChainConf)
    err := json.Unmarshal([]byte(GetDefaults()),out)
    if data == nil {
        return out,err
    }

    err = util.GetJSONBool(data,"noUpnp",&out.NoUpnp)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data,"defaultTimeout",&out.DefaultTimeout)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data,"mapSize",&out.MapSize)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data,"casperBlockStoreSize",&out.CasperBlockStoreSize)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONBool(data,"inMemoryStore",&out.InMemoryStore)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data,"maxNumOfConnections",&out.MaxNumOfConnections)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data,"validators",&out.Validators)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONInt64(data,"validatorCount",&out.ValidatorCount)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONString(data,"sigAlgorithm",&out.SigAlgorithm)
    if err != nil {
        return nil,err
    }

    err = util.GetJSONString(data,"command",&out.Command)
    if err != nil {
        return nil,err
    }
    
    return out, nil
}

func GetServices() []util.Service {
    return []util.Service{
        util.Service{
            Name:"wb_influx_proxy",
            Image:"gcr.io/wb-genesis/bitbucket.org/whiteblockio/influx-proxy:master",
            Env:map[string]string{
                "BASIC_AUTH_BASE64":base64.StdEncoding.EncodeToString([]byte(conf.InfluxUser+":"+conf.InfluxPassword)),
                "INFLUXDB_URL":conf.Influx,
                "BIND_PORT":"8086",
            },
        },
    }
}

func GetParams() string {
    return `[
    ["noUpnp","bool"],
    ["defaultTimeout","int"],
    ["mapSize","int"],
    ["casperBlockStoreSize","int"],
    ["inMemoryStore","bool"],
    ["maxNumOfConnections","int"],
    ["validatorCount","int"],
    ["validators","int"],
    ["sigAlgorithm","string"],
    ["command","string"]
]`
}

func GetDefaults() string {
    return `{
    "noUpnp":true,
    "defaultTimeout":2000,
    "mapSize":1073741824,
    "casperBlockStoreSize":1073741824,
    "inMemoryStore":false,
    "maxNumOfConnections":500,
    "validatorCount":5,
    "validators":0,
    "sigAlgorithm":"ed25519",
    "command":"/rchain/node/target/rnode-0.8.2/usr/share/rnode/bin/rnode"
}`
}
