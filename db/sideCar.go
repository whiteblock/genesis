/*
	Copyright 2019 Whiteblock Inc.
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
)

// SideCar represents a supporting node within the network
type SideCar struct {
	ID string `json:"id"`

	NodeID string `json:"nodeID"`

	AbsoluteNodeNum int `json:"absNum"`

	// TestNetID is the id of the testnet to which the node belongs to
	TestnetID string `json:"testnetID"`

	// Server is the id of the server on which the node resides
	Server int `json:"server"`

	//LocalID is the number of the node in the testnet
	LocalID int `json:"localID"`

	NetworkIndex int `json:"networkIndex"`

	// IP is the ip address of the node
	IP string `json:"ip"`

	// Image is the docker image on which the sidecar was built
	Image string `json:"image"`

	// Type is the type of sidecar
	Type string `json:"type"`
}

// GetAbsoluteNumber gets the absolute node number of the corresponding ndoe
func (n SideCar) GetAbsoluteNumber() int {
	return n.AbsoluteNodeNum
}

// GetIP gets the ip address of this side car
func (n SideCar) GetIP() string {
	return n.IP
}

// GetRelativeNumber gets the local id of the corresponding node
func (n SideCar) GetRelativeNumber() int {
	return n.LocalID
}

// GetServerID gets the id of the server this side car resides on
func (n SideCar) GetServerID() int {
	return n.Server
}

// GetTestNetID gets the id of the testnet this side car is a part of
func (n SideCar) GetTestNetID() string {
	return n.TestnetID
}

// GetNodeName gets the whiteblock name of this side car
func (n SideCar) GetNodeName() string {
	return fmt.Sprintf("%s%d-%d", conf.NodePrefix, n.AbsoluteNodeNum, n.NetworkIndex)
}
