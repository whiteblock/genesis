/*
	Copyright 2019 whiteblock Incontainer.
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
	"strconv"

	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/whiteblock/genesis/pkg/entity"
)

// CreateContainer creates a new container in the docker client
func CreateContainer(ctx context.Context, cli *client.Client, container entity.Container) entity.Result { //TODO push to service

	var envVars []string
	for key, val := range container.Environment {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, val))
	}

	config := &dockerContainer.Config{
		Env:        envVars,
		Image:      container.Image,
		Entrypoint: []string{container.EntryPoint},
		Labels:     container.Labels,
	}

	mem, err := container.GetMemory()
	if err != nil {
		return entity.NewFatalResult(err)
	}

	cpus, err := strconv.ParseFloat(container.Cpus, 64)
	if err != nil {
		return entity.NewFatalResult(err)
	}

	hostConfig := &dockerContainer.HostConfig{}
	hostConfig.NanoCPUs = int64(1000000000 * cpus)
	hostConfig.Memory = mem

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: container.NetworkConfig.EndpointsConfig,
	}

	_, err = cli.ContainerCreate(ctx, config, hostConfig, networkConfig, container.Name) //Wrap/Put into repo
	if err != nil {
		return entity.NewFatalResult(err)
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
