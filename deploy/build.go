package deploy

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	db "../db"
	util "../util"
)


var sem	= semaphore.NewWeighted(util.ThreadLimit)


/**
 * Builds out the Docker Network on pre-setup servers
 * Returns a string of all of the IP addresses 
 */
func Build(buildConf *Config,servers []db.Server) []db.Server {

	ctx := context.TODO()
	//Prepare(buildConf.Nodes,servers)
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
			servers[i].Ips = append(servers[i].Ips,util.GetNodeIP(servers[i].Id,j))
		}
		

		startCmd := fmt.Sprintf("~/umba/umba --n %d --i %s --s %d --I %s",
			nodes,
			buildConf.Image,
			servers[i].Id,
			servers[i].Iface)
		//Acquire resources
		if sem.Acquire(ctx,1) != nil {
			panic("Semaphore Error")
		}
		go buildInternalInfrastructure(servers[i].Addr,startCmd)

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

func buildInternalInfrastructure(server string,startCmd string){
	
	util.SshExec(server,startCmd)
	//Release the resource
	sem.Release(1)
}