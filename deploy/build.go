package deploy

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	db "../db"
	util "../util"
)

var conf *util.Config = util.GetConfig()
/**
 * Builds out the Docker Network on pre-setup servers
 * Returns a string of all of the IP addresses 
 */
func Build(buildConf *Config,servers []db.Server) []db.Server {
	var sem	= semaphore.NewWeighted(util.ThreadLimit)
	
	ctx := context.TODO()
	Prepare(buildConf.Nodes,servers)
	fmt.Println("-------------Building The Docker Containers-------------")
	n := buildConf.Nodes
	i := 0

	for n > 0 && i < len(servers){
		fmt.Printf("-------------Building on Server %d-------------\n",i)
		max_nodes := int(servers[i].Max - servers[i].Nodes)
		var nodes int
		if max_nodes > n {
			nodes = n
		}else{
			nodes = max_nodes
		}
		for j := 0; j < nodes; j++ {
			servers[i].Ips = append(servers[i].Ips,util.GetNodeIP(servers[i].ServerID,j))
		}
		prepareVlans(servers[i], nodes)
		var startCmd string
		fmt.Printf("Creating the docker containers on server %d\n",i)

		if conf.Builder == "local deploy legacy"{
			startCmd = fmt.Sprintf("~/local_deploy/whiteblock -n %d -i %s -s %d -a %d -b %d -c %d -S",
				nodes,
				buildConf.Image,
				servers[i].ServerID,
				util.ServerBits,
				util.ClusterBits,
				util.NodeBits)
		}else if conf.Builder == "local deploy" {
			startCmd = fmt.Sprintf("~/local_deploy/deploy -n %d -i %s -s %d -a %d -b %d -c %d -S",
				nodes,
				buildConf.Image,
				servers[i].ServerID,
				util.ServerBits,
				util.ClusterBits,
				util.NodeBits)
		}else if conf.Builder == "umba" {
			startCmd = fmt.Sprintf("~/umba/umba -n %d -i %s -s %d -I %s",
				nodes,
				buildConf.Image,
				servers[i].ServerID,
				servers[i].Iface)
		}else{
			panic("Invalid builder")
		}
		//Acquire resources
		if sem.Acquire(ctx,1) != nil {
			panic("Semaphore Error")
		}
		go func(server string,startCmd string){
			util.SshExec(server,startCmd)
			//Release the resource
			sem.Release(1)
		}(servers[i].Addr,startCmd)

		n -= nodes
		i++
	}
	//Acquire all of the resources here, then release and destroy
	if sem.Acquire(ctx,util.ThreadLimit) != nil {
		panic("Semaphore Error")
	}
	sem.Release(util.ThreadLimit)
	if n != 0 {
		fmt.Printf("ERROR: Only able to build %d/%d nodes\n",(buildConf.Nodes - n),buildConf.Nodes)
	}

	return servers
}


func prepareVlans(server db.Server, nodes int) {

	if conf.Builder == "local deploy" {
		util.SshExecIgnore(server.Addr,"~/local_deploy/deploy -k")
		cmd := fmt.Sprintf("cd ~/local_deploy && ./vlan -B && ./vlan -s %d -n %d -a %d -b %d -c %d -i %s", server.ServerID, nodes, util.ServerBits, util.ClusterBits, util.NodeBits, server.Iface)
		util.SshExec(server.Addr, cmd)
	}else if conf.Builder == "local deploy legacy" {
		util.SshExecIgnore(server.Addr,"~/local_deploy/whiteblock -k")
		cmd := fmt.Sprintf("cd ~/local_deploy && ./vlan -B && ./vlan -s %d -n %d -a %d -b %d -c %d -i %s", server.ServerID, nodes, util.ServerBits, util.ClusterBits, util.NodeBits, server.Iface)
		util.SshExec(server.Addr, cmd)
	}
}
