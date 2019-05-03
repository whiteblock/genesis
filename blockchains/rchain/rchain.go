//Package rchain handles rchain specific functionality
package rchain

import (
	"../../db"
	"../../ssh"
	"../../state"
	"../../testnet"
	"../../util"
	"../helpers"
	"../registrar"
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

	blockchain := "rchain"
	registrar.RegisterBuild(blockchain, Build)
	registrar.RegisterAddNodes(blockchain, Add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, GetDefaults)
	registrar.RegisterParams(blockchain, GetParams)
}

// Build builds out a fresh new rchain test network
func Build(tn *testnet.TestNet) ([]string, error) {
	buildState := tn.BuildState
	clients := tn.GetFlatClients()
	rConf, err := newRChainConf(tn.LDD.Params)
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
	err = createFirstConfigFile(tn.LDD, clients[0], 0, rConf, services["wb_influx_proxy"], tn.BuildState)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	/**Check to make sure the rnode command is valid**/
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, localNodeNum int, _ int) error {
		_, err := client.DockerExec(localNodeNum, fmt.Sprintf("bash -c '%s --help'", rConf.Command))
		if err != nil {
			fmt.Println(err)
			return fmt.Errorf("could not find command \"%s\"", rConf.Command)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	buildState.IncrementBuildProgress()
	km, err := helpers.NewKeyMaster(tn.LDD, "rchain")
	keyPairs := make([]util.KeyPair, tn.LDD.Nodes)
	validatorKeyPairs := make([]util.KeyPair, rConf.Validators)
	for i := range keyPairs {
		keyPairs[i], err = km.GetKeyPair(clients[0])
		if i > 0 && i-1 < len(validatorKeyPairs) {
			validatorKeyPairs[i-1] = keyPairs[i]
		}
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	for i := len(keyPairs) - 1; i < len(validatorKeyPairs); i++ {
		validatorKeyPairs[i], err = km.GetKeyPair(clients[0])
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	//fmt.Printf("Keypairs = %#v\n", keyPairs)
	//fmt.Printf("BalidatorKeyPairs = %#v\n", validatorKeyPairs)
	buildState.Set("keyPairs", keyPairs)
	buildState.Set("validatorKeyPairs", validatorKeyPairs)

	buildState.IncrementBuildProgress()

	buildState.SetBuildStage("Setting up bonds")
	/**Setup bonds**/
	{
		bonds := make([]string, len(validatorKeyPairs))
		for i, keyPair := range validatorKeyPairs {
			bonds[i] = fmt.Sprintf("%s %d", keyPair.PublicKey, rConf.BondsValue)
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
				rConf.Command, tn.Nodes[0].IP))
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
	buildState.Set("rConf", *rConf)

	err = helpers.CreateConfigs(tn, "/datadir/rnode.conf",
		func(node ssh.Node) ([]byte, error) {
			if node.GetAbsoluteNumber() == 0 {
				return nil, nil
			}
			return createConfigFile(enode, rConf, services["wb_influx_proxy"], node.GetAbsoluteNumber())
		})

	buildState.SetBuildStage("Configuring the other rchain nodes")
	/**Copy config files to the rest of the nodes**/
	buildState.IncrementBuildProgress()

	buildState.SetBuildStage("Starting the rest of the nodes")
	/**Start up the rest of the nodes**/
	mux := sync.Mutex{}

	var validators int64

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, node ssh.Node) error {
		defer buildState.IncrementBuildProgress()
		if node.GetAbsoluteNumber() == 0 {
			return nil
		}
		ip := tn.Nodes[node.GetAbsoluteNumber()].IP
		mux.Lock()
		isValidator := validators < rConf.Validators
		validators++
		mux.Unlock()
		if isValidator {
			err := client.DockerExecdLog(localNodeNum,
				fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --validator-private-key %s --host %s",
					rConf.Command, enode, keyPairs[node.GetAbsoluteNumber()].PrivateKey, node.GetIP()))
			return err
		}
		return client.DockerExecdLog(localNodeNum,
			fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --host %s",
				rConf.Command, enode, node.GetIP()))
	})
	return nil, err
}

func createFirstConfigFile(tn *testnet.TestNet, client *ssh.Client, node ssh.Node, rConf *rChainConf, influxIP string) error {

	raw := map[string]interface{}{
		"influxIp":       influxIP,
		"validatorCount": rConf.ValidatorCount,
		"standalone":     true,
	}
	raw = util.MergeStringMaps(raw, tn.LDD.Params) //Allow arbitrary custom options for rchain

	filler := util.ConvertToStringMap(raw)
	dat, err := helpers.GetBlockchainConfig("rchain", 0, "rchain.conf.mustache", tn.LDD)
	if err != nil {
		log.Println(err)
		return err
	}
	data, err := mustache.Render(string(dat), filler)
	if err != nil {
		log.Println(err)
		return nil
	}
	err = tn.BuildState.Write("rnode.conf", data)
	if err != nil {
		log.Println(err)
		return err
	}
	err = client.Scp("rnode.conf", "/home/appo/rnode.conf")
	tn.BuildState.Defer(func() { client.Run("rm -f ~/rnode.conf") })
	if err != nil {
		log.Println(err)
		return err
	}
	return client.DockerCp(node, "/home/appo/rnode.conf", "/datadir/rnode.conf")
}

/**********************************************************************ADD********************************************************************/

// Add handles the addition of nodes to the rchain testnet
func Add(tn *testnet.TestNet) ([]string, error) {

	rConf, err := newRChainConf(tn.CombinedDetails.Params)
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

	err = helpers.AllNewNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		_, err := client.DockerExec(node, "mkdir /datadir")
		return err
	})

	err = helpers.CreateConfigsNewNodes(tn, "/datadir/rnode.conf", func(node ssh.Node) ([]byte, error) {
		return createConfigFile(&tn.CombinedDetails, enode, rConf, services["wb_influx_proxy"], tn.BuildState, node.GetAbsoluteNumber())
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
	var validators int64
	mux := sync.Mutex{}
	err = helpers.AllNewNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()

		mux.Lock()
		isValidator := validators < rConf.Validators
		validators++
		mux.Unlock()

		if isValidator {
			err = client.DockerExecdLog(localNodeNum,
				fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --validator-private-key %s --host %s",
					rConf.Command, enode, keyPairs[node.GetAbsoluteNumber()].PrivateKey, node.GetIP()))
			return err
		}
		return client.DockerExecdLog(localNodeNum,
			fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --host %s",
				rConf.Command, enode, node.GetIP()))
	})
	return nil, err
}

func createConfigFile(tn *testnet.TestNet, bootnodeAddr string, rConf *rChainConf,
	influxIP string, buildState *state.BuildState, node int) ([]byte, error) {

	raw := map[string]interface{}{
		"influxIp":       influxIP,
		"validatorCount": rConf.ValidatorCount,
		"standalone":     false,
	}

	raw = util.MergeStringMaps(raw, tn.LDD.Params) //Allow arbitrary custom options for rchain

	filler := util.ConvertToStringMap(raw)
	filler["bootstrap"] = fmt.Sprintf("bootstrap = \"%s\"", bootnodeAddr)
	dat, err := helpers.GetBlockchainConfig("rchain", node, "rchain.conf.mustache", tn.LDD)
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
