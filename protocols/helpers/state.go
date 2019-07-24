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

package helpers

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"strings"
)

// Helper functions to hint at common things that one may want to set

const (
	alternativeCommandRegexKey = "__alternative_commands"
	functionalityGroupKey      = "namespace"
	protocolGroupKey           = "__protocol"
)

// SetAlternativeCmdExprs allows you to have your protocol support restart and related
// functionality if the blockchain main process looks diferent from the actual process.
func SetAlternativeCmdExprs(tn *testnet.TestNet, alts ...string) {
	tn.BuildState.Set(alternativeCommandRegexKey, alts)
}

// GetCommandExprs get the command expressions to match on to find the main
// blockchain process
func GetCommandExprs(tn *testnet.TestNet, node string) ([]string, error) {
	var cmd util.Command
	ok := tn.BuildState.GetP(node, &cmd)
	if !ok {
		log.WithFields(log.Fields{"node": node}).Error("node not found")
		return nil, fmt.Errorf("node not found")
	}
	out := []string{strings.Split(cmd.Cmdline, " ")[0]}
	var alts []string
	tn.BuildState.GetP(alternativeCommandRegexKey, &alts)
	return append(out, alts...), nil
}

//SetFunctionalityGroup allows you to mark your protocol
//as being part of a functionality group. Most common group right now
//is eth
func SetFunctionalityGroup(tn *testnet.TestNet, name string) {
	tn.BuildState.SetExt(functionalityGroupKey, name)
}

// SetProtocolGroup sets the protocol group for the testnet
func SetProtocolGroup(tn *testnet.TestNet, name string) {
	tn.BuildState.Set(protocolGroupKey, name)
}

// GetProtocolGroup gets the protocol group for the testnet
func GetProtocolGroup(tn *testnet.TestNet) (string, error) {
	var out string
	exists := tn.BuildState.GetP(functionalityGroupKey, &out)
	if !exists {
		return "", fmt.Errorf("protocol group is not set")
	}
	return out, nil
}
