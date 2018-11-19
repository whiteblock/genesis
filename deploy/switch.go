package deploy

import (
	vyos "../vyos"
	db "../db"
	util "../util"
	"regexp"
	"fmt"
)

func getConfig(host string) (*vyos.Config, string) {

	data := util.SshExec(host,"cat /config/config.boot")
	config := vyos.NewConfig(data)
	metaPattern := regexp.MustCompile(`\/\*[^(*\/)]*\*\/`)
	metaResults := metaPattern.FindAllString(data,-1)
	meta := ""
	for _,met := range metaResults {
		meta += fmt.Sprintf("%s\n",met)
	}
	return config,meta
}

func PrepareSwitches(server db.Server,nodes int){
	//Assume one switch per server
	if server.Switches[0].Brand == util.Hp {
		setupHPSwitch(server,nodes)
		return
	}
	config,meta := getConfig(server.Switches[0].Addr)
	gws := util.GetGateways(server.ServerID, nodes)
	config.RemoveVifs(server.Switches[0].Iface)
	config.SetIfaceAddr(server.Switches[0].Iface,fmt.Sprintf("%s/%d",server.Iaddr.Gateway,server.Iaddr.Subnet))//Update this later on to be more dynamic
	for i,gw := range gws {
		config.AddVif(
			fmt.Sprintf("%d",i+101),
			fmt.Sprintf("%s/%d",gw,util.GetSubnet()),
			server.Switches[0].Iface)
	}
	//fmt.Printf(config.ToString())
	//fmt.Printf(meta)
	util.Write("config.boot",fmt.Sprintf("%s\n%s",config.ToString(),meta))
	util.Scp(server.Switches[0].Addr,"./config.boot","/config/config.boot")
	util.Scp(server.Switches[0].Addr,"./install.sh",conf.VyosHomeDir+"/install.sh")
	util.SshExec(server.Switches[0].Addr,"chmod +x ./install.sh && ./install.sh")
	util.Rm("config.boot")
}