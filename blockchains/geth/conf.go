package geth

import (
	"encoding/json"
	"io/ioutil"
	"errors"
	util "../../util"
)

type EthConf struct {
	ExtraAccounts   int64	`json:"extraAccounts"`
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
	err := json.Unmarshal([]byte(GetDefaults()),out)

	if data == nil {
		return out,nil
	}

	err = util.GetJSONInt64(data,"extraAccounts",&out.ExtraAccounts)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONInt64(data,"chainId",&out.ChainId)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONInt64(data,"networkId",&out.NetworkId)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONInt64(data,"difficulty",&out.Difficulty)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONInt64(data,"maxPeers",&out.MaxPeers)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONInt64(data,"gasLimit",&out.GasLimit)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONInt64(data,"eip155Block",&out.Eip155Block)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONInt64(data,"homesteadBlock",&out.HomesteadBlock)
	if err != nil {
		return nil,err
	}

	err = util.GetJSONInt64(data,"eip158Block",&out.Eip158Block)
	if err != nil {
		return nil,err
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
	
	return out,nil
}

func GetParams() string {
	dat, err := ioutil.ReadFile("./resources/geth/params.json")
    if err != nil {
        panic(err)//Missing required files is a fatal error
    }
    return string(dat)
}

func GetDefaults() string {
	dat, err := ioutil.ReadFile("./resources/geth/defaults.json")
    if err != nil {
        panic(err)//Missing required files is a fatal error
    }
    return string(dat)
}

func GetServices() []util.Service{
	return nil
}