package main

import (
	vyos "./vyos"
	db "./db"
	"regexp"
	"fmt"
)



func getConfig(host string) (*vyos.Config, string) {

	data := sshExec(host,"cat /config/config.boot")
	conf := vyos.NewConfig(data)
	metaPattern := regexp.MustCompile(`\/\*[^(*\/)]*\*\/`)
	metaResults := metaPattern.FindAllString(data,-1)
	meta := ""
	for _,met := range metaResults {
		meta += fmt.Sprintf("%s\n",met)
	}
	return conf,meta
}

func prepareSwitches(server db.Server,nodes int){
	//Assume one switch per server
	if server.Switches[0].Brand == HP {
		setupHPSwitch(server,nodes)
		return
	}
	conf,meta := getConfig(server.Switches[0].Addr)
	gws := getGateways(server.Id, nodes)
	conf.RemoveVifs(server.Switches[0].Iface)
	conf.SetIfaceAddr(server.Switches[0].Iface,fmt.Sprintf("%s/%d",server.Iaddr.Gateway,server.Iaddr.Subnet))//Update this later on to be more dynamic
	for i,gw := range gws {
		conf.AddVif(
			fmt.Sprintf("%d",i+101),
			fmt.Sprintf("%s/%d",gw,getSubnet()),
			server.Switches[0].Iface)
	}
	//fmt.Printf(conf.ToString())
	//fmt.Printf(meta)
	write("config.boot",fmt.Sprintf("%s\n%s",conf.ToString(),meta))
	_scp(server.Switches[0].Addr,"./config.boot","/config/config.boot")
	_scp(server.Switches[0].Addr,"./install.sh","/home/appo/install.sh")
	sshExec(server.Switches[0].Addr,"chmod +x ./install.sh && ./install.sh")
	rm("config.boot")
}

func prepareVlans(server db.Server,nodes int){
	cmd := fmt.Sprintf("cd local_deploy && ./whiteblock -k && ./vlan -B && ./vlan -s %d -n %d -a %d -b %d -c %d -i %s",server.Id,nodes,SERVER_BITS,CLUSTER_BITS,NODE_BITS,server.Iface)
	sshExec(server.Addr,cmd)
}