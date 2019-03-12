package syscoin

import(
	"log"
	"io/ioutil"
	"encoding/json"
	util "../../util"
)

type SysConf struct {
	Options			[]string	`json:"options"`
	Extras			[]string	`json:"extras"`

	SenderOptions	[]string	`json:"senderOptions"`
	ReceiverOptions	[]string	`json:"receiverOptions"`
	MNOptions		[]string	`json:"mnOptions"`

	SenderExtras	[]string	`json:"senderExtras"`
	ReceiverExtras	[]string	`json:"receiverExtras"`
	MNExtras		[]string	`json:"mnExtras"`

	MasterNodeConns	int64		`json:"masterNodeConns"`
	NodeConns		int64		`json:"nodeConns"`
	PercOfMNodes	int64		`json:"percentMasternodes"`
	Validators		int64		`json:"validators"`
}


func NewConf(data map[string]interface{}) (*SysConf,error) {
	out := new(SysConf)
	err := json.Unmarshal([]byte(GetDefaults()),out)

	if data == nil {
		log.Println("No params given")
		return out, nil
	}

	err = util.GetJSONStringArr(data,"options",&out.Options)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONStringArr(data,"extras",&out.Extras)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONStringArr(data,"senderOptions",&out.SenderOptions)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONStringArr(data,"senderExtras",&out.SenderExtras)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONStringArr(data,"senderExtras",&out.SenderExtras)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONStringArr(data,"receiverOptions",&out.ReceiverOptions)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONStringArr(data,"receiverExtras",&out.ReceiverExtras)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONStringArr(data,"mnOptions",&out.MNOptions)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONStringArr(data,"mnExtras",&out.MNExtras)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONInt64(data,"masterNodeConns",&out.MasterNodeConns)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONInt64(data,"nodeConns",&out.NodeConns)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONInt64(data,"percentMasternodes",&out.PercOfMNodes)
	if err != nil {
		return nil,err
	}
	log.Printf("%+v\n",*out)
	return out, nil
} 

func (this *SysConf) Generate() string {
	out := ""
	for _,opt := range this.Options {
		out += opt +"=1\n"
	}
	out += "[regtest]\n"
	out += "rpcuser=user\n"
	out += "rpcpassword=password\n"
	for _,extra := range this.Extras {
		extra += extra +"\n"
	}

	return out
}

func (this *SysConf) GenerateReceiver() string {
	out := this.Generate()

	for _, opt := range this.ReceiverOptions {
		out += opt + "=1\n"
	}

	for _,extra := range this.ReceiverExtras {
		extra += extra+"\n"
	}
	return out
}

func (this *SysConf) GenerateSender() string {
	out := this.Generate()

	for _, opt := range this.SenderOptions {
		out += opt + "=1\n"
	}

	for _,extra := range this.SenderExtras {
		extra += extra+"\n"
	}
	return out
}

func (this *SysConf) GenerateMN() string {
	out := this.Generate()

	for _, opt := range this.MNOptions {
		out += opt + "=1\n"
	}

	for _,extra := range this.MNExtras {
		extra += extra+"\n"
	}
	return out
}


func GetParams() string {
    dat, err := ioutil.ReadFile("./resources/syscoin/params.json")
    if err != nil {
        panic(err)//Missing required files is a fatal error
    }
    return string(dat)
}

func GetDefaults() string {
    dat, err := ioutil.ReadFile("./resources/syscoin/defaults.json")
    if err != nil {
        panic(err)//Missing required files is a fatal error
    }
    return string(dat)
}

func GetServices() []util.Service {
	return nil
    /*return []util.Service{
        util.Service{
            Name:"Alpine",
            Image:"alpine:latest",
            Env:map[string]string{
                "HELLO":"HI",
                "INFLUXDB_URL":conf.Influx,
            },
        },
    }*/
}