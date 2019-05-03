package docker

import (
	"../db"
	"../util"
	"fmt"
)

//ContainerType represents the type of node the container is
type ContainerType int

const (
	// Node is a standard node in the network
	Node ContainerType = 0

	// SideCar is a sidecar for a node in the network
	SideCar ContainerType = 1

	// Service is a service node
	Service ContainerType = 2
)

// Container represents the basic functionality needed to build a container
type Container interface {
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

	// GetResources gets the maximum resource allocation of the node
	GetResources() util.Resources
}

// ContainerDetails represents a docker containers details
type ContainerDetails struct {
	Environment  map[string]string
	Image        string
	Node         int
	Resources    util.Resources
	SubnetID     int
	NetworkIndex int
	Type         ContainerType
}

// NewNodeContainer creates a representation of a container for a regular node
func NewNodeContainer(node *db.Node, env map[string]string, resources util.Resources, SubnetID int) Container {
	return &ContainerDetails{
		Environment:  env,
		Image:        node.Image,
		Node:         node.LocalID,
		Resources:    resources,
		SubnetID:     SubnetID,
		NetworkIndex: 0,
		Type:         Node,
	}
}

// NewSideCarContainer creates a representation of a container for a side car node
func NewSideCarContainer(sc *db.SideCar, env map[string]string, resources util.Resources, SubnetID int) Container {
	return &ContainerDetails{
		Environment:  env,
		Image:        sc.Image,
		Node:         sc.LocalID,
		Resources:    resources,
		SubnetID:     SubnetID,
		NetworkIndex: sc.NetworkIndex,
		Type:         SideCar,
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
	panic("Unsupported type")
}

// GetName gets the name of the container
func (cd *ContainerDetails) GetName() string {
	switch cd.Type {
	case Node:
		return fmt.Sprintf("%s%d", conf.NodePrefix, cd.Node)
	case SideCar:
		return fmt.Sprintf("%s%d-%d", conf.NodePrefix, cd.Node, cd.NetworkIndex)
	}
	panic("Unsupported type")
}

// GetNetworkName gets the name of the containers network
func (cd *ContainerDetails) GetNetworkName() string {
	return fmt.Sprintf("%s%d", conf.NodeNetworkPrefix, cd.Node)
}

// GetResources gets the maximum resource allocation of the node
func (cd *ContainerDetails) GetResources() util.Resources {
	return cd.Resources
}
