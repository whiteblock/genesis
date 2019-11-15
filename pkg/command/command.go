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

package command

import (
	"time"
)

//OrderType is the type of order
type OrderType string

const (
	//Createcontainer attempts to create a docker container
	Createcontainer = OrderType("createcontainer")
	//Startcontainer attempts to start an already created docker container
	Startcontainer = OrderType("startcontainer")
	//Removecontainer attempts to remove a container
	Removecontainer = OrderType("removecontainer")
	//Createnetwork attempts to create a network
	Createnetwork = OrderType("createnetwork")
	//Attachnetwork attempts to remove a network
	Attachnetwork = OrderType("attachnetwork")
	//Detachnetwork detaches network
	Detachnetwork = OrderType("detachnetwork")
	//Removenetwork removes network
	Removenetwork = OrderType("removenetwork")
	//Createvolume creates volume
	Createvolume = OrderType("createvolume")
	//Removevolume removes volume
	Removevolume = OrderType("removevolume")
	//Putfile puts file
	Putfile = OrderType("putfile")
	//Putfileincontainer puts file in container
	Putfileincontainer = OrderType("putfileincontainer")
	//Emulation emulates
	Emulation = OrderType("emulation")
)

// Generic order payload
type OrderPayload interface {
}

// Order to be executed by genesis
type Order struct {
	//Type is the type of the order
	Type OrderType `json:"type"`
	//Payload is the payload object of the order
	Payload OrderPayload `json:"payload"`
}

// simple order payload with just the container name
type SimpleName struct {
	OrderPayload
	// Name of the container.
	Name string `json:"name"`
}

// Container and network order payload.
type ContainerNetwork struct {
	OrderPayload
	// Name of the container.
	ContainerName string `json:"container"`
	// Name of the network.
	Network string `json:"network"`
}

// File and container order payload.
type FileAndContainer struct {
	OrderPayload
	// Name of the container.
	ContainerName string `json:"container"`
	// Name of the file.
	File File `json:"file"`
}

// File and volume order payload.
type FileAndVolume struct {
	OrderPayload
	// Name of the volume.
	VolumeName string `json:"volume"`
	// Name of the file.
	File File `json:"file"`
}

// Target sets the target of a command - which testnet, instance to hit
type Target struct {
	IP string `json:"ip"`
}

// Command is the command sent to Genesis.
type Command struct {
	// ID is the unique id of this command
	ID string `json:"id"`
	// Timestamp is the earliest time the command can be executed
	Timestamp int64 `json:"timestamp"`
	// Timeout is the maximum amount of time a command can take before being aborted
	Timeout time.Duration `json:"timeout"`
	//Retry is the number of times this command has been retried
	Retry uint8 `json:"retry"`
	//Target represents the target of this command
	Target Target `json:"target"`
	//Dependencies is an array of ids of commands which must execute before this one
	Dependencies []string `json:"dependencies"`
	//Order is the action of the command, it represents what needs to be done
	Order Order `json:"order"`
}

//GetRetryCommand creates a copy of this command which has been modified to be requeued after an error
func (cmd Command) GetRetryCommand(newTimestamp int64) Command {
	cmd.Timestamp = newTimestamp
	cmd.Retry++
	return cmd
}
