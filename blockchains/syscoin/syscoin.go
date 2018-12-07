package syscoin

import (
	"context"
	"fmt"
	"log"
	"golang.org/x/sync/semaphore"
	"errors"
	util "../../util"
	db "../../db"
	state "../../state"
)

var conf *util.Config

func init(){
	conf = util.GetConfig()
}

/**
 * Sets up Syscoin Testnet in Regtest mode
 * @param {[type]} data    map[string]interface{} 	The configuration optiosn given by the client
 * @param {[type]} nodes   int                      The number of nodes to build
 * @param {[type]} servers []db.Server) 			The servers to be built on       
 * @return ([]string,error [description]
 */
func RegTest(data map[string]interface{},nodes int,servers []db.Server) ([]string,error) {
	if nodes < 3 {
		log.Println("Tried to build syscoin with not enough nodes")
		return nil,errors.New("Tried to build syscoin with not enough nodes")
	}
	sem3 := semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()
	sysconf,err := NewConf(data)
	if err != nil {
		return nil,err
	}
	defer func(){
		fmt.Printf("Cleaning up...")
		util.Rm("config.boot")
		fmt.Printf("done\n")
	}()
	state.SetBuildSteps(1+(1*nodes))

	fmt.Println("-------------Setting Up Syscoin-------------")
	
	fmt.Printf("Creating the syscoin conf files...")
	out,err := handleConf(servers,sysconf)
	if err != nil {
		return nil,err
	}
	state.IncrementBuildProgress()
	fmt.Printf("done\n")


	fmt.Printf("Launching the nodes")
	for _,server := range servers {
		sem3.Acquire(ctx,1)
		go func(server db.Server){
			for j,_ := range server.Ips {
				fmt.Printf(".")
				container := fmt.Sprintf("whiteblock-node%d",j)
				execCmd := fmt.Sprintf("docker exec %s syscoind -daemon -conf=\"/syscoin/datadir/regtest.conf\" -datadir=\"/syscoin/datadir/\"",container)
				util.SshExec(server.Addr,execCmd)
				state.IncrementBuildProgress()
			}
			sem3.Release(1)

		}(server)
	}

	err = sem3.Acquire(ctx,conf.ThreadLimit)
	if err != nil{
		return nil,err
	}
	fmt.Printf("done\n")
	sem3.Release(conf.ThreadLimit)


	return out,nil 
}



func handleConf(servers []db.Server, sysconf *SysConf) ([]string,error) {
	ips := []string{}
	for _,server := range servers {
		for _, ip := range server.Ips {
			ips = append(ips,ip)
		}
	}

	noMasterNodes := int(float64(len(ips)) * (float64(sysconf.PercOfMNodes)/float64(100)))

	if (len(ips) - noMasterNodes) == 0 {
		log.Println("Warning: No sender/receiver nodes availible. Removing 2 master nodes and setting them as sender/receiver")
		noMasterNodes -= 2;
	}else if (len(ips) - noMasterNodes) % 2 != 0 {
		log.Println("Warning: Removing a master node to keep senders and receivers equal")
		noMasterNodes--;
		if noMasterNodes < 0 {
			log.Println("Warning: Attempt to remove a master node failed, adding one instead")
			noMasterNodes += 2
		}
	}

	connDistModel := make([]int,len(ips))
	for i := 0; i < len(ips); i++ {
		if i < noMasterNodes {
			connDistModel[i] = int(sysconf.MasterNodeConns)
		}else{
			connDistModel[i] = int(sysconf.NodeConns)
		}
	}

	connsDist,err := util.Distribute(ips,connDistModel)
	if err != nil {
		return nil,err
	}
	//Finally generate the configuration for each node
	sem := semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()
	node := 0
	labels := make([]string,len(ips))
	for _,server := range servers {
		for _,_ = range server.Ips{
			sem.Acquire(ctx,1)
			go func(node int){
				confData := ""
				maxConns := 1
				if node < noMasterNodes{//Master Node
					confData += sysconf.GenerateMN()
					labels[node] = "Master Node"
				}else if node%2 == 0 {//Sender
					confData += sysconf.GenerateSender()
					labels[node] = "Sender"
				}else{//Receiver
					confData += sysconf.GenerateReceiver()
					labels[node] = "Receiver"
				}
				confData += "rpcport=8369\nport=8370\n"
				for _, conn := range connsDist[node]{
					confData += fmt.Sprintf("connect=%s:8370\n",conn)
					maxConns += 4
				}
				confData += "rpcallowip=10.0.0.0/8\n"
				confData += fmt.Sprintf("maxconnections=%d\n",maxConns)
				util.Write(fmt.Sprintf("./regtest%d.conf",node),confData)

				util.Scp(server.Addr,fmt.Sprintf("./regtest%d.conf",node),fmt.Sprintf("/home/appo/regtest%d.conf",node))
				container := fmt.Sprintf("whiteblock-node%d",node)
				util.SshExec(server.Addr,fmt.Sprintf("docker exec %s mkdir -p /syscoin/datadir",container))
				util.SshExec(server.Addr,fmt.Sprintf("docker cp /home/appo/regtest%d.conf %s:/syscoin/datadir/regtest.conf",node,container))
				util.Rm(fmt.Sprintf("./regtest%d.conf",node))
				sem.Release(1)
				
			}(node)
			node++
		}
	}

	return labels,nil
}