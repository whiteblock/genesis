package syscoin

import(
	"errors"
)

type SysConf struct {
	RpcUser		string		`json:"rpcUser"`
	RpcPass		string 		`json:"rpcPass"`
	Options		[]string	`json:"options"`
	Extras		[]string	`json:"extras"`
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
		"debug",
		"unittest",
		"addressindex",
		"assetallocationindex",
		"tpstest",
	}
	out.Extras = []string{}

	if data == nil {
		return out, nil
	}

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