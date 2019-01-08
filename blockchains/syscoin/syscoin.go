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
func RegTest(data map[string]interface{},nodes int,servers []db.Server,clients []*util.SshClient) ([]string,error) {
	if nodes < 3 {
		log.Println("Tried to build syscoin with not enough nodes")
		return nil,errors.New("Tried to build syscoin with not enough nodes")
	}
	sem3 := semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()
	sysconf,err := NewConf(data)
	if err != nil {
		log.Println(err)
		return nil,err
	}
	defer func(){
		fmt.Printf("Cleaning up...")
		util.Rm("config.boot")
		fmt.Printf("done\n")
	}()
	state.SetBuildSteps(1+(6*nodes))


	fmt.Println("-------------Setting Up Syscoin-------------")
	
	fmt.Printf("Creating the syscoin conf files...")
	out,err := handleConf(servers,clients,sysconf)
	if err != nil {
		log.Println(err)
		return nil,err
	}
	state.IncrementBuildProgress()
	fmt.Printf("done\n")


	fmt.Printf("Launching the nodes")
	for i,server := range servers {
		sem3.Acquire(ctx,1)
		go func(server db.Server,i int){
			defer sem3.Release(1)

			for j,_ := range server.Ips {
				err := clients[i].DockerExecdLog(j,"syscoind -conf=\"/syscoin/datadir/regtest.conf\" -datadir=\"/syscoin/datadir/\"")
				if err != nil {
					state.ReportError(err)
					log.Println(err)
					return
				}
				state.IncrementBuildProgress()
			}
			

		}(server,i)
	}

	err = sem3.Acquire(ctx,conf.ThreadLimit)
	if err != nil{
		log.Println(err)
		return nil,err
	}
	fmt.Printf("done\n")
	sem3.Release(conf.ThreadLimit)

	if !state.ErrorFree() {
		return nil,state.GetError()
	}
	return out,nil 
}

func handleConf(servers []db.Server,clients []*util.SshClient, sysconf *SysConf) ([]string,error) {
	ips := []string{}
	for _,server := range servers {
		for _, ip := range server.Ips {
			ips = append(ips,ip)
		}
	}

	noMasterNodes := int(float64(len(ips)) * (float64(sysconf.PercOfMNodes)/float64(100)))
	//log.Println(fmt.Sprintf("PERC = %d; NUM = %d;",sysconf.PercOfMNodes,noMasterNodes))

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
		log.Println(err)
		return nil,err
	}
	//Finally generate the configuration for each node
	sem := semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()
	node := 0
	labels := make([]string,len(ips))
	for i,server := range servers {
		for _,_ = range server.Ips{
			sem.Acquire(ctx,1)
			go func(node int){
				confData := ""
				maxConns := 1
				if node < noMasterNodes {//Master Node
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
				confData += "rpcallowip=0.0.0.0/0\n"
				confData += fmt.Sprintf("maxconnections=%d\n",maxConns)
				err := util.Write(fmt.Sprintf("./regtest%d.conf",node),confData)
				if err != nil {
					state.ReportError(err)
					log.Println(err)
					return
				}
				err = clients[i].Scp(fmt.Sprintf("./regtest%d.conf",node),fmt.Sprintf("/home/appo/regtest%d.conf",node))
				if err != nil {
					state.ReportError(err)
					log.Println(err)
					return
				}
				state.IncrementBuildProgress()
				container := fmt.Sprintf("whiteblock-node%d",node)
				_,err = clients[i].Run(fmt.Sprintf("docker exec %s mkdir -p /syscoin/datadir",container))
				if err != nil {
					state.ReportError(err)
					log.Println(err)
					return
				}
				state.IncrementBuildProgress()
				_,err = clients[i].Run(fmt.Sprintf("docker cp /home/appo/regtest%d.conf %s:/syscoin/datadir/regtest.conf",node,container))
				if err != nil {
					state.ReportError(err)
					log.Println(err)
					return
				}
				state.IncrementBuildProgress()
				err = util.Rm(fmt.Sprintf("./regtest%d.conf",node))
				if err != nil {
					state.ReportError(err)
					log.Println(err)
					return
				}
				state.IncrementBuildProgress()
				_,err = clients[i].Run(fmt.Sprintf("rm /home/appo/regtest%d.conf",node))
				if err != nil {
					state.ReportError(err)
					log.Println(err)
					return
				}
				state.IncrementBuildProgress()
				sem.Release(1)
				
			}(node)
			node++
		}
	}
	sem.Acquire(ctx,conf.ThreadLimit)
	sem.Release(conf.ThreadLimit)
	if !state.ErrorFree() {
		return nil,state.GetError()
	}
	return labels,nil
}