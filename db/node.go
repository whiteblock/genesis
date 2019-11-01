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

package db

import (
	"fmt"
	"github.com/whiteblock/genesis/util"
)

var conf = util.GetConfig()

// Node represents a node within the network
type Node struct {
	// ID is the UUID of the node
	ID string `json:"id"`

	//AbsoluteNum is the number of the node in the testnet
	AbsoluteNum int `json:"absNum"`

	// TestNetId is the id of the testnet to which the node belongs to
	TestNetID string `json:"testnetId"`

	// Server is the id of the server on which the node resides
	Server int `json:"server"`

	// LocalID is the number of the node on the server it resides
	LocalID int `json:"localId"`

	// IP is the ip address of the node
	IP string `json:"ip"`

	// Label is the string given to the node by the build process
	Label string `json:"label"`

	// Image is the docker image used to build this node
	Image string `json:"image"`

	// Protocol is the protocol type of this node
	Protocol string `json:"protocol"`

	// PortMappings keeps tracks of the ports exposed externally on the for this
	// node
	PortMappings map[string]string `json:"portMappings,omitonempty"`
}

// GetID gets the id of this side car
func (n Node) GetID() string {
	return n.ID
}

// GetAbsoluteNumber gets the absolute number of the node in the testnet
func (n Node) GetAbsoluteNumber() int {
	return n.AbsoluteNum
}

// GetIP gets the ip address of this node
func (n Node) GetIP() string {
	return n.IP
}

// GetRelativeNumber gets the local id of the node
func (n Node) GetRelativeNumber() int {
	return n.LocalID
}

// GetServerID gets the id of the server on which this node resides
func (n Node) GetServerID() int {
	return n.Server
}

// GetTestNetID gets the id of the testnet this node is a part of
func (n Node) GetTestNetID() string {
	return n.TestNetID
}

// GetNodeName gets the whiteblock name of this node
func (n Node) GetNodeName() string {
	return fmt.Sprintf("%s%d", conf.NodePrefix, n.AbsoluteNum)
}
