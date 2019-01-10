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
    ValidatorCount          int64   `json:"validatorCount"`
    SigAlgorithm            string  `json:"sigAlgorithm"`
    Command                 string  `json:"command"`
}

func NewRChainConf(data map[string]interface{}) (*RChainConf,error) {
    out := new(RChainConf)
    json.Unmarshal([]byte(GetDefaults()),out)
    if data == nil {
        return out,nil
    }

    var err error

    if _,ok := data["noUpnp"]; ok {
        out.NoUpnp,err = util.GetJSONBool(data,"noUpnp")
        if err != nil {
            return nil,err
        }
    }
    
    if _,ok := data["defaultTimeout"]; ok {
        out.DefaultTimeout,err = util.GetJSONInt64(data,"defaultTimeout")
        if err != nil {
            return nil,err
        }
    }

    if _,ok := data["mapSize"]; ok {
        out.MapSize,err = util.GetJSONInt64(data,"mapSize")
        if err != nil {
            return nil,err
        }
    }

    if _,ok := data["casperBlockStoreSize"]; ok {
        out.CasperBlockStoreSize,err = util.GetJSONInt64(data,"casperBlockStoreSize")
        if err != nil {
            return nil,err
        }
    }

    if _,ok := data["inMemoryStore"]; ok {
        out.InMemoryStore,err = util.GetJSONBool(data,"inMemoryStore")
        if err != nil {
            return nil,err
        }
    }
    
    if _,ok := data["maxNumOfConnections"]; ok {
        out.MaxNumOfConnections,err = util.GetJSONInt64(data,"maxNumOfConnections")
        if err != nil {
            return nil,err
        }
    }

    if _,ok := data["validatorCount"]; ok {
        out.ValidatorCount,err = util.GetJSONInt64(data,"validatorCount")
        if err != nil {
            return nil,err
        }
    }

    if _,ok := data["sigAlgorithm"]; ok {
        out.SigAlgorithm,err = util.GetJSONString(data,"sigAlgorithm")
        if err != nil {
            return nil,err
        }
    }

    if _,ok := data["command"]; ok {
        out.Command,err = util.GetJSONString(data,"command")
        if err != nil {
            return nil,err
        }
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
    {"noUpnp":"bool"},
    {"defaultTimeout":"int"},
    {"mapSize":"int"},
    {"casperBlockStoreSize":"int"},
    {"inMemoryStore":"bool"},
    {"maxNumOfConnections":"int"},
    {"validatorCount":"int"},
    {"sigAlgorithm":"string"},
    {"command":"string"}
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
    "sigAlgorithm":"ed25519",
    "command":"/opt/docker/bin/rnode"
}`
}
