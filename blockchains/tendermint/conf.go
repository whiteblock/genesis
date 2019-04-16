package tendermint

import (
	util "../../util"
	"io/ioutil"
)

func GetParams() string {
	dat, err := ioutil.ReadFile("./resources/tendermint/params.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

func GetDefaults() string {
	dat, err := ioutil.ReadFile("./resources/tendermint/defaults.json")
	if err != nil {
		panic(err) //Missing required files is a fatal error
	}
	return string(dat)
}

func GetServices() []util.Service {
	return nil
}
