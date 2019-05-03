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
	registrar.RegisterAdditionalLogs(blockchain, map[string]string{
		"json": "/artemis/data/log.json"})
}

// Build builds out a fresh new artemis test network
func Build(tn *testnet.TestNet) ([]string, error) {
	aconf, err := newConf(tn.LDD.Params)
	if err != nil {
		return nil, util.LogError(err)
	}
	fetchedConfChan := make(chan string)

	go func(aconf artemisConf) {
		res, err := util.HTTPRequest("GET", aconf["constantsSource"].(string), "")
		if err != nil {
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
		func(node ssh.Node) ([]byte, error) {
			defer tn.BuildState.IncrementBuildProgress()
			identity := fmt.Sprintf("0x%.8x", node.GetAbsoluteNumber())
			artemisNodeConfig, err := makeNodeConfig(aconf, identity, peers, node.GetAbsoluteNumber(), tn.LDD, rawConstants)
			return []byte(artemisNodeConfig), err
		})

	tn.BuildState.SetBuildStage("Starting Artemis")
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()

		artemisCmd := `artemis -c /artemis/config/config.toml -o /artemis/data/log.json 2>&1 | tee /output.log`

		_, err := client.DockerExecd(node, "tmux new -s whiteblock -d")
		if err != nil {
			return util.LogError(err)
		}

		_, err = client.DockerExecd(node, fmt.Sprintf("tmux send-keys -t whiteblock '%s' C-m", artemisCmd))
		return err
	})
	if err != nil {
		return nil, util.LogError(err)
	}

	return nil, nil
}

// Add handles adding a node to the artemis testnet
// TODO
func Add(tn *testnet.TestNet) ([]string, error) {
	return nil, nil
}
