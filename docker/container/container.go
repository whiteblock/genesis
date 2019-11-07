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

package container

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/whiteblock/genesis/pkg/entity"
)

// CreateContainer creates a new container in the docker client
func CreateContainer(ctx context.Context, cli *client.Client, c entity.Container) entity.Result { // todo probably change to return RESULT in the future

	config := new(dockerContainer.Config)

	var envVars []string
	for key, val := range c.Environment {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, val))
	}
	config.Env = envVars

	config.Image = c.Image
	//config.Volumes =
	config.Entrypoint = []string{c.EntryPoint}
	config.Labels = c.Labels

	hostConfig := new(dockerContainer.HostConfig) //todo there should be a better way to create a hostconfig (a method or somethin)
	hostConfig.CpusetCpus = c.Resources.Cpus

	mem, err := c.Resources.GetMemory()
	if err != nil {
		return entity.Result{
			Error: err,
			Type:  entity.FatalType, //todo not sure what to put here yet
		}
	}
	hostConfig.Memory = mem

	networkConfig := new(network.NetworkingConfig)
	networkConfig.EndpointsConfig = c.NetworkConfig.EndpointsConfig

	_, err = cli.ContainerCreate(ctx, config, hostConfig, networkConfig, c.Name)
	if err != nil {
		return entity.Result{
			Error: err,
			Type:  entity.FatalType, //todo not sure what to put here yet
		}
	}

	return entity.Result{
		Error: nil,
		Type:  entity.SuccessType,
	}
}

// StartContainer starts a docker container
func StartContainer(ctx context.Context, cli *client.Client, name string) entity.Result { // todo probably change to return RESULT in the future
	opts := types.ContainerStartOptions{} //todo do we wanna do anything for this?

	err := cli.ContainerStart(ctx, name, opts)
	if err != nil {
		return entity.Result{
			Error: err,
			Type:  entity.FatalType, //todo not sure what to put here yet
		}
	}

	return entity.Result{
		Error: nil,
		Type:  entity.FatalType, //todo not sure what to put here yet
	}
}
