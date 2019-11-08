/*
	Copyright 2019 whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Genesis is distributed in the hope that it will be useful,
	but dock ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package entity

import (
	"github.com/docker/docker/api/types/network"
)

// NetworkConfig represents a docker network configuration
type NetworkConfig struct {
	//EndpointsConfig TODO: this will be removed
	EndpointsConfig map[string]*network.EndpointSettings
}

// Container represents a docker container, this is calculated from the payload of the Run command
type Container struct {
	// BoundCpus are the cpus which the container will be set with an affinity for.
	BoundCPUs []int `json:"boundCPUs,omitonempty"`
	// Detach indicates that we should wait for the containers entrypoint to finish execution
	Detach bool `json;"detach"`
	// EntryPoint overrides the docker containers entrypoint if non-empty
	EntryPoint string `json:"entrypoint"`
	// Environment represents the environment kv which will be provided to the container
	Environment map[string]string `json:"environment"`

	// Labels are any identifier which are to be attached to the container
	Labels map[string]string `json:"labels"`
	//Name is the unique name of the docker container
	Name string `json:"name"`
	//Network is the primary network for this container to be attached to
	Network string `json:"network"`
	//NetworkConfig: TODO remove from this struct
	NetworkConfig NetworkConfig `json:"NetworkConfig"`

	// Ports to be opened for each container, each port associated.
	Ports map[int]int `json:"ports"`

	// Volumes are the docker volumes to be mounted on this container
	Volumes map[string]Volume `json:"volumes"`

	Resources
	//Image is the docker image
	Image string `json:"image"`
	//Args are the arguments passed to the containers entrypoint
	Args []string `json:"args"`
}
