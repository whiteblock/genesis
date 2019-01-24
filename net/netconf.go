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
    Node    int     `json:"node"`  
    Limit   int     `json:"limit"`
    Loss    float64 `json:"loss"`
    Delay   int     `json:"delay"`
    Rate    string  `json:"rate"`
}

func Apply(client *util.SshClient,netconf Netconf) error {
    clearCmd := fmt.Sprintf("sudo tc qdisc del dev %s%d root netem",
                            conf.BridgePrefix,
                            netconf.Node)
    cmd := fmt.Sprintf("sudo tc qdisc add dev %s%d root netem",
        conf.BridgePrefix,
        netconf.Node)
    if netconf.Limit > 0 {
        cmd += fmt.Sprintf(" limit %d",netconf.Limit)
    }

    if netconf.Loss > 0 {
        cmd += fmt.Sprintf(" loss %.4f",netconf.Loss)
    }

    if netconf.Delay > 0 {
        cmd += fmt.Sprintf(" delay %dms",netconf.Delay)
    }

    if len(netconf.Rate) > 0 {
        cmd += fmt.Sprintf(" rate %s",netconf.Rate)
    }
    client.Run(clearCmd)
    res,err := client.Run(cmd)
    if err != nil {
        log.Println(res)
        log.Println(err)
        return err
    }
    return nil
}

func ApplyToAll(client *util.SshClient,netconf Netconf,nodes int) error {
    for i:=0;i<nodes;i++{
        clearCmd := fmt.Sprintf("sudo tc qdisc del dev %s%d root netem",
                                conf.BridgePrefix,
                                i)
        cmd := fmt.Sprintf("sudo tc qdisc add dev %s%d root netem",
            conf.BridgePrefix,
            i)
        if netconf.Limit > 0 {
            cmd += fmt.Sprintf(" limit %d",netconf.Limit)
        }

        if netconf.Loss > 0 {
            cmd += fmt.Sprintf(" loss %.4f",netconf.Loss)
        }

        if netconf.Delay > 0 {
            cmd += fmt.Sprintf(" delay %dms",netconf.Delay)
        }

        if len(netconf.Rate) > 0 {
            cmd += fmt.Sprintf(" rate %s",netconf.Rate)
        }
        client.Run(clearCmd)
        res,err := client.Run(cmd)
        if err != nil {
            log.Println(res)
            log.Println(err)
            return err
        }
    }
    return nil   
}

func ApplyAll(client *util.SshClient,netconfs []Netconf) error {
    for _,netconf := range netconfs {
        err := Apply(client,netconf)
        if err != nil{
            log.Println(err)
            return err
        }
    }
    return nil
}


func RemoveAll(client *util.SshClient,nodes int){
    for i := 0; i < nodes; i++ {
         client.Run(fmt.Sprintf("sudo tc qdisc del dev %s%d root netem",
                            conf.BridgePrefix,
                            i))
    }
}