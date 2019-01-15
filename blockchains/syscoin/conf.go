package syscoin

import(
	"errors"
	"log"
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

	out.Options = []string{
		"server",
		"regtest",
		"listen",
		"rest",
		//"debug",
		/*"unittest",
		"addressindex",
		"assetallocationindex",
		"tpstest",*/
	}

	out.SenderOptions = []string{
		"tpstest",
		"addressindex",
	}

	out.ReceiverOptions = []string{
		"tpstest",
	}

	out.MNExtras = []string{}

	out.Extras = []string{}
	out.SenderExtras = []string{}
	out.ReceiverExtras = []string{}
	out.MNOptions = []string{}

	out.MasterNodeConns = 25
	out.NodeConns = 8
	out.PercOfMNodes = 90

	if data == nil {
		log.Println("No params given")
		return out, nil
	}
	var err error


	options,exists := data["options"]
	if exists && options != nil {
		switch options.(type){
			case []string:
				out.Options = options.([]string)
			default:
				return nil,errors.New("Incorrect type for options given")
		}
	}

	extras,exists := data["extras"]
	if exists && extras != nil {
		switch extras.(type){
			case []string:
				out.Extras = extras.([]string)
			default:
				return nil,errors.New("Incorrect type for extras given")
		}
	}

	senderOptions,exists := data["senderOptions"]
	if exists && senderOptions != nil {
		switch senderOptions.(type){
			case []string:
				out.SenderOptions = senderOptions.([]string)
			default:
				return nil,errors.New("Incorrect type for senderOptions given")
		}
	}

	senderExtras,exists := data["senderExtras"]
	if exists && senderExtras != nil {
		switch senderExtras.(type){
			case []string:
				out.SenderExtras = senderExtras.([]string)
			default:
				return nil,errors.New("Incorrect type for senderExtras given")
		}
	}

	receiverOptions,exists := data["receiverOptions"]
	if exists && receiverOptions != nil {
		switch receiverOptions.(type){
			case []string:
				out.ReceiverOptions = receiverOptions.([]string)
			default:
				return nil,errors.New("Incorrect type for receiverOptions given")
		}
	}
	
	receiverExtras,exists := data["receiverExtras"]
	if exists && receiverExtras != nil {
		switch receiverExtras.(type){
			case []string:
				out.ReceiverExtras = receiverExtras.([]string)
			default:
				return nil,errors.New("Incorrect type for receiverExtras given")
		}
	}

	mnOptions,exists := data["mnOptions"]
	if exists && mnOptions != nil {
		switch mnOptions.(type){
			case []string:
				out.MNOptions = mnOptions.([]string)
			default:
				return nil,errors.New("Incorrect type for mnOptions given")
		}
	}
	mnExtras,exists := data["mnExtras"]
	if exists && mnExtras != nil {
		switch mnExtras.(type){
			case []string:
				out.MNExtras = mnExtras.([]string)
			default:
				return nil,errors.New("Incorrect type for mnExtras given")
		}
	}

	if _,ok := data["masterNodeConns"]; ok {
		out.MasterNodeConns,err = util.GetJSONInt64(data,"masterNodeConns")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["nodeConns"]; ok {
		out.NodeConns,err = util.GetJSONInt64(data,"nodeConns")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["percentMasternodes"]; ok {
		out.PercOfMNodes,err = util.GetJSONInt64(data,"percentMasternodes")
		if err != nil {
			return nil,err
		}
	}

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
	return `[
	{"options":"[]string"},
	{"extras":"[]string"},
	{"senderOptions":"[]string"},
	{"receiverOptions":"[]string"},
	{"mnOptions":"[]string"},
	{"senderExtras":"[]string"},
	{"receiverExtras":"[]string"},
	{"mnExtras":"[]string"},
	{"masterNodeConns":"int"},
	{"nodeConns":"int"},
	{"percentMasternodes":"int"}
]`
}

func GetDefaults() string {
	return `{
	"options":[
		"server",
		"regtest",
		"listen",
		"rest"
	],
	"extras":[],
	"senderOptions":[
		"tpstest",
		"addressindex"
	],
	"receiverOptions":[
		"tpstest"
	],
	"mnOptions":[],
	"senderExtras":[],
	"receiverExtras":[],
	"mnExtras":[],
	"masterNodeConns":25,
	"nodeConns":8,
	"percentMasternodes":90
}`
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