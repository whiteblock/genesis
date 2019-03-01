package rchain

import (
    "io/ioutil"
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
    dat, err := ioutil.ReadFile("./resources/rchain/params.json")
    if err != nil {
        panic(err)//Missing required files is a fatal error
    }
    return string(dat)
}

func GetDefaults() string {
    dat, err := ioutil.ReadFile("./resources/rchain/defaults.json")
    if err != nil {
        panic(err)//Missing required files is a fatal error
    }
    return string(dat)
}