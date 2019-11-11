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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	repository "github.com/whiteblock/genesis/mocks/pkg/repository"
	"github.com/whiteblock/genesis/pkg/entity"

	//"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	//"github.com/docker/docker/api/types/network"
)


func TestDockerService_CreateContainer(t *testing.T) {
	repo := new(repository.DockerRepository)
	repo.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		container.ContainerCreateCreatedBody{}, nil).Run(func(args mock.Arguments){
			assert.Equal(t,len(args),6)
			assert.Nil(t,args.Get(1))
			//network.NetworkingConfig
		})
	
	ds,err := NewDockerService(repo)
	assert.NoError(t,err)
	res := ds.CreateContainer(
		nil, 
		nil,
		entity.Container{
			BoundCPUs: nil,
			Detach: false,
			EntryPoint:"/bin/bash",
			Environment: map[string]string{
				"FOO":"BAR",
			},
			Labels:  map[string]string{
				"FOO":"BAR",
			},
			Name: "TEST",
			Network : "Testnet",
			Ports:map[int]int{8888:8889},
			Volumes: map[string]entity.Volume{},
			Image:"alpine",
			Args:[]string{"Test"},
		})
	assert.NoError(t,res.Error)
	//ContainerCreate(ctx, cli, config, hostConfig, networkConfig, dContainer.Name)
}
	//CreateContainer(ctx context.Context, cli *client.Client, container entity.Container) entity.Result
