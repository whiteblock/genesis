package deploy

/**
 * Prepare the vlans and the switch
 */
import (
    "fmt"
    db "../db"
    util "../util"
)

/**
 * Prepare the switches
 * @param  {[type]} noNodes int           [description]
 * @param  {[type]} servers []Server      [description]
 * @return {[type]}         [description]
 */
func Prepare(noNodes int,servers []db.Server){
    fmt.Println("-------------Setting Up Servers-------------")

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
        if servers[i].Switches != nil && len(servers[i].Switches) > 0 {
            PrepareSwitches(servers[i],nodes)
            fmt.Printf("Set up the Switch\n")
        }
        
        n -= nodes
        i++
    }
}