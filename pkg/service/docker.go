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
	"github.com/docker/docker/client"
	"github.com/whiteblock/genesis/pkg/entity"
)

type DockerService interface {
	CreateContainer(cli *client.Client, container entity.Container) entity.Result
	StartContainer(cli *client.Client, name string) entity.Result
	RemoveContainer(cli *client.Client, name string) entity.Result
	CreateNetwork(cli *client.Client, net entity.Network) entity.Result
	AttachNetwork(cli *client.Client, network string, container string) entity.Result
	CreateVolume(cli *client.Client, volume entity.Volume) entity.Result
	RemoveVolume(cli *client.Client, name string)
	PlaceFileInContainer(cli *client.Client, containerName string, file entity.File) entity.Result
	PlaceFileInVolume(cli *client.Client, volumeName string, file entity.File) entity.Result
	Emulation(cli *client.Client, netem entity.Netconf) entity.Result
}

type dockerService struct {
}

func NewDockerService() (DockerService, error) {
	return dockerService{}, nil
}
