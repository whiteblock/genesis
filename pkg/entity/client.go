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
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"io"
)

//Client is an interface which contains the needed methods from the docker client
type Client interface {
	//ContainerAttach attaches a connection to a container in the server. It returns a types.HijackedConnection with
	//the hijacked connection and the a reader to get output. It's up to the called to close
	//the hijacked connection by calling types.HijackedResponse.Close.
	ContainerAttach(ctx context.Context, container string, options types.ContainerAttachOptions) (types.HijackedResponse, error)

	//ContainerCreate creates a new container based in the given configuration. It can be associated with a name, but it's not mandatory.
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error)

	//ContainerList returns the list of containers in the docker host.
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)

	//ContainerRemove kills and removes a container from the docker host.
	ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error

	//ContainerStart sends a request to the docker daemon to start a container.
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error

	//CopyToContainer copies content into the container filesystem. Note that `content` must be a Reader for a TAR archive
	CopyToContainer(ctx context.Context, containerID, dstPath string, content io.Reader,
		options types.CopyToContainerOptions) error

	//ImageList returns a list of images in the docker host
	ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error)

	//ImageLoad is used to upload a docker image
	ImageLoad(ctx context.Context, input io.Reader, quiet bool) (types.ImageLoadResponse, error)

	//ImagePull is used to pull a docker image
	ImagePull(ctx context.Context, refStr string, options types.ImagePullOptions) (io.ReadCloser, error)

	//NetworkCreate sends a request to the docker daemon to create a network
	NetworkCreate(ctx context.Context, name string, options types.NetworkCreate) (types.NetworkCreateResponse, error)

	//NetworkConnect connects a container to an existent network in the docker host.
	NetworkConnect(ctx context.Context, networkID, containerID string, config *network.EndpointSettings) error

	//NetworkDisconnect disconnects a container from an existent network in the docker host.
	NetworkDisconnect(ctx context.Context, networkID, containerID string, force bool) error

	//NetworkInspect returns the information for a specific network configured in the docker host.
	NetworkInspect(ctx context.Context, networkID string, options types.NetworkInspectOptions) (types.NetworkResource, error)

	//NetworkRemove sends a request to the docker daemon to remove a network
	NetworkRemove(ctx context.Context, networkID string) error

	//NetworkList lists the networks known to the docker daemon
	NetworkList(ctx context.Context, options types.NetworkListOptions) ([]types.NetworkResource, error)

	VolumeCreate(ctx context.Context, options volume.VolumeCreateBody) (types.Volume, error)

	//VolumeList returns the volumes configured in the docker host.
	VolumeList(ctx context.Context, filter filters.Args) (volume.VolumeListOKBody, error)

	//VolumeRemove removes a volume from the docker host.
	VolumeRemove(ctx context.Context, volumeID string, force bool) error
}
