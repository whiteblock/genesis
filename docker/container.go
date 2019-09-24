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

package docker

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/util"
)

//ContainerType represents the type of sideCar the container is
type ContainerType int

const (
	// Node is a standard sideCar in the network
	Node ContainerType = 0

	// SideCar is a sidecar for a sideCar in the network
	SideCar ContainerType = 1

	// Service is a service sideCar
	Service ContainerType = 2
)

// Container represents the basic functionality needed to build a container
type Container interface {
	// AddVolume adds a volume to the container
	AddVolume(vol string)

	// GetEnvironment gives the environment variables for the container
	GetEnvironment() map[string]string

	// GetImage gives the image from which the container will be built
	GetImage() string

	// GetIP gives the IP address for the container
	GetIP() (string, error)

	// GetName gets the name of the container
	GetName() string

	// GetNetworkName gets the name of the containers network
	GetNetworkName() string

	// GetPorts gets the ports to open for the sideCar, if instructed.
	GetPorts() []string

	// GetResources gets the maximum resource allocation of the sideCar
	GetResources() util.Resources

	// GetEntryPoint gets the entrypoint for the container
	GetEntryPoint() string

	// GetArgs gets the arguments for the entrypoint
	GetArgs() []string

	// GetVolumes gets the container volumes
	GetVolumes() []string

	// SetEntryPoint sets the entrypoint
	SetEntryPoint(ep string)

	// SetArgs sets the arguments for the entrypoint
	SetArgs(args []string)
}

// ContainerDetails represents a docker containers details
type ContainerDetails struct {
	Environment  map[string]string
	Image        string
	Node         int
	Resources    util.Resources
	Labels		 ContainerLabels
	SubnetID     int
	NetworkIndex int
	Type         ContainerType
	EntryPoint   string
	Args         []string
}

type ContainerLabels struct {
	TestNetID string
	OrgID     string
}

// NewNodeContainer creates a representation of a container for a regular sideCar or regular node
func NewNodeContainer(node *db.Node, env map[string]string, resources util.Resources, SubnetID int) Container {
	return &ContainerDetails{
		Environment:  env,
		Image:        node.Image,
		Node:         node.LocalID,
		Resources:    resources,
		Labels: 	  ContainerLabels{
			TestNetID: node.TestNetID,
			OrgID: "", // TODO
		},
		SubnetID:     SubnetID,
		NetworkIndex: 0,
		Type:         Node,
		EntryPoint:   "/bin/sh",
		Args:         []string{},
	}
}

// NewSideCarContainer creates a representation of a container for a side car sideCar
func NewSideCarContainer(sc *db.SideCar, env map[string]string, resources util.Resources, SubnetID int) Container {
	return &ContainerDetails{
		Environment:  env,
		Image:        sc.Image,
		Node:         sc.LocalID,
		Resources:    resources,
		SubnetID:     SubnetID,
		NetworkIndex: sc.NetworkIndex,
		Type:         SideCar,
		EntryPoint:   "/bin/sh",
		Args:         []string{},
	}
}

// GetEnvironment gives the environment variables for the container
func (cd *ContainerDetails) GetEnvironment() map[string]string {
	return cd.Environment
}

// GetImage gives the image from which the container will be built
func (cd *ContainerDetails) GetImage() string {
	return cd.Image
}

// GetIP gives the IP address for the container
func (cd *ContainerDetails) GetIP() (string, error) {
	switch cd.Type {
	case Node:
		return util.GetNodeIP(cd.SubnetID, cd.Node, 0)
	case SideCar:
		return util.GetNodeIP(cd.SubnetID, cd.Node, cd.NetworkIndex)
	}
	log.Panic("Unsupported type")
	return "", nil
}

// GetName gets the name of the container
func (cd *ContainerDetails) GetName() string {
	switch cd.Type {
	case Node:
		return fmt.Sprintf("%s%d", conf.NodePrefix, cd.Node)
	case SideCar:
		return fmt.Sprintf("%s%d-%d", conf.NodePrefix, cd.Node, cd.NetworkIndex)
	}
	log.Panic("Unsupported type")
	return ""
}

// GetPorts gets the ports to open for the sideCar, if instructed.
func (cd *ContainerDetails) GetPorts() []string {
	return cd.Resources.Ports
}

// GetNetworkName gets the name of the containers network
func (cd *ContainerDetails) GetNetworkName() string {
	return fmt.Sprintf("%s%d", conf.NodeNetworkPrefix, cd.Node)
}

// GetResources gets the maximum resource allocation of the sideCar
func (cd *ContainerDetails) GetResources() util.Resources {
	return cd.Resources
}

func (cd *ContainerDetails) GetTestnetID() string {
	return cd.Labels.TestNetID
}

func (cd *ContainerDetails) GetOrgID() string {
	return cd.Labels.OrgID
}

// GetEntryPoint gets the entrypoint for the container
func (cd *ContainerDetails) GetEntryPoint() string {
	return cd.EntryPoint
}

// GetArgs gets the arguments for the entrypoint
func (cd *ContainerDetails) GetArgs() []string {
	return cd.Args
}

// SetEntryPoint sets the entrypoint
func (cd *ContainerDetails) SetEntryPoint(ep string) {
	cd.EntryPoint = ep
}

// SetArgs sets the arguments for the entrypoint
func (cd *ContainerDetails) SetArgs(args []string) {
	cd.Args = args
}

// GetVolumes gets the container's volumes
func (cd *ContainerDetails) GetVolumes() []string {
	return cd.Resources.Volumes
}

// AddVolume adds a volume to the container
func (cd *ContainerDetails) AddVolume(vol string) {
	cd.Resources.Volumes = append(cd.Resources.Volumes, vol)
}
