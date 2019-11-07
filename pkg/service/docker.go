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
	"github.com/docker/docker/client"
	"github.com/whiteblock/genesis/docker/container"
	"github.com/whiteblock/genesis/pkg/entity"
)

type DockerService interface {
	CreateClient(conf entity.DockerConfig, host string) (*client.Client, error)

	CreateContainer(ctx context.Context, cli *client.Client, container entity.Container) entity.Result
	StartContainer(ctx context.Context, cli *client.Client, name string) entity.Result
	RemoveContainer(ctx context.Context, cli *client.Client, name string) entity.Result
	CreateNetwork(ctx context.Context, cli *client.Client, net entity.Network) entity.Result
	AttachNetwork(ctx context.Context, cli *client.Client, network string, container string) entity.Result
	CreateVolume(ctx context.Context, cli *client.Client, volume entity.Volume) entity.Result
	RemoveVolume(ctx context.Context, cli *client.Client, name string) entity.Result
	PlaceFileInContainer(ctx context.Context, cli *client.Client, containerName string, file entity.File) entity.Result
	PlaceFileInVolume(ctx context.Context, cli *client.Client, volumeName string, file entity.File) entity.Result
	Emulation(ctx context.Context, cli *client.Client, netem entity.Netconf) entity.Result
}

type dockerService struct {
}

func NewDockerService() (DockerService, error) {
	return dockerService{}, nil
}

func (ds dockerService) CreateClient(conf entity.DockerConfig, host string) (*client.Client, error) {
	return client.NewClientWithOpts(
		client.WithAPIVersionNegotiation(),
		client.WithHost(host),
		client.WithTLSClientConfig(conf.CACertPath, conf.CertPath, conf.KeyPath),
	)
}

func (ds dockerService) CreateContainer(ctx context.Context, cli *client.Client, c entity.Container) entity.Result {
	//TODO
	return container.CreateContainer(ctx, cli, c)
}

func (ds dockerService) StartContainer(ctx context.Context, cli *client.Client, name string) entity.Result {
	//TODO
	return container.StartContainer(ctx, cli, name)
}

func (ds dockerService) RemoveContainer(ctx context.Context, cli *client.Client, name string) entity.Result {
	//TODO
	return entity.Result{}
}

func (ds dockerService) CreateNetwork(ctx context.Context, cli *client.Client, net entity.Network) entity.Result {
	//TODO
	return entity.Result{}
}

func (ds dockerService) AttachNetwork(ctx context.Context, cli *client.Client, network string, containerName string) entity.Result {
	//TODO
	return entity.Result{}
}

func (ds dockerService) CreateVolume(ctx context.Context, cli *client.Client, volume entity.Volume) entity.Result {
	//TODO
	return entity.Result{}
}

func (ds dockerService) RemoveVolume(ctx context.Context, cli *client.Client, name string) entity.Result {
	//TODO
	return entity.Result{}
}

func (ds dockerService) PlaceFileInContainer(ctx context.Context, cli *client.Client, containerName string, file entity.File) entity.Result {
	//TODO
	return entity.Result{}
}

func (ds dockerService) PlaceFileInVolume(ctx context.Context, cli *client.Client, volumeName string, file entity.File) entity.Result {
	//TODO
	return entity.Result{}
}

func (ds dockerService) Emulation(ctx context.Context, cli *client.Client, netem entity.Netconf) entity.Result {
	//TODO
	return entity.Result{}
}
