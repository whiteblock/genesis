//Package beam handles beam specific functionality
package beam

import (
	"../../db"
	"../../ssh"
	"../../testnet"
	"../../util"
	"../helpers"
	"../registrar"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
	blockchain := "beam"
	registrar.RegisterBuild(blockchain, Build)
	registrar.RegisterAddNodes(blockchain, Add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, GetDefaults)
	registrar.RegisterParams(blockchain, GetParams)
}

const port int = 10000

// Build builds out a fresh new beam test network
func Build(tn *testnet.TestNet) ([]string, error) {
	bConf, err := newConf(tn.LDD.Params)
	if err != nil {
		return nil, util.LogError(err)
	}
	tn.BuildState.SetBuildSteps(0 + (tn.LDD.Nodes * 4))

	tn.BuildState.SetBuildStage("Setting up the wallets")
	/**Set up wallets**/
	ownerKeys := make([]string, tn.LDD.Nodes)
	secretMinerKeys := make([]string, tn.LDD.Nodes)
	mux := sync.Mutex{}
	// walletIDs := []string{}
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {

		client.DockerExec(node, "beam-wallet --command init --pass password") //ign err

		res1, _ := client.DockerExec(node, "beam-wallet --command export_owner_key --pass password") //ign err

		tn.BuildState.IncrementBuildProgress()

		re := regexp.MustCompile(`(?m)^Owner([A-z|0-9|\s|\:|\/|\+|\=])*$`)
		ownKLine := re.FindAllString(res1, -1)[0]

		mux.Lock()
		ownerKeys[node.GetAbsoluteNumber()] = strings.Split(ownKLine, " ")[3]
		mux.Unlock()

		res2, _ := client.DockerExec(node, "beam-wallet --command export_miner_key --subkey=1 --pass password") //ign err

		re = regexp.MustCompile(`(?m)^Secret([A-z|0-9|\s|\:|\/|\+|\=])*$`)
		secMLine := re.FindAllString(res2, -1)[0]

		mux.Lock()
		secretMinerKeys[node.GetAbsoluteNumber()] = strings.Split(secMLine, " ")[3]
		mux.Unlock()

		tn.BuildState.IncrementBuildProgress()
		return nil
	})

	ips := []string{}

	for _, node := range tn.Nodes {

		ips = append(ips, node.IP)
	}
	tn.BuildState.SetBuildStage("Creating node configuration files")
	/**Create node config files**/

	err = helpers.CreateConfigs(tn, "/beam/beam-node.cfg",
		func(node ssh.Node) ([]byte, error) {
			ipsCpy := make([]string, len(ips))
			copy(ipsCpy, ips)
			beamNodeConfig, err := makeNodeConfig(bConf, ownerKeys[node.GetAbsoluteNumber()],
				secretMinerKeys[node.GetAbsoluteNumber()], tn.LDD, node.GetAbsoluteNumber())
			if err != nil {
				return nil, util.LogError(err)
			}
			for _, ip := range append(ipsCpy[:node.GetAbsoluteNumber()], ipsCpy[node.GetAbsoluteNumber()+1:]...) {
				beamNodeConfig += fmt.Sprintf("peer=%s:%d\n", ip, port)
			}
			return []byte(beamNodeConfig), nil
		})
	if err != nil {
		return nil, util.LogError(err)
	}
	err = helpers.CreateConfigs(tn, "/beam/beam-wallet.cfg",
		func(_ ssh.Node) ([]byte, error) {
			beamWalletConfig := []string{
				"# Emission.Value0=800000000",
				"# Emission.Drop0=525600",
				"# Emission.Drop1=2102400",
				"Maturity.Coinbase=1",
				"# Maturity.Std=0",
				"# MaxBodySize=0x100000",
				"DA.Target_s=1",
				"# DA.MaxAhead_s=900",
				"# DA.WindowWork=120",
				"# DA.WindowMedian0=25",
				"# DA.WindowMedian1=7",
				"DA.Difficulty0=100",
				"# AllowPublicUtxos=0",
				"# FakePoW=0",
			}
			return []byte(util.CombineConfig(beamWalletConfig)), nil
		})
	if err != nil {
		return nil, util.LogError(err)
	}

	tn.BuildState.SetBuildStage("Starting beam")
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		miningFlag := ""
		if node.GetAbsoluteNumber() >= int(bConf.Validators) {
			miningFlag = " --mining_threads 1"
		}
		_, err := client.DockerExecd(node, fmt.Sprintf("beam-node%s", miningFlag))
		if err != nil {
			return util.LogError(err)
		}
		return client.DockerExecdLog(node, fmt.Sprintf("beam-wallet --command listen -n 0.0.0.0:%d --pass password", port))
	})

	return nil, err
}

// Add handles adding nodes to the testnet
func Add(tn *testnet.TestNet) ([]string, error) {
	return nil, nil
}
