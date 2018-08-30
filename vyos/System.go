package vyos

import (
	"fmt"
	"regexp"
)


type System struct{
	config_management	string //config-management //(No Modify)
	console 			string //(No Modify) 
	
	gateway_address		string //gateway-address
	host_name			string //host-name
	
	login				string //(No Modify)
	ntp 				string //(No Modify)
	sPackage 			string //package (No modify)
	syslog				string //(No modify)
	time_zone			string //time-zone
	extra				string
}


func LoadSystem(data string) *System {

	system := new(System)
	system.extra = ""

	startPattern := regexp.MustCompile(`(?m)^( ){3,5}[A-z]([A-z|0-9| |\t|\-])*{`)
	startIndexes := startPattern.FindAllIndex([]byte(data),-1)
	endPattern := regexp.MustCompile(`(?m)^( ){3,5}}`)
	endIndexes := endPattern.FindAllIndex([]byte(data),-1)
	//fmt.Printf("Start indexes are %v\n",startIndexes)
	//fmt.Printf("End index are %v\n",endIndexes)
	for i, start := range startIndexes {
		system.Parse(data[start[1]:endIndexes[i][0]],data[start[0]:start[1]])
	}
	system.gateway_address = GrabValue(4,"gateway-address",data)
	system.host_name = GrabValue(4,"host-name",data)
	system.time_zone = GrabValue(4,"time-zone",data)

	return system
}


func (this *System) ToString() string{
	indent := MakeIndent(4)
	out := "system {\n"
	out += this.config_management
	out += this.console;
	out += fmt.Sprintf("%sgateway-address %s\n",indent,this.gateway_address)
	out += fmt.Sprintf("%shost-name %s\n",indent,this.host_name)
	out += this.login
	out += this.ntp
	out += this.sPackage
	out += this.syslog
	out += fmt.Sprintf("%stime-zone %s\n",indent,this.time_zone)
	out += "}\n"
	return out;
}

/**
 * Maintain the proper output order 
 */
func (this *System) Parse(data string,nameLine string) {
	namePattern := regexp.MustCompile(`([A-z|0-9]|\.|\-)+`)
	name := string(namePattern.Find([]byte(nameLine)))
	indent := MakeIndent(4)
	out := fmt.Sprintf("%s%s%s}\n",nameLine,data,indent)
	switch (name) {
		case "config-management":
			this.config_management = out
		case "console":
			this.console = out
		case "login":
			this.login = out
		case "ntp":
			this.ntp = out
		case "package":
			this.sPackage = out
		case "syslog":
			this.syslog = out
		default:
			this.extra += out //If anything unexpected is there, make sure it is still included
	}

}







/**
 * OLD CODE FOR MORE FUNCTIONALITY
 */


/*type ConfigManagement struct{
	commit_revisions	int //commit-revisions
}

type ConsoleDevice struct{
	name 	string
	speed	int 
}

type User struct{
	name 			string
	athentication 	Authentication
}

type Authentication struct{
	encrypted_password	string //encrypted-password
	plaintext_password 	string //plaintext-password
}


type Repository struct{
	name 			string
	components		string
	distribution	string
	password 		string
	url 			string
	username 		string
}*/

/*func LoadConfigManagement(data string,indent int) *ConfigManagement {
	out := new(ConfigManagement)
	out.commit_revisions = GrabValue(indent,"commit-revisions",data)
	return out	
}

func LoadConsole(in string, indent int) *ConsoleDevice {
	out := new(ConsoleDevice)
	dataPattern := regexp.MustCompile(`([A-z|0-9]|\.|\-)+`)
	data := dataPattern.FindAllString(in,-1)
	out.name = data[1]
	out.speed = data[3]
	return out;
}*/

