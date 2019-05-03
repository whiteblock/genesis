package docker

import (
	"../db"
	"../util"
	"fmt"
)

type ContainerType int

const (
	Node    ContainerType = 0
	SideCar ContainerType = 1
	Service ContainerType = 2
)

type Container interface {
	GetEnvironment() map[string]string
	GetImage() string
	GetIP() (string, error)
	GetName() string
	GetNetworkName() string
	GetResources() util.Resources
}

type ContainerDetails struct {
	Environment  map[string]string
	Image        string
	Node         int
	Resources    util.Resources
	SubnetID     int
	NetworkIndex int
	Type         ContainerType
}

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

func (cd *ContainerDetails) GetEnvironment() map[string]string {
	return cd.Environment
}

func (cd *ContainerDetails) GetImage() string {
	return cd.Image
}

func (cd *ContainerDetails) GetIP() (string, error) {
	switch cd.Type {
	case Node:
		return util.GetNodeIP(cd.SubnetID, cd.Node, 0)
	case SideCar:
		return util.GetNodeIP(cd.SubnetID, cd.Node, cd.NetworkIndex)
	}
	panic("Unsupported type")
}

func (cd *ContainerDetails) GetName() string {
	switch cd.Type {
	case Node:
		return fmt.Sprintf("%s%d", conf.NodePrefix, cd.Node)
	case SideCar:
		return fmt.Sprintf("%s%d-%d", conf.NodePrefix, cd.Node, cd.NetworkIndex)
	}
	panic("Unsupported type")
}

func (cd *ContainerDetails) GetNetworkName() string {
	return fmt.Sprintf("%s%d", conf.NodeNetworkPrefix, cd.Node)
}

func (cd *ContainerDetails) GetResources() util.Resources {
	return cd.Resources
}
