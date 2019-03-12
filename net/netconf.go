/*
    Allows the ability for the simulation of network conditions accross nodes.
 */
package netconf

import(
    "fmt"
    "log"
    util "../util"
)
/**
 [ limit PACKETS ]
 [ delay TIME [ JITTER [CORRELATION]]]
 [ distribution {uniform|normal|pareto|paretonormal} ]
 [ corrupt PERCENT [CORRELATION]]
 [ duplicate PERCENT [CORRELATION]]
 [ loss random PERCENT [CORRELATION]]
 [ loss state P13 [P31 [P32 [P23 P14]]]
 [ loss gemodel PERCENT [R [1-H [1-K]]]
 [ ecn ]
 [ reorder PRECENT [CORRELATION] [ gap DISTANCE ]]
 [ rate RATE [PACKETOVERHEAD] [CELLSIZE] [CELLOVERHEAD]]
 */

var conf *util.Config = util.GetConfig()

type Netconf struct {
    Node            int     `json:"node"`  
    Limit           int     `json:"limit"`
    Loss            float64 `json:"loss"`//Loss % ie 100% = 100
    Delay           int     `json:"delay"`
    Rate            string  `json:"rate"`
    Duplication     float64 `json:"duplicate"`
    Corrupt         float64 `json:"corrupt"`
    Reorder         float64 `json:"reorder"`
}

/*
    CreateCommands generates the commands needed to obtain the desired
    network conditions
 */
func CreateCommands(netconf Netconf,serverId int) []string {
    const offset int = 6
    out := []string{
        fmt.Sprintf("sudo tc qdisc del dev %s%d root",conf.BridgePrefix,netconf.Node),
        fmt.Sprintf("sudo tc qdisc add dev %s%d root handle 1: prio",conf.BridgePrefix,netconf.Node),
        fmt.Sprintf("sudo tc qdisc add dev %s%d parent 1:1 handle 2: netem",conf.BridgePrefix,netconf.Node),//unf
        fmt.Sprintf("sudo tc filter add dev %s%d parent 1:0 protocol ip pref 55 handle %d fw flowid 2:1",
                    conf.BridgePrefix,netconf.Node,offset),
        fmt.Sprintf("sudo iptables -t mangle -A PREROUTING  ! -d %s -j MARK --set-mark %d",
            util.GetGateway(serverId,netconf.Node),offset),
    }
    
    if netconf.Limit > 0 {
        out[2] += fmt.Sprintf(" limit %d",netconf.Limit)
    }

    if netconf.Loss > 0 {
        out[2] += fmt.Sprintf(" loss %.4f",netconf.Loss)
    }

    if netconf.Delay > 0 {
        out[2] += fmt.Sprintf(" delay %dus",netconf.Delay)
    }

    if len(netconf.Rate) > 0 {
        out[2] += fmt.Sprintf(" rate %s",netconf.Rate)
    }

    if netconf.Duplication > 0 {
        out[2] += fmt.Sprintf(" duplicate %.4f",netconf.Duplication)
    }

    if netconf.Corrupt > 0 {
        out[2] += fmt.Sprintf(" corrupt %.4f",netconf.Duplication)
    }

    if netconf.Reorder > 0 {
        out[2] += fmt.Sprintf(" reorder %.4f",netconf.Reorder)
    }

    return out
}



/*
    Apply applies the given network config. 
 */
func Apply(client *util.SshClient,netconf Netconf,serverId int) error {
    cmds := CreateCommands(netconf,serverId)
    for i,cmd := range cmds {
        res,err := client.Run(cmd)
        if i == 0 {
            //Don't check the success of the first command which clears
            continue
        }
        if err != nil {
            log.Println(res)
            log.Println(err)
            return err
        }
    }
    return nil
}

/*
    ApplyToAll applies the given netconf to `nodes` nodes in the network on the given server
 */
func ApplyToAll(client *util.SshClient,netconf Netconf,serverId int,nodes int) error {
    for i:=0;i<nodes;i++{
        netconf.Node = i
        cmds := CreateCommands(netconf,serverId)
        for i,cmd := range cmds {
            res,err := client.Run(cmd)
            if i == 0 {
                //Don't check the success of the first command which clears
                continue
            }
            if err != nil {
                log.Println(res)
                log.Println(err)
                return err
            }
        }
    }
    return nil   
}

/*
    ApplyAll applies all of the given netconfs
 */
func ApplyAll(client *util.SshClient,netconfs []Netconf,serverId int) error {
    for _,netconf := range netconfs {
        err := Apply(client,netconf,serverId)
        if err != nil{
            log.Println(err)
            return err
        }
    }
    return nil
}

/*
    RemoveAll removes network conditions from the given number of nodes
 */
func RemoveAll(client *util.SshClient,nodes int){
    for i := 0; i < nodes; i++ {
         client.Run(fmt.Sprintf("sudo tc qdisc del dev %s%d root prio",
                                conf.BridgePrefix,
                                i))
    }
}