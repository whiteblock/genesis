package rchain

import (
	"../../db"
	"../../ssh"
	"../../state"
	"../../testnet"
	"../../util"
	"../helpers"
	"fmt"
	"github.com/Whiteblock/mustache"
	"log"
	"regexp"
	"sync"
	"time"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

func Build(tn *testnet.TestNet) ([]string, error) {
	buildState := tn.BuildState
	clients := tn.GetFlatClients()
	rchainConf, err := NewRChainConf(tn.LDD.Params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	buildState.SetBuildSteps(9 + (len(tn.Servers) * 2) + (tn.LDD.Nodes * 2))
	buildState.SetBuildStage("Setting up data collection")

	services, err := util.GetServiceIps(GetServices())
	buildState.IncrementBuildProgress()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	/**Make the data directories**/

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, localNodeNum int, _ int) error {
		buildState.IncrementBuildProgress()
		_, err := client.DockerExec(localNodeNum, "mkdir /datadir")
		return err
	})
	/**Setup the first node**/
	err = createFirstConfigFile(tn.LDD, clients[0], 0, rchainConf, services["wb_influx_proxy"], tn.BuildState)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	buildState.IncrementBuildProgress()
	km, err := helpers.NewKeyMaster(tn.LDD, "rchain")
	keyPairs := make([]util.KeyPair, tn.LDD.Nodes)

	for i := range keyPairs {
		keyPairs[i], err = km.GetKeyPair(clients[0])
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	buildState.Set("keyPairs", keyPairs)

	buildState.IncrementBuildProgress()

	buildState.SetBuildStage("Setting up bonds")
	/**Setup bonds**/
	{
		bonds := make([]string, tn.LDD.Nodes)
		for i, keyPair := range keyPairs {
			bonds[i] = fmt.Sprintf("%s 1000000", keyPair.PublicKey)
		}
		buildState.IncrementBuildProgress()
		err = buildState.Write("bonds.txt", util.CombineConfig(bonds))
		if err != nil {
			log.Println(err)
			return nil, err
		}
		buildState.IncrementBuildProgress()

		err = clients[0].Scp("bonds.txt", "/home/appo/bonds.txt")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		buildState.IncrementBuildProgress()
		buildState.Defer(func() { clients[0].Run("rm -f /home/appo/bonds.txt") })

		err = clients[0].DockerCp(0, "/home/appo/bonds.txt", "/bonds.txt")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		buildState.IncrementBuildProgress()

	}

	buildState.SetBuildStage("Starting the boot node")
	var enode string
	{
		err = clients[0].DockerExecdLog(0,
			fmt.Sprintf("%s run --standalone --data-dir \"/datadir\" --host %s --bonds-file /bonds.txt --has-faucet",
				rchainConf.Command, tn.Nodes[0].Ip))
		buildState.IncrementBuildProgress()
		if err != nil {
			log.Println(err)
			return nil, err
		}
		//fmt.Println("Attempting to get the enode address")
		buildState.SetBuildStage("Waiting for the boot node's address")
		for i := 0; i < 1000; i++ {
			fmt.Println("Checking if the boot node is ready...")
			time.Sleep(time.Duration(1 * time.Second))
			output, err := clients[0].DockerExec(0, fmt.Sprintf("cat %s", conf.DockerOutputFile))
			if err != nil {
				log.Println(err)
				return nil, err
			}
			re := regexp.MustCompile(`(?m)rnode:\/\/[a-z|0-9]*\@([0-9]{1,3}\.){3}[0-9]{1,3}\?protocol=[0-9]*\&discovery=[0-9]*`)

			if !re.MatchString(output) {
				fmt.Println("Not ready")
				continue
			}
			enode = re.FindAllString(output, 1)[0]
			fmt.Println("Ready")
			break
		}
		buildState.IncrementBuildProgress()
		/*
		   influxIp
		   validators
		*/
		log.Println("Got the address for the bootnode: " + enode)
	}
	buildState.Set("bootnode", enode)
	buildState.Set("rchainConf", *rchainConf)

	err = helpers.CreateConfigs(tn, "/datadir/rnode.conf",
		func(serverNum int, localNodeNum int, absoluteNodeNum int) ([]byte, error) {
			if absoluteNodeNum == 0 {
				return nil, nil
			}
			return createConfigFile(tn.LDD, enode, rchainConf, services["wb_influx_proxy"], buildState, absoluteNodeNum)
		})

	buildState.SetBuildStage("Configuring the other rchain nodes")
	/**Copy config files to the rest of the nodes**/
	buildState.IncrementBuildProgress()

	buildState.SetBuildStage("Starting the rest of the nodes")
	/**Start up the rest of the nodes**/
	mux := sync.Mutex{}

	var validators int64 = 0

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, localNodeNum int, absoluteNodeNum int) error {
		defer buildState.IncrementBuildProgress()
		if absoluteNodeNum == 0 {
			return nil
		}
		ip := tn.Nodes[absoluteNodeNum].Ip
		mux.Lock()
		isValidator := validators < rchainConf.Validators
		validators++
		mux.Unlock()
		if isValidator {
			err := client.DockerExecdLog(localNodeNum,
				fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --validator-private-key %s --host %s",
					rchainConf.Command, enode, keyPairs[absoluteNodeNum].PrivateKey, ip))
			return err
		}
		return client.DockerExecdLog(localNodeNum,
			fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --host %s",
				rchainConf.Command, enode, ip))
	})
	return nil, err
}

