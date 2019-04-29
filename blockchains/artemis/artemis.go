//Package artemis handles artemis specific functionality
package artemis

import (
	"../../db"
	"../../ssh"
	"../../testnet"
	"../../util"
	"../helpers"
	"../registrar"
	"fmt"
	"log"
	"strings"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
	blockchain := "artemis"
	registrar.RegisterBuild(blockchain, Build)
	registrar.RegisterAddNodes(blockchain, Add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, GetDefaults)
	registrar.RegisterParams(blockchain, GetParams)
}

// Build builds out a fresh new artemis test network
func Build(tn *testnet.TestNet) ([]string, error) {
	aconf, err := newConf(tn.LDD.Params)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	fetchedConfChan := make(chan string)

	go func(aconf artemisConf) {
		res, err := util.HTTPRequest("GET", aconf["constantsSource"].(string), "")
		if err != nil {
			log.Println(err)
			tn.BuildState.ReportError(err)
			return
		}
		fetchedConfChan <- string(res)

	}(aconf)

	tn.BuildState.SetBuildSteps(0 + (tn.LDD.Nodes * 4))

	port := 9000
	peers := "["
	var peer string
	for i, node := range tn.Nodes {
		peer = fmt.Sprintf("%s://whiteblock-node%d@%s:%d",
			aconf["networkMode"],
			node.LocalID,
			node.IP,
			port,
		)
		if i != len(tn.Nodes)-1 {
			peers = peers + "\"" + peer + "\"" + ","
		} else {
			peers = peers + "\"" + peer + "\""
		}
		tn.BuildState.IncrementBuildProgress()
	}

	peers = peers + "]"
	fmt.Println(peers)

	tn.BuildState.SetBuildStage("Creating node configuration files")
	/**Create node config files**/
	fetchedConf := <-fetchedConfChan

	constantsIndex := strings.Index(fetchedConf, "[constants]")
	if constantsIndex == -1 {
		return nil, fmt.Errorf("couldn't find \"[constants]\" in file fetched from given source")
	}
	rawConstants := fetchedConf[constantsIndex:]
	err = helpers.CreateConfigs(tn, "/artemis/config/config.toml",
		func(serverId int, localNodeNum int, absoluteNodeNum int) ([]byte, error) {
			defer tn.BuildState.IncrementBuildProgress()
			identity := fmt.Sprintf("0x%.8x", absoluteNodeNum)
			artemisNodeConfig, err := makeNodeConfig(aconf, identity, peers, absoluteNodeNum, tn.LDD, rawConstants)
			return []byte(artemisNodeConfig), err
		})

	tn.BuildState.SetBuildStage("Starting Artemis")
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, localNodeNum int, absoluteNodeNum int) error {
		defer tn.BuildState.IncrementBuildProgress()

		artemisCmd := `artemis -c /artemis/config/config.toml -o /artemis/data/data.json 2>&1 | tee /output.log`

		_, err := client.DockerExecd(localNodeNum, "tmux new -s whiteblock -d")
		if err != nil {
			log.Println(err)
			return err
		}

		_, err = client.DockerExecd(localNodeNum, fmt.Sprintf("tmux send-keys -t whiteblock '%s' C-m", artemisCmd))
		return err
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return nil, nil
}

// Add handles adding a node to the artemis testnet
// TODO
func Add(tn *testnet.TestNet) ([]string, error) {
	return nil, nil
}
