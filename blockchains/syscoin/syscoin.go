package syscoin

import (
	db "../../db"
	ssh "../../ssh"
	testnet "../../testnet"
	util "../../util"
	helpers "../helpers"
	"errors"
	"fmt"
	"log"
	"sync"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

/**
 * Sets up Syscoin Testnet in Regtest mode
 * @param {[type]} data    map[string]interface{}   The configuration optiosn given by the client
 * @param {[type]} nodes   int                      The number of nodes to build
 * @param {[type]} servers []db.Server)             The servers to be built on
 * @return ([]string,error [description]
 */
func RegTest(tn *testnet.TestNet) ([]string, error) {
	if tn.LDD.Nodes < 3 {
		log.Println("Tried to build syscoin with not enough nodes")
		return nil, errors.New("Tried to build syscoin with not enough nodes")
	}

	sysconf, err := NewConf(tn.LDD.Params)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tn.BuildState.SetBuildSteps(1 + (4 * tn.LDD.Nodes))

	tn.BuildState.SetBuildStage("Creating the syscoin conf files")
	out, err := handleConf(tn, sysconf)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	tn.BuildState.IncrementBuildProgress()
	fmt.Printf("done\n")

	tn.BuildState.SetBuildStage("Launching the nodes")
	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, _ *db.Server, localNodeNum int, _ int) error {
		defer tn.BuildState.IncrementBuildProgress()
		return client.DockerExecdLog(localNodeNum,
			"syscoind -conf=\"/syscoin/datadir/regtest.conf\" -datadir=\"/syscoin/datadir/\"")
	})

	return out, err
}

func Add(tn *testnet.TestNet) ([]string, error) {
	return nil, nil
}

func handleConf(tn *testnet.TestNet, sysconf *SysConf) ([]string, error) {
	ips := []string{}
	for _, node := range tn.Nodes {
		ips = append(ips, node.Ip)
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
		log.Println(err)
		return nil, err
	}

	err = helpers.AllNodeExecCon(tn, func(client *ssh.Client, server *db.Server, localNodeNum int, absoluteNodeNum int) error {
		defer tn.BuildState.IncrementBuildProgress()
		_, err := client.DockerExec(localNodeNum, "mkdir -p /syscoin/datadir")
		return err
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//Finally generate the configuration for each node
	mux := sync.Mutex{}
	labels := make([]string, len(ips))
	err = helpers.CreateConfigs(tn, "/syscoin/datadir/regtest.conf",
		func(serverID int, localNodeNum int, absoluteNodeNum int) ([]byte, error) {
			defer tn.BuildState.IncrementBuildProgress()
			confData := ""
			maxConns := 1
			label := ""
			if absoluteNodeNum < noMasterNodes { //Master Node
				confData += sysconf.GenerateMN()
				label = "Master Node"
			} else if absoluteNodeNum%2 == 0 { //Sender
				confData += sysconf.GenerateSender()
				label = "Sender"
			} else { //Receiver
				confData += sysconf.GenerateReceiver()
				label = "Receiver"
			}

			mux.Lock()
			labels[absoluteNodeNum] = label
			mux.Unlock()
			confData += "rpcport=8369\nport=8370\n"
			for _, conn := range connsDist[absoluteNodeNum] {
				confData += fmt.Sprintf("connect=%s:8370\n", conn)
				maxConns += 4
			}
			confData += "rpcallowip=0.0.0.0/0\n"
			//confData += fmt.Sprintf("maxconnections=%d\n",maxConns)
			return []byte(confData), nil
		})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return labels, nil
}