func createFirstConfigFile(details *db.DeploymentDetails, client *ssh.Client, node int, rchainConf *RChainConf, influxIP string, buildState *state.BuildState) error {

	raw := map[string]interface{}{
		"influxIp":       influxIP,
		"validatorCount": rchainConf.ValidatorCount,
		"standalone":     true,
	}
	raw = util.MergeStringMaps(raw, details.Params) //Allow arbitrary custom options for rchain

	filler := util.ConvertToStringMap(raw)
	dat, err := helpers.GetBlockchainConfig("rchain", 0, "rchain.conf.mustache", details)
	if err != nil {
		log.Println(err)
		return err
	}
	data, err := mustache.Render(string(dat), filler)
	if err != nil {
		log.Println(err)
		return nil
	}
	err = buildState.Write("rnode.conf", data)
	if err != nil {
		log.Println(err)
		return err
	}
	err = client.Scp("rnode.conf", "/home/appo/rnode.conf")
	buildState.Defer(func() { client.Run("rm -f ~/rnode.conf") })
	if err != nil {
		log.Println(err)
		return err
	}
	return client.DockerCp(node, "/home/appo/rnode.conf", "/datadir/rnode.conf")
}

/**********************************************************************ADD********************************************************************/
func Add(tn *testnet.TestNet) ([]string, error) {

	rchainConf, err := NewRChainConf(tn.CombinedDetails.Params)
	tn.BuildState.SetBuildSteps(1 + 2*len(tn.NewlyBuiltNodes)) //TODO
	if err != nil {
		log.Println(err)
		return nil, err
	}
	iEnode, ok := tn.BuildState.Get("bootnode")
	if !ok {
		err = fmt.Errorf("rebuild: missing bootnode")
		log.Println(err)
		return nil, err
	}
	enode := iEnode.(string)

	services, err := util.GetServiceIps(GetServices())
	if err != nil {
		log.Println(err)
		return nil, err
	}
	keyPairs := []util.KeyPair{}
	km, err := helpers.NewKeyMaster(&tn.CombinedDetails, "rchain")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	for range tn.Nodes {
		kp, err := km.GetKeyPair(tn.GetFlatClients()[0])
		if err != nil {
			log.Println(err)
			return nil, err
		}
		keyPairs = append(keyPairs, kp)
	}

	err = helpers.AllNewNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, localNodeNum int, _ int) error {
		defer tn.BuildState.IncrementBuildProgress()
		_, err := client.DockerExec(localNodeNum, "mkdir /datadir")
		return err
	})

	err = helpers.CreateConfigsNewNodes(tn, "/datadir/rnode.conf", func(_ int, _ int, absoluteNodeNum int) ([]byte, error) {
		return createConfigFile(&tn.CombinedDetails, enode, rchainConf, services["wb_influx_proxy"], tn.BuildState, absoluteNodeNum)
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if !ok {
		err = fmt.Errorf("rebuild: missing key pairs")
		log.Println(err)
		return nil, err
	}
	tn.BuildState.SetBuildStage("Configuring the other rchain nodes")
	/**Copy config files to the rest of the nodes**/
	tn.BuildState.IncrementBuildProgress()

	tn.BuildState.SetBuildStage("Starting the rest of the nodes")
	/**Start up the rest of the nodes**/
	var validators int64 = 0
	mux := sync.Mutex{}
	err = helpers.AllNewNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, localNodeNum int, absoluteNodeNum int) error {
		defer tn.BuildState.IncrementBuildProgress()
		ip := tn.Nodes[absoluteNodeNum].Ip

		mux.Lock()
		isValidator := validators < rchainConf.Validators
		validators++
		mux.Unlock()

		if isValidator {
			err = client.DockerExecdLog(localNodeNum,
				fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --validator-private-key %s --host %s",
					rchainConf.Command, enode, keyPairs[absoluteNodeNum].PrivateKey, ip))
			return err
		}
		return client.DockerExecdLog(localNodeNum,
			fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --host %s",
				rchainConf.Command, enode, ip))
	})
	return nil, err
}

func createConfigFile(details *db.DeploymentDetails, bootnodeAddr string, rchainConf *RChainConf,
	influxIP string, buildState *state.BuildState, node int) ([]byte, error) {

	raw := map[string]interface{}{
		"influxIp":       influxIP,
		"validatorCount": rchainConf.ValidatorCount,
		"standalone":     false,
	}

	raw = util.MergeStringMaps(raw, details.Params) //Allow arbitrary custom options for rchain

	filler := util.ConvertToStringMap(raw)
	filler["bootstrap"] = fmt.Sprintf("bootstrap = \"%s\"", bootnodeAddr)
	dat, err := helpers.GetBlockchainConfig("rchain", node, "rchain.conf.mustache", details)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	data, err := mustache.Render(string(dat), filler)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return []byte(data), nil
}
