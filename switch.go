package main

import (
	"bitbucket.org/whiteblockio/vyos"
	"regexp"
	"fmt"
)

type Switch struct {
	addr 	string
	iface	string
	brand	int
	id		int
}

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

func prepareSwitches(server Server,nodes int){
	//Assume one switch per server
	if server.switches[0].brand == HP {
		setupHPSwitch(server,nodes)
		return
	}
	conf,meta := getConfig(server.switches[0].addr)
	gws := getGateways(server.id, nodes)
	conf.RemoveVifs(server.switches[0].iface)
	conf.SetIfaceAddr(server.switches[0].iface,fmt.Sprintf("%s/%d",server.iaddr.gateway,server.iaddr.subnet))//Update this later on to be more dynamic
	for i,gw := range gws {
		conf.AddVif(
			fmt.Sprintf("%d",i+101),
			fmt.Sprintf("%s/%d",gw,getSubnet()),
			server.switches[0].iface)
	}
	//fmt.Printf(conf.ToString())
	//fmt.Printf(meta)
	write("config.boot",fmt.Sprintf("%s\n%s",conf.ToString(),meta))
	_scp(server.switches[0].addr,"./config.boot","/config/config.boot")
	_scp(server.switches[0].addr,"./install.sh","/home/appo/install.sh")
	sshExec(server.switches[0].addr,"chmod +x ./install.sh && ./install.sh")
	rm("config.boot")
}

func prepareVlans(server Server,nodes int){
	cmd := fmt.Sprintf("cd local_deploy && ./whiteblock -k && ./vlan -B && ./vlan -s %d -n %d -a %d -b %d -c %d -i %s",server.id,nodes,SERVER_BITS,CLUSTER_BITS,NODE_BITS,server.iface)
	sshExec(server.addr,cmd)
}