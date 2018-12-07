package syscoin

import(
	"errors"
	"encoding/json"
)

type SysConf struct {
	RpcUser			string		`json:"rpcUser"`
	RpcPass			string 		`json:"rpcPass"`
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

	
}


func NewConf(data map[string]interface{}) (*SysConf,error) {
	out := new(SysConf)
	out.RpcUser = "appo"
	out.RpcPass = "w@ntest"

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
	out.MNExtras = []string{}

	out.MasterNodeConns = 25
	out.NodeConns = 8
	out.PercOfMNodes = 90

	if data == nil {
		return out, nil
	}
	var err error

	rpcUser,exists := data["rpcUser"]
	if exists {
		switch rpcUser.(type){
			case string:
				out.RpcUser = rpcUser.(string)
			default:
				return nil,errors.New("Incorrect type for rpcUser given")
		}
	}

	rpcPass,exists := data["rpcPass"]
	if exists {
		switch rpcPass.(type) {
			case string:
				out.RpcPass = rpcPass.(string)
			default:
				return nil,errors.New("Incorrect type for rpcPass given")
		}
	}

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
	
	masterNodeConns,exists := data["masterNodeConns"]
	if exists && masterNodeConns != nil {
		switch masterNodeConns.(type){
			case json.Number:
				out.MasterNodeConns,err = masterNodeConns.(json.Number).Int64()
				if err != nil {
					return nil,err
				}
			default:
				return nil,errors.New("Incorrect type for masterNodeConns given")
		}
	}

	nodeConns,exists := data["nodeConns"]
	if exists && nodeConns != nil {
		switch nodeConns.(type){
			case json.Number:
				out.NodeConns,err = nodeConns.(json.Number).Int64()
				if err != nil {
					return nil,err
				}
			default:
				return nil,errors.New("Incorrect type for nodeConns given")
		}
	}

	percOfMNodes,exists := data["percOfMNodes"]
	if exists && percOfMNodes != nil {
		switch percOfMNodes.(type){
			case json.Number:
				out.PercOfMNodes,err = percOfMNodes.(json.Number).Int64()
				if err != nil {
					return nil,err
				}
			default:
				return nil,errors.New("Incorrect type for percOfMNodes given")
		}
	}		

	return out, nil
} 

func (this *SysConf) Generate() string {
	out := "rpcuser="+this.RpcUser+"\n"
	out += "rpcpassword="+this.RpcPass+"\n"
	
	for _,opt := range this.Options {
		out += opt +"=1\n"
	}

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
	{"rpcUser":"string"},
	{"rpcPass":"string"},
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
	{"percentMasternodes","int"}
]`
}