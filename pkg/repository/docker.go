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

	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/whiteblock/genesis/pkg/entity"
)

//NetworkDisconnect(ctx context.Context, networkID, containerID string, force bool) error

//DockerRepository represents direct interacts with the docker daemon
type DockerRepository interface {
	//ContainerAttach attaches to a container
	ContainerAttach(ctx context.Context, cli entity.Client, container string,
		options types.ContainerAttachOptions) (types.HijackedResponse, error)

	//ContainerCreate creates a new container based in the given configuration. It can be associated with a name, but it's not mandatory.
	ContainerCreate(ctx context.Context, cli entity.Client, config *container.Config, hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error)

	//ContainerList returns the list of containers in the docker host.
	ContainerList(ctx context.Context, cli entity.Client, options types.ContainerListOptions) ([]types.Container, error)

	//ContainerRemove kills and removes a container from the docker host.
	ContainerRemove(ctx context.Context, cli entity.Client, containerID string, options types.ContainerRemoveOptions) error

	//ContainerStart sends a request to the docker daemon to start a container.
	ContainerStart(ctx context.Context, cli entity.Client, containerID string, options types.ContainerStartOptions) error

	//CopyToContainer copies content into the container filesystem. Note that `content` must be a Reader for a TAR archive
	CopyToContainer(ctx context.Context, cli entity.Client, containerID, dstPath string, content io.Reader,
		options types.CopyToContainerOptions) error

	//ImageList returns a list of images in the docker host
	ImageList(ctx context.Context, cli entity.Client, options types.ImageListOptions) ([]types.ImageSummary, error)

	//ImageLoad is used to upload a docker image
	ImageLoad(ctx context.Context, cli entity.Client, input io.Reader, quiet bool) (types.ImageLoadResponse, error)

	//ImagePull is used to pull a docker image
	ImagePull(ctx context.Context, cli entity.Client, refStr string, options types.ImagePullOptions) (io.ReadCloser, error)

	//NetworkCreate sends a request to the docker daemon to create a network
	NetworkCreate(ctx context.Context, cli entity.Client, name string, options types.NetworkCreate) (types.NetworkCreateResponse, error)

	//NetworkConnect connects a container to an existent network in the docker host.
	NetworkConnect(ctx context.Context, cli entity.Client, networkID, containerID string, config *network.EndpointSettings) error

	//NetworkDisconnect disconnects a container from an existent network in the docker host.
	NetworkDisconnect(ctx context.Context, cli entity.Client, networkID, containerID string, force bool) error

	//NetworkRemove sends a request to the docker daemon to remove a network
	NetworkRemove(ctx context.Context, cli entity.Client, networkID string) error

	//NetworkList lists the networks known to the docker daemon
	NetworkList(ctx context.Context, cli entity.Client, options types.NetworkListOptions) ([]types.NetworkResource, error)

	//VolumeList returns the volumes configured in the docker host.
	VolumeList(ctx context.Context, cli entity.Client, filter filters.Args) (volume.VolumeListOKBody, error)

	//VolumeRemove removes a volume from the docker host.
	VolumeRemove(ctx context.Context, cli entity.Client, volumeID string, force bool) error

	//VolumeCreate creates a volume in the container
	VolumeCreate(ctx context.Context, cli entity.Client, options volume.VolumeCreateBody) (types.Volume, error)
}

type dockerRepository struct {
}

//NewDockerRepository creates a new DockerRepository
func NewDockerRepository() DockerRepository {
	return &dockerRepository{}
}

//ContainerAttach attaches to a container
func (dr dockerRepository) ContainerAttach(ctx context.Context, cli entity.Client,
	container string, options types.ContainerAttachOptions) (types.HijackedResponse, error) {

	return cli.ContainerAttach(ctx, container, options)
}

//ContainerCreate creates a new container based in the given configuration. It can be associated with a name, but it's not mandatory.
func (dr dockerRepository) ContainerCreate(ctx context.Context, cli entity.Client, config *container.Config,
	hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig,
	containerName string) (container.ContainerCreateCreatedBody, error) {
	return cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, containerName)
}

