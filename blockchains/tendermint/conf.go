package tendermint

import (
	"../../util"
	"../helpers"
)

// GetParams fetchs tendermint related parameters
func GetParams() string {
	dat, err := helpers.GetStaticBlockchainConfig(blockchain, "params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetDefaults fetchs tendermint related parameter defaults
func GetDefaults() string {
	dat, err := helpers.GetStaticBlockchainConfig(blockchain, "defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetServices returns the services which are used by tendermint
func GetServices() []util.Service {
	return nil
}
