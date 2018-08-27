package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
)


var sem	= semaphore.NewWeighted(THREAD_LIMIT)


/**
 * Builds out the Docker Network on pre-setup servers
 * Returns a string of all of the IP addresses 
 */
func build(buildConf *Config,_servers []Server) []Server {

	ctx := context.TODO()
	servers := prepare(buildConf.nodes,_servers)
	fmt.Println("-------------Building The Docker Containers-------------")
	n := buildConf.nodes
	i := 0
	for n > 0 && i < len(servers){
		fmt.Printf("-------------Building on Server %d-------------\n",i)
		max_nodes := int(servers[i].max - servers[i].nodes)
		var nodes int

		if max_nodes > n {
			nodes = n
		}else{
			nodes = max_nodes
		}
		for j := 0; j < nodes; j++ {
			servers[i].ips = append(servers[i].ips,getNodeIP(servers[i].id,j))
		}
		

		startCmd := fmt.Sprintf("~/local_deploy/whiteblock -n %d -i %s -s %d -a %d -b %d -c %d -S",
			nodes,
			buildConf.image,
			servers[i].id,
			SERVER_BITS,
			CLUSTER_BITS,
			NODE_BITS)
		//Acquire resources
		if sem.Acquire(ctx,1) != nil {
			panic("Semaphore Error")
		}
		go buildInternalInfrastructure(servers[i].addr,startCmd)

		n -= nodes
		i++
	}
	//Acquire all of the resources here,  then release and destroy
	if sem.Acquire(ctx,THREAD_LIMIT) != nil {
		panic("Semaphore Error")
	}
	sem.Release(THREAD_LIMIT)
	if n != 0 {
		fmt.Printf("ERROR: Only able to build %d/%d nodes\n",(buildConf.nodes - n),buildConf.nodes)
	}

	return servers
}

func buildInternalInfrastructure(server string,startCmd string){
	
	sshExec(server,startCmd)
	//Release the resource
	sem.Release(1)
}