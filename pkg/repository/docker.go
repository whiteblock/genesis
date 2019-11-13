//+build !test

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

package repository

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"io"
)

//DockerRepository represents direct interacts with the docker daemon
type DockerRepository interface {
	//ContainerCreate creates a new container based in the given configuration. It can be associated with a name, but it's not mandatory.
	ContainerCreate(ctx context.Context, cli *client.Client, config *container.Config, hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error)

	//ContainerList returns the list of containers in the docker host.
	ContainerList(ctx context.Context, cli *client.Client, options types.ContainerListOptions) ([]types.Container, error)

	//ContainerRemove kills and removes a container from the docker host.
	ContainerRemove(ctx context.Context, cli *client.Client, containerID string, options types.ContainerRemoveOptions) error

	//ContainerStart sends a request to the docker daemon to start a container.
	ContainerStart(ctx context.Context, cli *client.Client, containerID string, options types.ContainerStartOptions) error

	//ImageList returns a list of images in the docker host
	ImageList(ctx context.Context, cli *client.Client, options types.ImageListOptions) ([]types.ImageSummary, error)

	//ImageLoad is used to upload a docker image
	ImageLoad(ctx context.Context, cli *client.Client, input io.Reader, quiet bool) (types.ImageLoadResponse, error)

	//ImagePull is used to pull a docker image
	ImagePull(ctx context.Context, cli *client.Client, refStr string, options types.ImagePullOptions) (io.ReadCloser, error)

	//NetworkCreate sends a request to the docker daemon to create a network
	NetworkCreate(ctx context.Context, cli *client.Client, name string, options types.NetworkCreate) (types.NetworkCreateResponse, error)

	//NetworkRemove sends a request to the docker daemon to remove a network
	NetworkRemove(ctx context.Context, cli *client.Client, networkID string) error

	//NetworkList lists the networks known to the docker daemon
	NetworkList(ctx context.Context, cli *client.Client, options types.NetworkListOptions) ([]types.NetworkResource, error)

	//NetworkConnect connects a container to a network
	NetworkConnect(ctx context.Context, cli *client.Client, networkID, containerID string, config *network.EndpointSettings) error
}

type dockerRepository struct {
}

//NewDockerRepository creates a new DockerRepository
func NewDockerRepository() DockerRepository {
	return &dockerRepository{}
}

//ContainerCreate creates a new container based in the given configuration. It can be associated with a name, but it's not mandatory.
func (dr dockerRepository) ContainerCreate(ctx context.Context, cli *client.Client, config *container.Config,
	hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig,
	containerName string) (container.ContainerCreateCreatedBody, error) {
	return cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, containerName)
}

//ContainerList returns the list of containers in the docker host.
func (dr dockerRepository) ContainerList(ctx context.Context, cli *client.Client,
	options types.ContainerListOptions) ([]types.Container, error) {

	return cli.ContainerList(ctx, options)
}

//ContainerRemove kills and removes a container from the docker host.
func (dr dockerRepository) ContainerRemove(ctx context.Context, cli *client.Client,
	containerID string, options types.ContainerRemoveOptions) error {

	return cli.ContainerRemove(ctx, containerID, options)
}

//ContainerStart sends a request to the docker daemon to start a container.
func (dr dockerRepository) ContainerStart(ctx context.Context, cli *client.Client,
	containerID string, options types.ContainerStartOptions) error {
	return cli.ContainerStart(ctx, containerID, options)
}

//ImageLoad is used to upload a docker image
func (dr dockerRepository) ImageLoad(ctx context.Context, cli *client.Client,
	input io.Reader, quiet bool) (types.ImageLoadResponse, error) {

	return cli.ImageLoad(ctx, input, quiet)
}

//ImagePull is used to pull a docker image
func (dr dockerRepository) ImagePull(ctx context.Context, cli *client.Client,
	refStr string, options types.ImagePullOptions) (io.ReadCloser, error) {

	return cli.ImagePull(ctx, refStr, options)
}

//ImageList returns a list of images in the docker host
func (dr dockerRepository) ImageList(ctx context.Context, cli *client.Client,
	options types.ImageListOptions) ([]types.ImageSummary, error) {

	return cli.ImageList(ctx, options)
}

//NetworkCreate sends a request to the docker daemon to create a network
func (dr dockerRepository) NetworkCreate(ctx context.Context, cli *client.Client, name string,
	options types.NetworkCreate) (types.NetworkCreateResponse, error) {
	return cli.NetworkCreate(ctx, name, options)
}

//NetworkRemove sends a request to the docker daemon to remove a network
func (dr dockerRepository) NetworkRemove(ctx context.Context, cli *client.Client, networkID string) error {
	return cli.NetworkRemove(ctx, networkID)
}

//NetworkList lists the networks known to the docker daemon
func (dr dockerRepository) NetworkList(ctx context.Context, cli *client.Client,
	options types.NetworkListOptions) ([]types.NetworkResource, error) {
	return cli.NetworkList(ctx, options)
}

//NetworkConnect connects a container to a network
func (dr dockerRepository) NetworkConnect(ctx context.Context, cli *client.Client, networkID,
	containerID string, config *network.EndpointSettings) error {

	return cli.NetworkConnect(ctx, networkID, containerID, config)
}