//ContainerList returns the list of containers in the docker host.
func (dr dockerRepository) ContainerList(ctx context.Context, cli entity.Client,
	options types.ContainerListOptions) ([]types.Container, error) {

	return cli.ContainerList(ctx, options)
}

//ContainerRemove kills and removes a container from the docker host.
func (dr dockerRepository) ContainerRemove(ctx context.Context, cli entity.Client,
	containerID string, options types.ContainerRemoveOptions) error {

	return cli.ContainerRemove(ctx, containerID, options)
}

//ContainerStart sends a request to the docker daemon to start a container.
func (dr dockerRepository) ContainerStart(ctx context.Context, cli entity.Client,
	containerID string, options types.ContainerStartOptions) error {
	return cli.ContainerStart(ctx, containerID, options)
}

//CopyToContainer copies content into the container filesystem. Note that `content` must be a Reader for a TAR archive
func (dr dockerRepository) CopyToContainer(ctx context.Context, cli entity.Client,
	containerID, dstPath string, content io.Reader,
	options types.CopyToContainerOptions) error {
	return cli.CopyToContainer(ctx, containerID, dstPath, content, options)
}

//ImageLoad is used to upload a docker image
func (dr dockerRepository) ImageLoad(ctx context.Context, cli entity.Client,
	input io.Reader, quiet bool) (types.ImageLoadResponse, error) {

	return cli.ImageLoad(ctx, input, quiet)
}

//ImagePull is used to pull a docker image
func (dr dockerRepository) ImagePull(ctx context.Context, cli entity.Client,
	refStr string, options types.ImagePullOptions) (io.ReadCloser, error) {

	return cli.ImagePull(ctx, refStr, options)
}

//ImageList returns a list of images in the docker host
func (dr dockerRepository) ImageList(ctx context.Context, cli entity.Client,
	options types.ImageListOptions) ([]types.ImageSummary, error) {

	return cli.ImageList(ctx, options)
}

//NetworkConnect connects a container to an existent network in the docker host.
func (dr dockerRepository) NetworkConnect(ctx context.Context, cli entity.Client,
	networkID, containerID string, config *network.EndpointSettings) error {
	return cli.NetworkConnect(ctx, networkID, containerID, config)
}

//NetworkCreate sends a request to the docker daemon to create a network
func (dr dockerRepository) NetworkCreate(ctx context.Context, cli entity.Client, name string,
	options types.NetworkCreate) (types.NetworkCreateResponse, error) {
	return cli.NetworkCreate(ctx, name, options)
}

//NetworkDisconnect disconnects a container from an existent network in the docker host.
func (dr dockerRepository) NetworkDisconnect(ctx context.Context, cli entity.Client,
	networkID, containerID string, force bool) error {

	return cli.NetworkDisconnect(ctx, networkID, containerID, force)
}

//NetworkRemove sends a request to the docker daemon to remove a network
func (dr dockerRepository) NetworkRemove(ctx context.Context, cli entity.Client, networkID string) error {
	return cli.NetworkRemove(ctx, networkID)
}

//NetworkList lists the networks known to the docker daemon
func (dr dockerRepository) NetworkList(ctx context.Context, cli entity.Client,
	options types.NetworkListOptions) ([]types.NetworkResource, error) {
	return cli.NetworkList(ctx, options)
}

//VolumeList returns the volumes configured in the docker host.
func (dr dockerRepository) VolumeList(ctx context.Context, cli entity.Client,
	filter filters.Args) (volume.VolumeListOKBody, error) {
	return cli.VolumeList(ctx, filter)
}

//VolumeRemove removes a volume from the docker host.
func (dr dockerRepository) VolumeRemove(ctx context.Context, cli entity.Client, volumeID string, force bool) error {
	return cli.VolumeRemove(ctx, volumeID, force)
}

//VolumeCreate creates a volume in the container
func (dr dockerRepository) VolumeCreate(ctx context.Context, cli entity.Client,
	options volume.VolumeCreateBody) (types.Volume, error) {
	return cli.VolumeCreate(ctx, options)
}
