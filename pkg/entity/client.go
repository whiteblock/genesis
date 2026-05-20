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
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/volume"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Client is the subset of [github.com/docker/docker/client.APIClient] used by genesis.
// A real *client.Client returned from client.NewClientWithOpts satisfies this interface.
type Client interface {
	Close() error
	DaemonHost() string
	HTTPClient() *http.Client
	Ping(ctx context.Context) (types.Ping, error)

	// Containers
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
	ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error)
	ContainerRemove(ctx context.Context, container string, options container.RemoveOptions) error
	ContainerStart(ctx context.Context, container string, options container.StartOptions) error
	ContainerStatPath(ctx context.Context, container, path string) (container.PathStat, error)
	ContainerWait(ctx context.Context, container string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)
	CopyToContainer(ctx context.Context, container, path string, content io.Reader, options container.CopyToContainerOptions) error

	// Exec
	ContainerExecCreate(ctx context.Context, container string, options container.ExecOptions) (container.ExecCreateResponse, error)
	ContainerExecInspect(ctx context.Context, execID string) (container.ExecInspect, error)
	ContainerExecStart(ctx context.Context, execID string, options container.ExecStartOptions) error

	// Images
	ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error)
	ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error)

	// Networks
	NetworkConnect(ctx context.Context, network, container string, config *network.EndpointSettings) error
	NetworkCreate(ctx context.Context, name string, options network.CreateOptions) (network.CreateResponse, error)
	NetworkDisconnect(ctx context.Context, network, container string, force bool) error
	NetworkList(ctx context.Context, options network.ListOptions) ([]network.Summary, error)
	NetworkRemove(ctx context.Context, network string) error

	// Swarm
	SwarmInit(ctx context.Context, req swarm.InitRequest) (string, error)
	SwarmInspect(ctx context.Context) (swarm.Swarm, error)
	SwarmJoin(ctx context.Context, req swarm.JoinRequest) error

	// Volumes
	VolumeCreate(ctx context.Context, options volume.CreateOptions) (volume.Volume, error)
	VolumeRemove(ctx context.Context, volumeID string, force bool) error
}
