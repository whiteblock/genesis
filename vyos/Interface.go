package vyos

import (
	"regexp"
	"fmt"
)

type NetInterface struct{
	name 			InterfaceName
	address 		string
	duplex 			string
	hw_id 			string //hw-id
	smp_affinity 	string
	speed 			string
	vlans			[]*NetInterface
}

type InterfaceName struct{
	iType 	string
	name 	string
}

func ParseInterfaces(data string) []*NetInterface {
	startPattern := regexp.MustCompile("(?m)^( ){3,5}[A-z]([A-z|0-9| |\t])*{")
	startIndexes := startPattern.FindAllIndex([]byte(data),-1)
	endPattern := regexp.MustCompile("(?m)^( ){3,5}}")
	endIndexes := endPattern.FindAllIndex([]byte(data),-1)

	fmt.Printf("Start indexes are %v\n",startIndexes)
	fmt.Printf("End index are %v\n",endIndexes)

	ifaces := []*NetInterface{}
	for i,start := range startIndexes {
		ifaces = append(ifaces,NewNetInterface(data[start[1]:endIndexes[i][0]],data[start[0]:start[1]],8) )
	}
	

	return ifaces
}


func ParseViffs(data string) []*NetInterface {
	startPattern := regexp.MustCompile("(?m)^( ){8}vif([A-z|0-9| |\t])*{")
	startIndexes := startPattern.FindAllIndex([]byte(data),-1)
	endPattern := regexp.MustCompile("(?m)^( ){8}}")
	endIndexes := endPattern.FindAllIndex([]byte(data),-1)

	fmt.Printf("Start indexes are %v\n",startIndexes)
	fmt.Printf("End index are %v\n",endIndexes)

	ifaces := []*NetInterface{}
	for i,start := range startIndexes {
		ifaces = append(ifaces,NewNetInterface(data[start[1]:endIndexes[i][0]],data[start[0]:start[1]],12) )
	}
	fmt.Printf("Interfaces : \n\n %#v \n",ifaces);
	fmt.Printf(IfacesToString(ifaces,4))
	return ifaces
}

func NewNetInterface(data string,name string, indent int) *NetInterface {
	out := new(NetInterface)
	namePattern := regexp.MustCompile("[A-z|0-9]+")
	names := namePattern.FindAllString(name,2)
	out.name = InterfaceName{iType:names[0],name:names[1]}
	fmt.Printf("Given:\n%s\n",data)

	//Grap address
	out.address = GrabValue(indent,"address",data)
	out.duplex = GrabValue(indent,"duplex",data)
	out.hw_id = GrabValue(indent,"hw-id",data)
	out.smp_affinity = GrabValue(indent,"smp_affinity",data)
	out.speed = GrabValue(indent,"speed",data)
	out.vlans = ParseViffs(data)
	fmt.Printf("Out:\n%s\n",out.ToString(indent - 4))
	return out;
}

func IfacesToString(ifaces []*NetInterface,indent int) string {
	out := ""
	for _,iface := range ifaces {
		out += iface.ToString(indent)
	}
	return out;
}

func (this *NetInterface) ToString(indent int) string {
	//fmt.Println("ToString")
	shallowIndent := MakeIndent(indent)
	mediumIndent := MakeIndent(indent + 4)
	out := fmt.Sprintf("%s%s %s {\n",shallowIndent,this.name.iType,this.name.name)
	if len(this.address) != 0 {
		out += fmt.Sprintf("%saddress %s\n",mediumIndent,this.address)
	}
	if len(this.duplex) != 0 {
		out += fmt.Sprintf("%sduplex %s\n",mediumIndent,this.duplex)
	}
	if len(this.hw_id) != 0 {
		out += fmt.Sprintf("%shw-id %s\n",mediumIndent,this.hw_id)
	}
	if len(this.smp_affinity) != 0 {
		out += fmt.Sprintf("%ssmp_affinity %s\n",mediumIndent,this.smp_affinity)
	}
	if len(this.speed) != 0 {
		out += fmt.Sprintf("%sspeed %s\n",mediumIndent,this.speed)
	}
	out += IfacesToString(this.vlans,indent + 4)
	out += fmt.Sprintf("%s}\n",shallowIndent)
	return out
}

func NewVif(name string,address string) *NetInterface {
	out := new(NetInterface)
	out.address = address
	out.name = InterfaceName{iType:"vif",name:name}
	return out
}

func (this *NetInterface) AddVif(name string,address string){
	this.vlans = append(this.vlans,NewVif(name,address))
}


/**
 * NewNetInterface helper functions
 * 	address 		string
	duplex 			string
	hw_id 			string //hw-id
	smp_affinity 	string
	speed 			string
 */

