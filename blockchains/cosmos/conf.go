package cosmos

import (
	util "../../util"
)

func GetParams() string {
	dat, err := util.GetBlockchainConfig("cosmos", "params.json", nil)
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

func GetDefaults() string {
	dat, err := util.GetBlockchainConfig("cosmos", "defaults.json", nil)
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

func GetServices() []util.Service {
	return nil
}
