package deploy

import (
    vyos "../vyos"
    db "../db"
    util "../util"
    "regexp"
    "fmt"
    "log"
)

func getConfig(host string) (*vyos.Config, string,error) {

    data,err := util.SshExec(host,"cat /config/config.boot")
    if err != nil{
        return nil,"",err
    }
    config := vyos.NewConfig(data)
    metaPattern := regexp.MustCompile(`\/\*[^(*\/)]*\*\/`)
    metaResults := metaPattern.FindAllString(data,-1)
    meta := ""
    for _,met := range metaResults {
        meta += fmt.Sprintf("%s\n",met)
    }
    return config,meta,nil
}

func PrepareSwitches(server db.Server,nodes int) error {
    //Assume one switch per server
    if server.Switches[0].Brand & 0x03 == util.Hp {
        setupHPSwitch(server,nodes)
        return nil
    }
    config,_,err := getConfig(server.Switches[0].Addr)
    gws := util.GetGateways(server.ServerID, nodes)
    config.RemoveVifs(server.Switches[0].Iface)
    config.SetIfaceAddr(server.Switches[0].Iface,fmt.Sprintf("%s/%d",server.Iaddr.Gateway,server.Iaddr.Subnet))//Update this later on to be more dynamic
    for i,gw := range gws {
        config.AddVif(
            fmt.Sprintf("%d",i+conf.NetworkVlanStart),
            fmt.Sprintf("%s/%d",gw,util.GetSubnet()),
            server.Switches[0].Iface)
    }
    config.AddVif(
        fmt.Sprintf("%d",conf.ServiceVlan),
        conf.ServiceNetwork,
        server.Switches[0].Iface)
    //fmt.Printf(config.ToString())
    //fmt.Printf(meta)
    err = util.Write("install.sh",GenerateFile(server))
    if err != nil{
        log.Println(err)
        return err
    }
    defer util.Rm("install.sh")
    err = util.Write("config.boot",fmt.Sprintf("%s\n",config.ToString()))
    if err != nil{
        log.Println(err)
        return err
    }
    defer util.Rm("config.boot")
    err = util.Scp(server.Switches[0].Addr,"./config.boot","/config/config.boot")
    if err != nil{
        log.Println(err)
        return err
    }
    util.Scp(server.Switches[0].Addr,"./install.sh",conf.VyosHomeDir+"/install.sh")
    if err != nil{
        log.Println(err)
        return err
    }
    _,err = util.SshExec(server.Switches[0].Addr,"chmod +x ./install.sh && ./install.sh")
    if err != nil {
        log.Println(err)
        return err
    }
    

    return nil
}


func GenerateFile(server db.Server) string {
    if !conf.SetupMasquerade {
        return util.CombineConfig([]string{
            "#!/bin/vbash",
            "source /opt/vyatta/etc/functions/script-template",
            "configure",
            "/bin/cli-shell-api loadFile /config/config.boot",
            "commit",
            "save",
        })
    }
    _,subnet,err := util.GetServiceNetwork()
    if err != nil{
        log.Println(err)
        panic(err)
    }
    ruleid := server.Id * 2

    eth := server.Switches[0].Brand >> 2

    return util.CombineConfig([]string{
        "#!/bin/vbash",
        "source /opt/vyatta/etc/functions/script-template",
        "configure",
        "/bin/cli-shell-api loadFile /config/config.boot",
        fmt.Sprintf("delete nat source rule %d",ruleid),
        fmt.Sprintf("delete nat source rule %d",ruleid+1),
        fmt.Sprintf("set nat source rule %d source address %s",ruleid,fmt.Sprintf("%s/%d",util.GetWholeNetworkIp(server.ServerID),
                        32-(conf.ClusterBits + conf.NodeBits ) )),
        fmt.Sprintf("set nat source rule %d outbound-interface eth%d",ruleid,eth),
        fmt.Sprintf("set nat source rule %d translation address masquerade",ruleid),
        fmt.Sprintf("set nat source rule %d protocol all",ruleid),

        fmt.Sprintf("set nat source rule %d source address %s",ruleid+1,subnet),
        fmt.Sprintf("set nat source rule %d outbound-interface eth%d",ruleid+1,eth),
        fmt.Sprintf("set nat source rule %d translation address masquerade",ruleid+1),
        fmt.Sprintf("set nat source rule %d protocol all",ruleid+1),
        
        "commit",
        "save",
    })
}
