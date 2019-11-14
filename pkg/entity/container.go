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
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	"strconv"
	"strings"
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
	Detach bool `json:"detach"`
	// EntryPoint overrides the docker containers entrypoint if non-empty
	EntryPoint string `json:"entrypoint"`
	// Environment represents the environment kv which will be provided to the container
	Environment map[string]string `json:"environment"`

	// Labels are any identifier which are to be attached to the container
	Labels map[string]string `json:"labels"`
	//Name is the unique name of the docker container
	Name string `json:"name"`
	//Network is the primary network(s) for this container to be attached to
	Network strslice.StrSlice `json:"network"`

	// Ports to be opened for each container, each port associated.
	Ports map[int]int `json:"ports"`

	// Volumes are the docker volumes to be mounted on this container
	Volumes []Mount `json:"volumes"`

	// Cpus should be a floating point value represented as a string, and
	// is  equivalent to the percentage of a single cores time which can be used
	// by a node. Can be more than 1.0, meaning the node can use multiple cores at
	// a time.
	Cpus string `json:"cpus"`

	// Memory supports values up to Terrabytes (tb). If the unit is omitted, then it
	// is assumed to be bytes. This is not case sensitive.
	Memory string `json:"memory"`

	//Image is the docker image
	Image string `json:"image"`
	//Args are the arguments passed to the containers entrypoint
	Args []string `json:"args"`
}

// GetMemory gets the memory value as an integer.
func (c Container) GetMemory() (int64, error) {
	return memconv(c.Memory)
}

// NoLimits checks if the resources object doesn't specify any limits
func (c Container) NoLimits() bool {
	return c.NoCPULimits() && c.NoMemoryLimits()
}

// NoCPULimits checks if the resources object doesn't specify any cpu limits
func (c Container) NoCPULimits() bool {
	return len(c.Cpus) == 0
}

// NoMemoryLimits checks if the resources object doesn't specify any memory limits
func (c Container) NoMemoryLimits() bool {
	return len(c.Memory) == 0
}

//GetEnv gets the environment variables in the format which is
//expected by docker
func (c Container) GetEnv() (envVars []string) {
	for key, val := range c.Environment {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, val))
	}
	return envVars
}

//GetPortBindings gets the port bindings in the proper docker container types
func (c Container) GetPortBindings() (nat.PortSet, nat.PortMap, error) {
	if c.Ports == nil {
		return nil, nil, nil
	}
	dockerPorts := []string{}
	for hostPort, containerPort := range c.Ports {
		dockerPorts = append(dockerPorts, fmt.Sprintf("0.0.0.0:%d:%d/tcp", hostPort, containerPort))
	}
	portSet, portMap, err := nat.ParsePortSpecs(dockerPorts)
	return nat.PortSet(portSet), nat.PortMap(portMap), err
}

//GetEntryPoint returns the properly formatted entrypiont if this container has one,
//otherwise returns nil
func (c Container) GetEntryPoint() strslice.StrSlice {
	if len(c.EntryPoint) == 0 {
		return nil
	}
	return strslice.StrSlice(append([]string{c.EntryPoint}, c.Args...))
}

//GetMounts gets the docker mounts for the docker container
func (c Container) GetMounts() []mount.Mount {
	out := []mount.Mount{}
	for _, vol := range c.Volumes {
		out = append(out, mount.Mount{
			Type:     mount.TypeVolume,
			Source:   vol.Name,
			Target:   vol.Directory,
			ReadOnly: vol.ReadOnly,
		})
	}
	return out
}

func memconv(mem string) (int64, error) {

	m := strings.ToLower(mem)

	var multiplier int64 = 1

	if strings.HasSuffix(m, "kb") || strings.HasSuffix(m, "k") {
		multiplier = 1000
	} else if strings.HasSuffix(m, "mb") || strings.HasSuffix(m, "m") {
		multiplier = 1000000
	} else if strings.HasSuffix(m, "gb") || strings.HasSuffix(m, "g") {
		multiplier = 1000000000
	} else if strings.HasSuffix(m, "tb") || strings.HasSuffix(m, "t") {
		multiplier = 1000000000000
	}

	i, err := strconv.ParseInt(strings.Trim(m, "bgkmt"), 10, 64)
	if err != nil {
		return -1, err
	}

	return i * multiplier, nil
}
