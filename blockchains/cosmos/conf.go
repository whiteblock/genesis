package cosmos

import (
	"../../util"
	"../helpers"
)

// GetParams fetchs cosmos related parameters
func GetParams() string {
	dat, err := helpers.GetStaticBlockchainConfig("cosmos", "params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetDefaults fetchs cosmos related parameter defaults
func GetDefaults() string {
	dat, err := helpers.GetStaticBlockchainConfig("cosmos", "defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetServices returns the services which are used by cosmos
func GetServices() []util.Service {
	return nil
}
