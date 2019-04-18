package cosmos

import (
	util "../../util"
	helpers "../helpers"
)

func GetParams() string {
	dat, err := helpers.GetStaticBlockchainConfig("cosmos", "params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

func GetDefaults() string {
	dat, err := helpers.GetStaticBlockchainConfig("cosmos", "defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

func GetServices() []util.Service {
	return nil
}
