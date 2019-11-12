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

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	repoMocks "github.com/whiteblock/genesis/mocks/pkg/repository"
	"github.com/whiteblock/genesis/pkg/entity"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

func TestDockerService_CreateContainer(t *testing.T) {
	testContainer := entity.Container{
		BoundCPUs:  nil, //TODO
		Detach:     false,
		EntryPoint: "/bin/bash", //TODO
		Environment: map[string]string{ //TODO
			"FOO": "BAR",
		},
		Labels: map[string]string{
			"FOO": "BAR",
		},
		Name:    "TEST",
		Network: "Testnet",                  //TODO
		Ports:   map[int]int{8888: 8889},    //TODO
		Volumes: map[string]entity.Volume{}, //TODO
		Image:   "alpine",
		Args:    []string{"Test"},
	}
	testContainer.Cpus = "2.5"
	testContainer.Memory = "5gb"

	repo := new(repoMocks.DockerRepository)
	repo.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		container.ContainerCreateCreatedBody{}, nil).Run(func(args mock.Arguments) {
		require.Len(t, args, 6)
		assert.Nil(t, args.Get(0))
		assert.Nil(t, args.Get(1))

		config, ok := args.Get(2).(*container.Config)
		require.True(t, ok)
		require.NotNil(t, config)
		require.Len(t, config.Entrypoint, 1)
		assert.Equal(t, testContainer.EntryPoint, config.Entrypoint[0])
		assert.Equal(t, testContainer.Name, config.Hostname)
		assert.Equal(t, testContainer.Labels, config.Labels)
		assert.Equal(t, testContainer.Image, config.Image)

		hostConfig, ok := args.Get(3).(*container.HostConfig)
		require.True(t, ok)
		require.NotNil(t, hostConfig)
		assert.Equal(t, int64(2500000000), hostConfig.NanoCPUs)
		assert.Equal(t, int64(5000000000), hostConfig.Memory)

		networkingConfig, ok := args.Get(4).(*network.NetworkingConfig)
		require.True(t, ok)
		require.NotNil(t, networkingConfig)

		containerName, ok := args.Get(5).(string)
		require.True(t, ok)
		assert.Equal(t, testContainer.Name, containerName)
	})

	ds, err := NewDockerService(repo)
	assert.NoError(t, err)

	res := ds.CreateContainer(nil, nil, testContainer)
	assert.NoError(t, res.Error)
}

