package beam

import (
	db "../../db"
	ssh "../../ssh"
	testnet "../../testnet"
	util "../../util"
	helpers "../helpers"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

const port int = 10000

func Build(tn *testnet.TestNet) ([]string, error) {
	beamConf, err := NewConf(tn.LDD.Params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	tn.BuildState.SetBuildSteps(0 + (tn.LDD.Nodes * 4))

	tn.BuildState.SetBuildStage("Setting up the wallets")
	/**Set up wallets**/
	ownerKeys := make([]string, tn.LDD.Nodes)
	secretMinerKeys := make([]string, tn.LDD.Nodes)
	mux := sync.Mutex{}
	// walletIDs := []string{}
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, localNodeNum int, absoluteNodeNum int) error {

		client.DockerExec(localNodeNum, "beam-wallet --command init --pass password") //ign err

		res1, _ := client.DockerExec(localNodeNum, "beam-wallet --command export_owner_key --pass password") //ign err

		tn.BuildState.IncrementBuildProgress()

		re := regexp.MustCompile(`(?m)^Owner([A-z|0-9|\s|\:|\/|\+|\=])*$`)
		ownKLine := re.FindAllString(res1, -1)[0]

		mux.Lock()
		ownerKeys[absoluteNodeNum] = strings.Split(ownKLine, " ")[3]
		mux.Unlock()

		res2, _ := client.DockerExec(localNodeNum, "beam-wallet --command export_miner_key --subkey=1 --pass password") //ign err

		re = regexp.MustCompile(`(?m)^Secret([A-z|0-9|\s|\:|\/|\+|\=])*$`)
		secMLine := re.FindAllString(res2, -1)[0]

		mux.Lock()
		secretMinerKeys[absoluteNodeNum] = strings.Split(secMLine, " ")[3]
		mux.Unlock()

		tn.BuildState.IncrementBuildProgress()
		return nil
	})

	ips := []string{}

	for _, node := range tn.Nodes {

		ips = append(ips, node.Ip)
	}
	tn.BuildState.SetBuildStage("Creating node configuration files")
	/**Create node config files**/

	err = helpers.CreateConfigs(tn, "/beam/beam-node.cfg",
		func(_ int, _ int, absoluteNodeNum int) ([]byte, error) {
			ipsCpy := make([]string, len(ips))
			copy(ipsCpy, ips)
			beam_node_config, err := makeNodeConfig(beamConf, ownerKeys[absoluteNodeNum],
				secretMinerKeys[absoluteNodeNum], tn.LDD, absoluteNodeNum)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			for _, ip := range append(ipsCpy[:absoluteNodeNum], ipsCpy[absoluteNodeNum+1:]...) {
				beam_node_config += fmt.Sprintf("peer=%s:%d\n", ip, port)
			}
			return []byte(beam_node_config), nil
		})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = helpers.CreateConfigs(tn, "/beam/beam-wallet.cfg",
		func(_ int, _ int, absoluteNodeNum int) ([]byte, error) {
			beam_wallet_config := []string{
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
			return []byte(util.CombineConfig(beam_wallet_config)), nil
		})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tn.BuildState.SetBuildStage("Starting beam")
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, localNodeNum int, absoluteNodeNum int) error {
		defer tn.BuildState.IncrementBuildProgress()
		miningFlag := ""
		if absoluteNodeNum >= int(beamConf.Validators) {
			miningFlag = " --mining_threads 1"
		}
		_, err := client.DockerExecd(localNodeNum, fmt.Sprintf("beam-node%s", miningFlag))
		if err != nil {
			log.Println(err)
			return err
		}
		return client.DockerExecdLog(localNodeNum, fmt.Sprintf("beam-wallet --command listen -n 0.0.0.0:%d --pass password", port))
	})

	return nil, err
}

func Add(tn *testnet.TestNet) ([]string, error) {
	return nil, nil
}
