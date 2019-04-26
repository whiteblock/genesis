package tendermint

import (
	"../../util"
	"io/ioutil"
)

// GetParams fetchs tendermint related parameters
func GetParams() string {
	dat, err := ioutil.ReadFile("./resources/tendermint/params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetDefaults fetchs tendermint related parameter defaults
func GetDefaults() string {
	dat, err := ioutil.ReadFile("./resources/tendermint/defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

// GetServices returns the services which are used by tendermint
func GetServices() []util.Service {
	return nil
}
