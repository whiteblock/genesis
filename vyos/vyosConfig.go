package vyos

import (
	"regexp"
	//"fmt"
)
/**
 * Abstraction of vyos configuration file
 * Possibly useful regex
 * (?m)^([A-z|0-9| |\t])*{
 * (?m)^( )*[A-z]*
 * (?m)^[[A-z]([A-z|0-9| |\t])*{
 */

type Config struct{
	interfaces 		[]*NetInterface
	services		[]*Service
	system			*System
}

func NewConfig(data string) *Config{
	startPattern := regexp.MustCompile("(?m)^[A-z]([A-z|0-9| |\t])*{")
	startIndexes := startPattern.FindAllIndex([]byte(data),-1)
	endPattern := regexp.MustCompile("(?m)^}")
	endIndexes := endPattern.FindAllIndex([]byte(data),-1)

	//fmt.Printf("Start indexes are %v\n",startIndexes)
	//fmt.Printf("End index are %v\n",endIndexes)

	parts := []string{}
	for i,start := range startIndexes {
		parts = append(parts,data[start[1]:endIndexes[i][0]])
	}
	//fmt.Printf("Slices are:\n%v\n",parts)
	conf := new(Config)
	conf.interfaces = ParseInterfaces(parts[0])
	conf.services = LoadServices(parts[1])
	conf.system = LoadSystem(parts[2])
	/*fmt.Printf(IfacesToString(conf.interfaces,4))
	fmt.Printf("\n%s\n",ServicesToString(conf.services))
	fmt.Println(conf.system.ToString())*/
	return conf
}



func (this *Config) ToString() string {
	out := "interfaces {\n"
	out += IfacesToString(this.interfaces,4)
	out += "}\n"
	out += "service {\n"
	out += ServicesToString(this.services)
	out += "}\n"
	out += this.system.ToString()
	return out
}

func (this *Config) AddVif(name string,address string,parent string){
	this.GetInterfaceByName(parent).AddVif(name,address)
}

func (this *Config) SetIfaceAddr(name string,address string){
	this.GetInterfaceByName(name).address = address
}

func (this *Config) GetInterfaceByName(name string) *NetInterface {
	for _,iface := range this.interfaces {
		if iface.name.name == name {
			//fmt.Printf("Found interface %s\n",name)
			return iface
		}
	}
	//fmt.Printf("Unable to find interface %s\n",name)
	return nil
}

func (this *Config) RemoveAllVifs() {
	for i,_ := range this.interfaces {
		this.interfaces[i].vlans = []*NetInterface{}
	}
}

func (this *Config) RemoveVifs(parent string){
	iface := this.GetInterfaceByName(parent)
	iface.vlans = []*NetInterface{}
}

func (this *Config) RemoveVif(name string,parent string) {
	iface := this.GetInterfaceByName(parent)
	for i,vlan := range iface.vlans {
		if vlan.name.name == name {
			iface.vlans = append(iface.vlans[:i],iface.vlans[i+1:]...)
			return
		}
	} 
}

func (this *Config) SetGateway(address string) {
	this.system.gateway_address = address
}

func (this *Config) SetHostName(name string) {
	this.system.host_name = name
}
