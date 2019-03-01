package cosmos

import(
    "io/ioutil"
    util "../../util"
)

func GetParams() string {
    dat, err := ioutil.ReadFile("./resources/cosmos/params.json")
    if err != nil {
        panic(err)//Missing required files is a fatal error
    }
    return string(dat)
}

func GetDefaults() string {
    dat, err := ioutil.ReadFile("./resources/cosmos/defaults.json")
    if err != nil {
        panic(err)//Missing required files is a fatal error
    }
    return string(dat)
}

func GetServices() []util.Service {
    return nil
}