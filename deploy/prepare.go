package deploy

/**
 * Prepare the vlans and the switch
 *
 * 
 */
import (
 	"fmt"
 	//"context"
	//"golang.org/x/sync/semaphore"
	db "../db"
	//util "../util"
)

//var prepareSem = semaphore.NewWeighted(util.ThreadLimit)

func Prepare(noNodes int,servers []db.Server){
	fmt.Println("-------------Setting Up Servers-------------")
	//ctx := context.TODO()
	
	prepareSwitchesThread(noNodes,servers)
	//wait for completion
	fmt.Println("Awaiting completion of prepare part 1")
	//prepareSem.Acquire(ctx,util.ThreadLimit)
	//prepareSem.Release(util.ThreadLimit)

	
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
		PrepareSwitches(servers[i],nodes)
		fmt.Printf("Set up the Switch\n")
		n -= nodes
		i++
	}
}