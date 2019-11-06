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
	"github.com/moby/moby/client"
	"github.com/whiteblock/genesis/pkg/entity"
)

type DockerService interface {
	CreateContainer(context context.Context, cli *client.Client, container entity.Container) entity.Result
	StartContainer(context context.Context, cli *client.Client, name string) entity.Result
	RemoveContainer(context context.Context, cli *client.Client, name string) entity.Result
	CreateNetwork(context context.Context, cli *client.Client, net entity.Network) entity.Result
	AttachNetwork(context context.Context, cli *client.Client, network string, container string) entity.Result
	CreateVolume(context context.Context, cli *client.Client, volume entity.Volume) entity.Result
	RemoveVolume(context context.Context, cli *client.Client, name string) entity.Result
	PlaceFileInContainer(context context.Context, cli *client.Client, containerName string, file entity.File) entity.Result
	PlaceFileInVolume(context context.Context, cli *client.Client, volumeName string, file entity.File) entity.Result
	Emulation(context context.Context, cli *client.Client, netem entity.Netconf) entity.Result
}

type dockerService struct {
}

func NewDockerService() (DockerService, error) {
	return dockerService{}, nil
}

func (ds dockerService) CreateContainer(cli *client.Client, container entity.Container) entity.Result {
	//TODO
	return entity.Result{}
}

func (ds dockerService) StartContainer(cli *client.Client, name string) entity.Result {
	//TODO
	return entity.Result{}
}

func (ds dockerService) RemoveContainer(cli *client.Client, name string) entity.Result {
	//TODO
	return entity.Result{}
}

func (ds dockerService) CreateNetwork(cli *client.Client, net entity.Network) entity.Result {
	//TODO
	return entity.Result{}
}

func (ds dockerService) AttachNetwork(cli *client.Client, network string, container string) entity.Result {
	//TODO
	return entity.Result{}
}

func (ds dockerService) CreateVolume(cli *client.Client, volume entity.Volume) entity.Result {
	//TODO
	return entity.Result{}
}

func (ds dockerService) RemoveVolume(cli *client.Client, name string) entity.Result {
	//TODO
	return entity.Result{}
}

func (ds dockerService) PlaceFileInContainer(cli *client.Client, containerName string, file entity.File) entity.Result {
	//TODO
	return entity.Result{}
}

func (ds dockerService) PlaceFileInVolume(cli *client.Client, volumeName string, file entity.File) entity.Result {
	//TODO
	return entity.Result{}
}

func (ds dockerService) Emulation(cli *client.Client, netem entity.Netconf) entity.Result {
	//TODO
	return entity.Result{}
}
