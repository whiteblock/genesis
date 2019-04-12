package rchain

import (
	db "../../db"
	state "../../state"
	util "../../util"
	helpers "../helpers"
	"fmt"
	"github.com/Whiteblock/mustache"
	"log"
	"regexp"
	"time"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

func Build(details db.DeploymentDetails, servers []db.Server, clients []*util.SshClient, buildState *state.BuildState) ([]string, error) {

	rchainConf, err := NewRChainConf(details.Params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	buildState.SetBuildSteps(9 + (len(servers) * 2) + (details.Nodes * 3))
	buildState.SetBuildStage("Setting up data collection")

	services, err := util.GetServiceIps(GetServices())
	buildState.IncrementBuildProgress()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	/**Make the data directories**/

	err = helpers.AllNodeExecCon(servers, buildState, func(serverNum int, localNodeNum int, absoluteNodeNum int) error {
		buildState.IncrementBuildProgress()
		_, err := clients[serverNum].DockerExec(localNodeNum, "mkdir /datadir")
		return err
	})

	/**Setup the first node**/
	err = createFirstConfigFile(details, clients[0], 0, rchainConf, services["wb_influx_proxy"], buildState)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	buildState.IncrementBuildProgress()
	km, err := helpers.NewKeyMaster(&details, "rchain")
	keyPairs := make([]util.KeyPair, details.Nodes)

	for i, _ := range keyPairs {
		keyPairs[i], err = km.GetKeyPair(clients[0])
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	buildState.IncrementBuildProgress()

	buildState.SetBuildStage("Setting up bonds")
	/**Setup bonds**/
	{
		bonds := make([]string, details.Nodes)
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
		defer clients[0].Run("rm -f /home/appo/bonds.txt")

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
				rchainConf.Command, servers[0].Ips[0]))
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
		err = createConfigFile(details, enode, rchainConf, services["wb_influx_proxy"], buildState)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	buildState.SetBuildStage("Configuring the other rchain nodes")
	/**Copy config files to the rest of the nodes**/
	for i, server := range servers {
		err = clients[i].Scp("rnode.conf", "/home/appo/rnode.conf")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		buildState.IncrementBuildProgress()
		for node, _ := range server.Ips {
			if node == 0 && i == 0 {
				continue
			}
			err = clients[i].DockerCp(node, "/home/appo/rnode.conf", "/datadir/rnode.conf")
			if err != nil {
				log.Println(err)
				return nil, err
			}
			buildState.IncrementBuildProgress()

		}
		_, err = clients[i].Run("rm -f ~/rnode.conf")
		if err != nil {
			log.Println(err)
			return nil, err
		}
		buildState.IncrementBuildProgress()
	}

	if err != nil {
		log.Println(err)
		return nil, err
	}
	buildState.SetBuildStage("Starting the rest of the nodes")
	/**Start up the rest of the nodes**/
	node := 0
	var validators int64 = 0
	for i, server := range servers {
		for j, ip := range server.Ips {
			if node == 0 {
				node++
				continue
			}
			if validators < rchainConf.Validators {
				err = clients[i].DockerExecdLog(j,
					fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --validator-private-key %s --host %s",
						rchainConf.Command, enode, keyPairs[node].PrivateKey, ip))
				validators++
			} else {
				err = clients[i].DockerExecdLog(j,
					fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --host %s",
						rchainConf.Command, enode, ip))
			}
			buildState.IncrementBuildProgress()
			if err != nil {
				log.Println(err)
				return nil, err
			}
			node++
		}
	}
	/*err = SetupPrometheus(servers,clients)
	  if err != nil{
	      log.Println(err)
	      return nil,err
	  }*/
	return nil, nil
}

func createFirstConfigFile(details db.DeploymentDetails, client *util.SshClient, node int, rchainConf *RChainConf, influxIP string, buildState *state.BuildState) error {

	raw := map[string]interface{}{
		"influxIp":       influxIP,
		"validatorCount": rchainConf.ValidatorCount,
		"standalone":     true,
	}
	raw = util.MergeStringMaps(raw, details.Params) //Allow arbitrary custom options for rchain

	filler := util.ConvertToStringMap(raw)
	dat, err := util.GetBlockchainConfig("rchain", "rchain.conf.mustache", details.Files)
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
	if err != nil {
		log.Println(err)
		return err
	}
	err = client.DockerCp(node, "/home/appo/rnode.conf", "/datadir/rnode.conf")
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = client.Run("rm -f ~/rnode.conf")
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func Add(details db.DeploymentDetails, servers []db.Server, clients []*util.SshClient,
	newNodes map[int][]string, buildState *state.BuildState) ([]string, error) {
	return nil, nil
}

func createConfigFile(details db.DeploymentDetails, bootnodeAddr string, rchainConf *RChainConf,
	influxIP string, buildState *state.BuildState) error {

	raw := map[string]interface{}{
		"influxIp":       influxIP,
		"validatorCount": rchainConf.ValidatorCount,
		"standalone":     false,
		"bootstrap":      fmt.Sprintf("bootstrap = \"%s\"", bootnodeAddr),
	}
	raw = util.MergeStringMaps(raw, details.Params) //Allow arbitrary custom options for rchain

	filler := util.ConvertToStringMap(raw)
	dat, err := util.GetBlockchainConfig("rchain", "rchain.conf.mustache", details.Files)
	if err != nil {
		log.Println(err)
		return err
	}
	data, err := mustache.Render(string(dat), filler)
	if err != nil {
		log.Println(err)
		return err
	}
	return buildState.Write("rnode.conf", data)
}
