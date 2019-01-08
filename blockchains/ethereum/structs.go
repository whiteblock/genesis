package eth

import (
	"encoding/json"
	util "../../util"
	"errors"
)

type EthConf struct {
	ChainId			int64	`json:"chainId"`
	NetworkId		int64	`json:"networkId"`
	Difficulty		int64	`json:"difficulty"`
	InitBalance		string	`json:"initBalance"`
	MaxPeers		int64	`json:"maxPeers"`
	GasLimit		int64	`json:"gasLimit"`
	HomesteadBlock	int64	`json:"homesteadBlock"`
	Eip155Block		int64	`json:"eip155Block"`
	Eip158Block		int64	`json:"eip158Block"`
}

/**
 * Fills in the defaults for missing parts,
 */
func NewConf(data map[string]interface{}) (*EthConf,error) {
	out := new(EthConf)

	out.ChainId = 15468
	out.NetworkId = 15468
	out.Difficulty = 100000
	out.InitBalance = "100000000000000000000"
	out.MaxPeers = 1000
	out.GasLimit = 4000000
	out.HomesteadBlock = 0
	out.Eip155Block = 0
	out.Eip158Block = 0

	if data == nil {
		return out,nil
	}
	var err error

	if _,ok := data["chainId"]; ok {
		out.ChainId,err = util.GetJSONInt64(data,"chainId")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["networkId"]; ok {
		out.NetworkId,err = util.GetJSONInt64(data,"networkId")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["difficulty"]; ok {
		out.Difficulty,err = util.GetJSONInt64(data,"difficulty")
		if err != nil {
			return nil,err
		}
	}

	initBalance,exists := data["initBalance"]
	if exists && initBalance != nil {
		switch initBalance.(type){
			case json.Number:
				out.InitBalance = initBalance.(json.Number).String()
			case string:
				out.InitBalance = initBalance.(string)
			default:
				return nil,errors.New("Incorrect type for initBalance given")
		}
	}

	if _,ok := data["maxPeers"]; ok {
		out.MaxPeers,err = util.GetJSONInt64(data,"maxPeers")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["gasLimit"]; ok {
		out.GasLimit,err = util.GetJSONInt64(data,"gasLimit")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["homesteadBlock"]; ok {
		out.HomesteadBlock,err = util.GetJSONInt64(data,"homesteadBlock")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["eip155Block"]; ok {
		out.Eip155Block,err = util.GetJSONInt64(data,"eip155Block")
		if err != nil {
			return nil,err
		}
	}

	if _,ok := data["eip158Block"]; ok {
		out.Eip158Block,err = util.GetJSONInt64(data,"eip158Block")
		if err != nil {
			return nil,err
		}
	}

	return out,nil
}


func GetParams() string {
	return `[
	{"chainId":"int"},
	{"networkId":"int"},
	{"difficulty":"int"},
	{"initBalance":"string"},
	{"maxPeers":"int"},
	{"gasLimit":"int"},
	{"homesteadBlock":"int"},
	{"eip155Block":"int"},
	{"eip158Block":"int"}
]`
}

func GetDefaults() string {
	return `{
	"chainId":15468,
	"networkId":15468,
	"difficulty":100000,
	"initBalance":100000000000000000000,
	"maxPeers":1000,
	"gasLimit":4000000,
	"homesteadBlock":0,
	"eip155Block":0,
	"eip158Block":0
}`
}


func GetServices() []util.Service{
	return nil
}