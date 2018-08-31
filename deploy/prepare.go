package deploy

/**
 * Prepare the vlans and the switch
 *
 * 
 */
import (
 	"fmt"
 	"context"
	"golang.org/x/sync/semaphore"
	db "../db"
)

var prepareSem = semaphore.NewWeighted(THREAD_LIMIT)

func Prepare(noNodes int,servers []db.Server){
	fmt.Println("-------------Setting Up Servers-------------")
	ctx := context.TODO()

	for _, server := range servers {
		prepareSem.Acquire(ctx,1)
		go prepareLocalDeploy(server.Addr)
	}

	n := noNodes
	i := 0
	go prepareSwitchesThread(noNodes,servers)
	//wait for completion
	fmt.Println("Awaiting completion of prepare part 1")
	prepareSem.Acquire(ctx,THREAD_LIMIT)
	prepareSem.Release(THREAD_LIMIT)

	for n > 0 && i < len(servers){
		fmt.Printf("-------------Setting Up Server %d-------------\n",i)
		
		max_nodes := int(servers[i].Max - servers[i].Nodes)
		var nodes int

		if max_nodes > n {
			nodes = n
		}else{
			nodes = max_nodes
		}
		
		prepareSem.Acquire(ctx,1)
		go prepareVlansThread(servers[i],nodes)

		n -= nodes
		i++

	}

	prepareSem.Acquire(ctx,THREAD_LIMIT)
	prepareSem.Release(THREAD_LIMIT)
}

func prepareLocalDeploy(ip string){
	sshExec(ip,"cd ~ && rm -rf local_deploy.tar.gz* local_deploy && wget http://172.16.0.8/local_deploy.tar.gz && tar xf local_deploy.tar.gz && cd ~/local_deploy && make")
	fmt.Printf("Finished Downloading And Installing Local Deploy\n")
	prepareSem.Release(1)
}


func prepareVlansThread(server db.Server,nodes int){
	prepareVlans(server,nodes)
	fmt.Printf("Created the vlans\n")
	prepareSem.Release(1)
}

/**
 * Prepare the switches, this can be done whenever
 * @param  {[type]} noNodes int           [description]
 * @param  {[type]} servers []Server      [description]
 * @return {[type]}         [description]
 */
func prepareSwitchesThread(noNodes int,servers []db.Server){
	n := noNodes
	i := 0
	
	for n > 0 && i < len(servers){
		fmt.Printf("-------------Setting Up Server %d-------------\n",i)
		
		max_nodes := int(servers[i].Max - servers[i].Nodes)
		var nodes int

		if max_nodes > n {
			nodes = n
		}else{
			nodes = max_nodes
		}
		prepareSwitches(servers[i],nodes)
		fmt.Printf("Set up the Switch\n")
		n -= nodes
		i++
	}
}