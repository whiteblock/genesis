package deploy

import (
    vyos "../vyos"
    db "../db"
    util "../util"
    "regexp"
    "fmt"
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
    if server.Switches[0].Brand == util.Hp {
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
    err = util.Write("config.boot",fmt.Sprintf("%s\n",config.ToString()))
    if err != nil{
        return err
    }
    err = util.Scp(server.Switches[0].Addr,"./config.boot","/config/config.boot")
    if err != nil{
        return err
    }
    util.Scp(server.Switches[0].Addr,"./install.sh",conf.VyosHomeDir+"/install.sh")
    if err != nil{
        return err
    }
    _,err = util.SshExec(server.Switches[0].Addr,"chmod +x ./install.sh && ./install.sh")
    if err != nil {
        return err
    }
    util.Rm("config.boot")

    return nil
}