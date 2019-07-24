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

package ethereum

import (
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/testnet"
)

const (
	// FuncGroup is the functionality group of most ethereum nodes
	FuncGroup = "eth"
	// ProtocolGroup is the protocol group value for ethereum
	ProtocolGroup = "eth"
	// ClassicProtocolGroup is the protocol group value for ethereum classic
	ClassicProtocolGroup = "etc"
	// Eth2ProtocolGroup is the protocol group value for ethereum 2
	Eth2ProtocolGroup = "et2"
)

// IsEthereum checks if the testnet has been flagged as being an ethereum network
func IsEthereum(tn *testnet.TestNet) bool {
	group, err := helpers.GetProtocolGroup(tn)
	if err != nil {
		log.WithFields(log.Fields{"msg": err}).Warn("no protocol group found")
		return false
	}
	return group == ProtocolGroup
}

// IsEthereumClassic checks if the testnet has been flagged as being an ethereum classic network
func IsEthereumClassic(tn *testnet.TestNet) bool {
	group, err := helpers.GetProtocolGroup(tn)
	if err != nil {
		log.WithFields(log.Fields{"msg": err}).Warn("no protocol group found")
		return false
	}
	return group == ClassicProtocolGroup
}

// IsEthereum2 checks if the testnet has been flagged as being an eth2.0 network
func IsEthereum2(tn *testnet.TestNet) bool {
	group, err := helpers.GetProtocolGroup(tn)
	if err != nil {
		log.WithFields(log.Fields{"msg": err}).Warn("no protocol group found")
		return false
	}
	return group == Eth2ProtocolGroup
}

// MarkAsEthereum marks the given testnet as ethereum
func MarkAsEthereum(tn *testnet.TestNet) {
	helpers.SetProtocolGroup(tn, ProtocolGroup)
}

// MarkAsEthereumClassic marks the given testnet as ethereum classic
func MarkAsEthereumClassic(tn *testnet.TestNet) {
	helpers.SetProtocolGroup(tn, ClassicProtocolGroup)
}

// MarkAsEthereum2 marks the given testnet as ethereum 2.0
func MarkAsEthereum2(tn *testnet.TestNet) {
	helpers.SetProtocolGroup(tn, Eth2ProtocolGroup)
}
