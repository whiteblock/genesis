package vyos

/*import (
    "fmt"
    "regexp"
)


(?m)^nat {(.|(\n ))*\n}
(?m)^( ){4}source {(.|(\n ))*(\n( ){4})} //just get the source
(?m)^( ){4}[^s][^o][^u]?([A-z]*) {(.|(\n ))*(\n( ){4})} //close enough to not source
(?m)rule \d* //Get rule
(?m)^( ){8}rule( )?[0-9]* {(\n)(.|(\n( ){12}))*(\n( ){8})} //get the rules

type Nat struct{
    rules   map[int]string
    other   []string
}

    addressPattern := regexp.MustCompile(`\{[A-z|0-9]+\}`)
        addresses := addressPattern.FindAllString(gethResults,-1)
    


func LoadNat(data string) *Nat {
    out := new(Nat)
    
    return out
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

func (this *Nat) AddRule(number int,iface string,address string){
    this.rules[number] = "{\n"+
        fmt.Sprintf("%s%s %s\n",MakeIndent(12),"outbound-interface",iface) + 
        MakeIndent(12)+"protocol all\n"+
        MakeIndent(12)+"source {\n"+
            MakeIndent(16)+"address "+address+"\n"+
        MakeIndent(12)+"}\n"+
        MakeIndent(12)+"translation {\n"+
            MakeIndent(16)+"address masquerade\n"+
        MakeIndent(12)+"}"
}

func (this *Nat) RemoveRule(number int){
    delete(this.rules,number)
}

func (this *Nat) ToString() string{
    if this == null {
        return ""
    }
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
*/
/**
 * Maintain the proper output order 
 */
/*func (this *Nat) Parse(data string,nameLine string) {
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

}*/
