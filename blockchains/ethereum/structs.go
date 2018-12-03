package eth

import (
	"encoding/json"
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

	chainId,exists := data["chainId"]
	if exists {
		out.ChainId,err = chainId.(json.Number).Int64()
		if err != nil {
			return nil,err
		}
	}

	networkId,exists := data["networkId"]
	if exists {
		out.NetworkId,err = networkId.(json.Number).Int64()
		if err != nil {
			return nil,err
		}
	}

	difficulty,exists := data["difficulty"]
	if exists {
		out.Difficulty,err = difficulty.(json.Number).Int64()
		if err != nil {
			return nil,err
		}
	}

	initBalance,exists := data["initBalance"]
	if exists {
		out.InitBalance = initBalance.(json.Number).String()
	}

	maxPeers,exists := data["maxPeers"]
	if exists {
		out.MaxPeers,err = maxPeers.(json.Number).Int64()
		if err != nil {
			return nil,err
		}
	}

	gasLimit,exists := data["gasLimit"]
	if exists {
		out.GasLimit,err = gasLimit.(json.Number).Int64()
		if err != nil {
			return nil,err
		}
	}

	homesteadBlock,exists := data["homesteadBlock"]
	if exists {
		out.HomesteadBlock,err = homesteadBlock.(json.Number).Int64()
		if err != nil {
			return nil,err
		}
	}

	eip155Block,exists := data["eip155Block"]
	if exists {
		out.Eip155Block,err = eip155Block.(json.Number).Int64()
		if err != nil {
			return nil,err
		}
	}

	eip158Block,exists := data["eip158Block"]
	if exists {
		out.Eip158Block,err = eip158Block.(json.Number).Int64()
		if err != nil {
			return nil,err
		}
	}

	return out,nil
}