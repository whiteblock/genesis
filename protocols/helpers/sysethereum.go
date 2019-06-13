package helpers

import (
	"encoding/json"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/util"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/mustache"
)

// SysethereumService represents the SysethereumService service
type SysethereumService struct {
	SimpleService
}

type sysethereumConf map[string]interface{}


func newConf(data map[string]interface{}) (sysethereumConf, error) {
	rawDefaults := DefaultGetDefaultsFn("sysethereum")()
	defaults := map[string]interface{}{}

	err := json.Unmarshal([]byte(rawDefaults), &defaults)
	if err != nil {
		return nil, util.LogError(err)
	}
	finalData := util.MergeStringMaps(defaults, data)
	out := new(sysethereumConf)
	*out = sysethereumConf(finalData)

	return *out, nil
}

// Prepare prepares the sysethereum service
func (p SysethereumService) Prepare(client ssh.Client, tn *testnet.TestNet) error {
	aconf, err := newConf(tn.LDD.Params)
	if err != nil {
		return util.LogError(err)
	}

	err = CreateConfigs(tn, "/sysethereum.conf", func(node ssh.Node) ([]byte, error) {
		defer tn.BuildState.IncrementBuildProgress()
		conf, err := makeConfig(aconf, &tn.CombinedDetails)
		return []byte(conf), err
	})
	if err != nil {
		return util.LogError(err)
	}
	return nil
}

func makeConfig(aconf sysethereumConf, details *db.DeploymentDetails) (string, error) {

	sysEthConf, err := util.CopyMap(aconf)
	filler := util.ConvertToStringMap(sysEthConf)
	filler["contractsDirectory"] = "/contracts"
	filler["dataDirectory"] = "/data"
	if err != nil {
		return "", util.LogError(err)
	}
	dat, err := GetBlockchainConfig("sysethereum", 0, "sysethereum.conf.mustache", details)
	if err != nil {
		return "", util.LogError(err)
	}
	return mustache.Render(string(dat), filler)
}

func (p SysethereumService) GetCommand() string {
	return "-Dsysethereum.agents.conf.file=/sysethereum.conf"
}

// RegisterSysethereum exposes a Sysethereum service on the testnet.
func RegisterSysethereum() Service {
	return SysethereumService{
		SimpleService{
			Name:    "ganache",
			Image:   "gcr.io/whiteblock/sysethereum-agents",
			Env:     map[string]string{},
			Ports:   []string{},
			Volumes: []string{},
		},
	}
}
