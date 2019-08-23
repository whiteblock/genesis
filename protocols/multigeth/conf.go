/*
	Copyright 2019 whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Genesis is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package multigeth

import (
	"encoding/json"
	"fmt"

	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/testnet"
)

type ethConf struct {
	NetworkID          int64  `json:"networkId"`
	ChainID            int64  `json:"chainId"`
	HomesteadBlock     int64  `json:"homesteadBlock"`
	Difficulty         int64  `json:"difficulty"`
	InitBalance        string `json:"initBalance"`
	MaxPeers           int64  `json:"maxPeers"`
	GasLimit           int64  `json:"gasLimit"`
	Consensus          string `json:"consensus"`
	BlockPeriodSeconds int64  `json:"blockPeriodSeconds"`
	Epoch              int64  `json:"epoch"`
	EIP150Block        int64  `json:"eip150Block"`
	EIP155Block        int64  `json:"eip155Block"`
	EIP158Block        int64  `json:"eip158Block"`
	ByzantiumBlock     int64  `json:"byzantiumBlock"`
	DisposalBlock      int64  `json:"disposalBlock"`
	//	ConstantinopleBlock int64  `json:"constantinopleBlock"`
	ECIP1017EraRounds  int64 `json:"ecip1017EraRounds"`
	EIP160FBlock       int64 `json:"eip160FBlock"`
	ECIP1010PauseBlock int64 `json:"ecip1010PauseBlock"`
	ECIP1010Length     int64 `json:"ecip1010Length"`
	//	TrustedCheckpoint   int64  `json:"trustedCheckpoint"`
	ExposedAccounts int64  `json:"exposedAccounts"`
	ExtraAccounts   int64  `json:"extraAccounts"`
	MixHash         string `json:"mixHash"`
	Verbosity       int    `json:"verbosity"`
	Nonce           string `json:"nonce"`
	Timestamp       int64  `json:"timestamp"`
}

/**
 * Fills in the defaults for missing parts,
 */
func newConf(tn *testnet.TestNet) (*ethConf, error) {
	data := tn.LDD.Params
	out := new(ethConf)
	err := helpers.HandleBlockchainConfig(blockchain, data, out)
	if err != nil || data == nil {
		return out, err
	}

	initBalance, exists := data["initBalance"]
	if exists && initBalance != nil {
		switch initBalance.(type) {
		case json.Number:
			out.InitBalance = initBalance.(json.Number).String()
		case string:
			out.InitBalance = initBalance.(string)
		default:
			return nil, fmt.Errorf("incorrect type for initBalance given")
		}
	}
	if out.ExposedAccounts != -1 && out.ExposedAccounts > out.ExtraAccounts+int64(tn.LDD.Nodes) {
		out.ExtraAccounts = out.ExposedAccounts - int64(tn.LDD.Nodes)
	}

	return out, nil
}

func newConfFromStore(tn *testnet.TestNet) (*ethConf, error) {
	out, err := newConf(tn)
	if err != nil {
		return nil, util.LogError()
	}
	tn.BuildState.GetP("etcconf", out)
	return out, nil
}
