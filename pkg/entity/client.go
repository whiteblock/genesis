/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package entity

import (
	"context"
	"io"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/volume"
)

// Client is an interface which contains the needed methods from the docker client
type Client interface {
	// Close the transport used by the client
	Close() error
	// ContainerAttach attaches a connection to a container in the server. It returns a types.HijackedConnection with
	// the hijacked connection and the a reader to get output. It's up to the called to close
	// the hijacked connection by calling types.HijackedResponse.Close.
	ContainerAttach(ctx context.Context, container string, options types.ContainerAttachOptions) (types.HijackedResponse, error)

	// ContainerCreate creates a new container based in the given configuration. It can be associated with a name, but it's not mandatory.
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error)

	// ContainerExecAttach attaches a connection to an exec process in the server. It returns a
	// types.HijackedConnection with the hijacked connection and the a reader to get output. It's up
	// to the called to close the hijacked connection by calling types.HijackedResponse.Close.
	ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error)

	// ContainerExecCreate creates a new exec configuration to run an exec process.
	ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error)

	// ContainerExecInspect returns information about a specific exec process on the docker host.
	ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error)

	// ContainerExecStart starts an exec process already created in the docker host.
	ContainerExecStart(ctx context.Context, execID string, config types.ExecStartCheck) error

	// ContainerInspect returns the container information.
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)

	// ContainerList returns the list of containers in the docker host.
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)

	// ContainerRemove kills and removes a container from the docker host.
	ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error

	// ContainerStart sends a request to the docker daemon to start a container.
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error

	// ContainerStatPath returns Stat information about a path inside the container filesystem.
	ContainerStatPath(ctx context.Context, containerID, path string) (types.ContainerPathStat, error)

	// ContainerWait waits until the specified container is in a certain state indicated by the given condition, either "not-running" (default), "next-exit", or "removed".
	ContainerWait(ctx context.Context, containerID string,
		condition container.WaitCondition) (<-chan container.ContainerWaitOKBody, <-chan error)

	// CopyToContainer copies content into the container filesystem. Note that `content` must be a Reader for a TAR archive
	CopyToContainer(ctx context.Context, containerID, dstPath string, content io.Reader,
		options types.CopyToContainerOptions) error

	// DaemonHost returns the host address used by the client
	DaemonHost() string

	// HTTPClient returns a copy of the HTTP client bound to the server
	HTTPClient() *http.Client

	// ImageList returns a list of images in the docker host
	ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error)

	//ImageLoad is used to upload a docker image
	ImageLoad(ctx context.Context, input io.Reader, quiet bool) (types.ImageLoadResponse, error)

	//ImagePull is used to pull a docker image
	ImagePull(ctx context.Context, refStr string, options types.ImagePullOptions) (io.ReadCloser, error)

	// NetworkCreate sends a request to the docker daemon to create a network
	NetworkCreate(ctx context.Context, name string, options types.NetworkCreate) (types.NetworkCreateResponse, error)

	// NetworkConnect connects a container to an existent network in the docker host.
	NetworkConnect(ctx context.Context, networkID, containerID string, config *network.EndpointSettings) error

	// NetworkDisconnect disconnects a container from an existent network in the docker host.
	NetworkDisconnect(ctx context.Context, networkID, containerID string, force bool) error

	// NetworkInspect returns the information for a specific network configured in the docker host.
	NetworkInspect(ctx context.Context, networkID string, options types.NetworkInspectOptions) (types.NetworkResource, error)

	// NetworkRemove sends a request to the docker daemon to remove a network
	NetworkRemove(ctx context.Context, networkID string) error

	// NetworkList lists the networks known to the docker daemon
	NetworkList(ctx context.Context, options types.NetworkListOptions) ([]types.NetworkResource, error)

	// Ping pings the server and returns the value of the "Docker-Experimental", "Builder-Version",
	// "OS-Type" & "API-Version" headers. It attempts to use a HEAD request on the endpoint, but
	// falls back to GET if HEAD is not supported by the daemon.
	Ping(ctx context.Context) (types.Ping, error)

	// SwarmInit initializes the swarm.
	SwarmInit(ctx context.Context, req swarm.InitRequest) (string, error)

	// SwarmJoin joins the swarm.
	SwarmJoin(ctx context.Context, req swarm.JoinRequest) error

	// SwarmInspect inspects the swarm.
	SwarmInspect(ctx context.Context) (swarm.Swarm, error)

	// VolumeCreate creates a volume in the docker host.
	VolumeCreate(ctx context.Context, options volume.VolumeCreateBody) (types.Volume, error)

	// VolumeList returns the volumes configured in the docker host.
	VolumeList(ctx context.Context, filter filters.Args) (volume.VolumeListOKBody, error)

	// VolumeRemove removes a volume from the docker host.
	VolumeRemove(ctx context.Context, volumeID string, force bool) error
}
