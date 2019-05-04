//Package syscoin handles syscoin specific functionality
package syscoin

import (
	"../../db"
	"../../ssh"
	"../../testnet"
	"../../util"
	"../helpers"
	"../registrar"
	"fmt"
	"log"
	"sync"
)

var conf *util.Config

const blockchain = "syscoin"

func init() {
	conf = util.GetConfig()
	registrar.RegisterBuild(blockchain, regTest)
	registrar.RegisterAddNodes(blockchain, add)
	registrar.RegisterServices(blockchain, GetServices)
	registrar.RegisterDefaults(blockchain, GetDefaults)
	registrar.RegisterParams(blockchain, GetParams)
}

// regTest sets up Syscoin Testnet in Regtest mode
func regTest(tn *testnet.TestNet) error {
	if tn.LDD.Nodes < 3 {
		log.Println("Tried to build syscoin with not enough nodes")
		return fmt.Errorf("not enough nodes")
	}

	sysconf, err := newConf(tn.LDD.Params)
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.SetBuildSteps(1 + (4 * tn.LDD.Nodes))

	tn.BuildState.SetBuildStage("Creating the syscoin conf files")
	err = handleConf(tn, sysconf)
	if err != nil {
		return util.LogError(err)
	}
	tn.BuildState.IncrementBuildProgress()
	fmt.Printf("done\n")

	tn.BuildState.SetBuildStage("Launching the nodes")

	return helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		return client.DockerExecdLog(node,
			"syscoind -conf=\"/syscoin/datadir/regtest.conf\" -datadir=\"/syscoin/datadir/\"")
	})
}

// Add handles adding a node to the artemis testnet
// TODO
func add(tn *testnet.TestNet) error {
	return nil
}

func handleConf(tn *testnet.TestNet, sysconf *sysConf) error {
	ips := []string{}
	for _, node := range tn.Nodes {
		ips = append(ips, node.IP)
	}

	noMasterNodes := int(float64(len(ips)) * (float64(sysconf.PercOfMNodes) / float64(100)))
	//log.Println(fmt.Sprintf("PERC = %d; NUM = %d;",sysconf.PercOfMNodes,noMasterNodes))

	if (len(ips) - noMasterNodes) == 0 {
		log.Println("Warning: No sender/receiver nodes availible. Removing 2 master nodes and setting them as sender/receiver")
		noMasterNodes -= 2
	} else if (len(ips)-noMasterNodes)%2 != 0 {
		log.Println("Warning: Removing a master node to keep senders and receivers equal")
		noMasterNodes--
		if noMasterNodes < 0 {
			log.Println("Warning: Attempt to remove a master node failed, adding one instead")
			noMasterNodes += 2
		}
	}

	connDistModel := make([]int, len(ips))
	for i := 0; i < len(ips); i++ {
		if i < noMasterNodes {
			connDistModel[i] = int(sysconf.MasterNodeConns)
		} else {
			connDistModel[i] = int(sysconf.NodeConns)
		}
	}

	connsDist, err := util.Distribute(ips, connDistModel)
	if err != nil {
		return util.LogError(err)
	}

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, node ssh.Node) error {
		defer tn.BuildState.IncrementBuildProgress()
		_, err := client.DockerExec(node, "mkdir -p /syscoin/datadir")
		return err
	})
	if err != nil {
		return util.LogError(err)
	}
	//Finally generate the configuration for each node
	mux := sync.Mutex{}

	masterNodes := []string{}
	senders := []string{}
	receivers := []string{}

	err = helpers.CreateConfigs(tn, "/syscoin/datadir/regtest.conf", func(node ssh.Node) ([]byte, error) {
		defer tn.BuildState.IncrementBuildProgress()
		confData := ""
		maxConns := 1
		if node.GetAbsoluteNumber() < noMasterNodes { //Master Node
			confData += sysconf.GenerateMN()
			mux.Lock()
			masterNodes = append(masterNodes, node.GetIP())
			mux.Unlock()
		} else if node.GetAbsoluteNumber()%2 == 0 { //Sender
			confData += sysconf.GenerateSender()
			mux.Lock()
			senders = append(senders, node.GetIP())
			mux.Unlock()
		} else { //Receiver
			confData += sysconf.GenerateReceiver()
			mux.Lock()
			receivers = append(receivers, node.GetIP())
			mux.Unlock()
		}

		confData += "rpcport=8369\nport=8370\n"
		for _, conn := range connsDist[node.GetAbsoluteNumber()] {
			confData += fmt.Sprintf("connect=%s:8370\n", conn)
			maxConns += 4
		}
		confData += "rpcallowip=0.0.0.0/0\n"
		//confData += fmt.Sprintf("maxconnections=%d\n",maxConns)
		return []byte(confData), nil
	})
	if err != nil {
		return util.LogError(err)
	}

	tn.BuildState.SetExt("masterNodes", masterNodes)
	tn.BuildState.SetExt("senders", senders)
	tn.BuildState.SetExt("receivers", receivers)

	return nil
}
