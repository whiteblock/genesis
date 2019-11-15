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
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/repository"
	"github.com/whiteblock/genesis/pkg/service/auxillary"
	"strconv"
)

//DockerService provides a intermediate interface between docker and the order from a command
type DockerService interface {

	//CreateContainer attempts to create a docker container
	CreateContainer(ctx context.Context, cli *client.Client, container command.Container) entity.Result

	//StartContainer attempts to start an already created docker container
	StartContainer(ctx context.Context, cli *client.Client, name string) entity.Result

	//RemoveContainer attempts to remove a container
	RemoveContainer(ctx context.Context, cli *client.Client, name string) entity.Result

	//CreateNetwork attempts to create a network
	CreateNetwork(ctx context.Context, cli *client.Client, net command.Network) entity.Result

	//RemoveNetwork attempts to remove a network
	RemoveNetwork(ctx context.Context, cli *client.Client, name string) entity.Result
	AttachNetwork(ctx context.Context, cli *client.Client, network string, container string) entity.Result
	DetachNetwork(ctx context.Context, cli *client.Client, network string, container string) entity.Result
	CreateVolume(ctx context.Context, cli *client.Client, volume command.Volume) entity.Result
	RemoveVolume(ctx context.Context, cli *client.Client, name string) entity.Result
	PlaceFileInContainer(ctx context.Context, cli *client.Client, containerName string, file command.IFile) entity.Result
	PlaceFileInVolume(ctx context.Context, cli *client.Client, volumeName string, file command.IFile) entity.Result
	Emulation(ctx context.Context, cli *client.Client, netem command.Netconf) entity.Result

	//CreateClient creates a new client for connecting to the docker daemon
	CreateClient(host string) (*client.Client, error)

	//GetNetworkingConfig determines the proper networking config based on the docker hosts state and the networks
	GetNetworkingConfig(ctx context.Context, cli *client.Client, networks strslice.StrSlice) (*network.NetworkingConfig, error)
}

type dockerService struct {
	repo repository.DockerRepository
	aux  auxillary.DockerAuxillary
	conf entity.DockerConfig
}

//NewDockerService creates a new DockerService
func NewDockerService(repo repository.DockerRepository, aux auxillary.DockerAuxillary,
	conf entity.DockerConfig) (DockerService, error) {
	return dockerService{conf: conf, repo: repo, aux: aux}, nil
}

//CreateClient creates a new client for connecting to the docker daemon
func (ds dockerService) CreateClient(host string) (*client.Client, error) {
	if ds.conf.LocalMode {
		return client.NewClientWithOpts(
			client.WithAPIVersionNegotiation(),
		)
	}
	return client.NewClientWithOpts(
		client.WithAPIVersionNegotiation(),
		client.WithHost(host),
		client.WithTLSClientConfig(ds.conf.CACertPath, ds.conf.CertPath, ds.conf.KeyPath),
	)
}

//GetNetworkingConfig determines the proper networking config based on the docker hosts state and the networks
func (ds dockerService) GetNetworkingConfig(ctx context.Context, cli *client.Client,
	networks strslice.StrSlice) (*network.NetworkingConfig, error) {

	resourceChan := make(chan types.NetworkResource, len(networks))
	errChan := make(chan error, len(networks))

	for _, net := range networks {
		go func(net string) {
			resource, err := ds.aux.GetNetworkByName(ctx, cli, net)
			errChan <- err
			resourceChan <- resource
		}(net)
	}
	out := &network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{}}
	for range networks {
		err := <-errChan
		if err != nil {
			return out, err
		}
		resource := <-resourceChan
		out.EndpointsConfig[resource.Name] = &network.EndpointSettings{
			NetworkID: resource.ID,
		}
	}
	return out, nil
}

//CreateContainer attempts to create a docker container
func (ds dockerService) CreateContainer(ctx context.Context, cli *client.Client, dContainer command.Container) entity.Result {
	errChan := make(chan error)
	netConfChan := make(chan *network.NetworkingConfig)
	defer close(errChan)
	defer close(netConfChan)

	go func(image string) {
		errChan <- ds.aux.EnsureImagePulled(ctx, cli, image)
	}(dContainer.Image)

	go func(networks strslice.StrSlice) {
		networkConfig, err := ds.GetNetworkingConfig(ctx, cli, networks)
		netConfChan <- networkConfig
		errChan <- err
	}(dContainer.Network)

	portSet, portMap, err := dContainer.GetPortBindings()
	if err != nil {
		return entity.NewFatalResult(err)
	}

	config := &container.Config{
		Hostname:     dContainer.Name,
		Domainname:   dContainer.Name,
		ExposedPorts: portSet,
		Env:          dContainer.GetEnv(),
		Image:        dContainer.Image,
		Entrypoint:   dContainer.GetEntryPoint(),
		Labels:       dContainer.Labels,
	}

	mem, err := dContainer.GetMemory()
	if err != nil {
		return entity.NewFatalResult(err)
	}

	cpus, err := strconv.ParseFloat(dContainer.Cpus, 64)
	if err != nil {
		return entity.NewFatalResult(err)
	}

	hostConfig := &container.HostConfig{
		PortBindings: portMap,
		AutoRemove:   true,
		Mounts:       dContainer.GetMounts(),
	}
	hostConfig.NanoCPUs = int64(1000000000 * cpus)
	hostConfig.Memory = mem

	networkConfig := <-netConfChan

	for i := 0; i < 2; i++ {
		err = <-errChan
		if err != nil {
			return entity.NewErrorResult(err)
		}
	}

	_, err = ds.repo.ContainerCreate(ctx, cli, config, hostConfig, networkConfig, dContainer.Name)
	if err != nil {
		return entity.NewFatalResult(err)
	}

	return entity.NewSuccessResult()
}

//StartContainer attempts to start an already created docker container
func (ds dockerService) StartContainer(ctx context.Context, cli *client.Client, name string) entity.Result {
	log.WithFields(log.Fields{"name": name}).Trace("starting container")
	opts := types.ContainerStartOptions{}
	err := ds.repo.ContainerStart(ctx, cli, name, opts)
	if err != nil {
		return entity.NewErrorResult(err)
	}

	return entity.NewSuccessResult()
}

//RemoveContainer attempts to remove a container
func (ds dockerService) RemoveContainer(ctx context.Context, cli *client.Client, name string) entity.Result {
	cntr, err := ds.aux.GetContainerByName(ctx, cli, name)
	if err != nil {
		return entity.NewErrorResult(err)
	}
	err = ds.repo.ContainerRemove(ctx, cli, cntr.ID, types.ContainerRemoveOptions{
		RemoveVolumes: false,
		RemoveLinks:   false,
		Force:         true,
	})
	if err != nil {
		return entity.NewErrorResult(err)
	}
	return entity.NewSuccessResult()
}

//CreateNetwork attempts to create a network
func (ds dockerService) CreateNetwork(ctx context.Context, cli *client.Client, net command.Network) entity.Result {
	networkCreate := types.NetworkCreate{
		CheckDuplicate: true,
		Attachable:     true,
		Ingress:        false,
		Internal:       false,
		Labels:         net.Labels,
		IPAM: &network.IPAM{
			Driver:  "default",
			Options: nil,
			Config: []network.IPAMConfig{
				network.IPAMConfig{
					Subnet:  net.Subnet,
					Gateway: net.Gateway,
				},
			},
		},
		Options: map[string]string{},
	}
	if net.Global {
		networkCreate.Driver = "overlay"
		networkCreate.Scope = "swarm"
	} else {
		networkCreate.Driver = "bridge"
		networkCreate.Scope = "local"
		networkCreate.Options["com.docker.network.bridge.name"] = net.Name
	}
	_, err := ds.repo.NetworkCreate(ctx, cli, net.Name, networkCreate)
	if err != nil {
		return entity.NewErrorResult(err)
	}
	return entity.NewSuccessResult()
}

//RemoveNetwork attempts to remove a network
func (ds dockerService) RemoveNetwork(ctx context.Context, cli *client.Client, name string) entity.Result {
	net, err := ds.aux.GetNetworkByName(ctx, cli, name)
	if err != nil {
		return entity.NewErrorResult(err)
	}
	err = ds.repo.NetworkRemove(ctx, cli, net.ID)
	if err != nil {
		return entity.NewErrorResult(err)
	}
	return entity.NewSuccessResult()
}

func (ds dockerService) getNetworkAndContainerByName(ctx context.Context, cli *client.Client, networkName string,
	containerName string) (container types.Container, network types.NetworkResource, err error) {
	errChan := make(chan error, 2)
	netChan := make(chan types.NetworkResource, 1)
	cntrChan := make(chan types.Container, 1)
	defer close(errChan)
	defer close(netChan)
	defer close(cntrChan)

	go func(networkName string) {
		net, err := ds.aux.GetNetworkByName(ctx, cli, networkName)
		errChan <- err
		netChan <- net
	}(networkName)

	go func(containerName string) {
		cntr, err := ds.aux.GetContainerByName(ctx, cli, containerName)
		errChan <- err
		cntrChan <- cntr
	}(containerName)

	for i := 0; i < 2; i++ {
		err = <-errChan
		if err != nil {
			return
		}
	}
	network = <-netChan
	container = <-cntrChan
	return
}

func (ds dockerService) AttachNetwork(ctx context.Context, cli *client.Client, networkName string,
	containerName string) entity.Result {

	cntr, net, err := ds.getNetworkAndContainerByName(ctx, cli, networkName, containerName)
	if err != nil {
		return entity.NewFatalResult(err)
	}

	err = ds.repo.NetworkConnect(ctx, cli, net.ID, cntr.ID, &network.EndpointSettings{
		NetworkID: net.ID,
	})
	if err != nil {
		return entity.NewErrorResult(err)
	}
	return entity.NewSuccessResult()
}

func (ds dockerService) DetachNetwork(ctx context.Context, cli *client.Client,
	networkName string, containerName string) entity.Result {

	cntr, net, err := ds.getNetworkAndContainerByName(ctx, cli, networkName, containerName)
	if err != nil {
		return entity.NewFatalResult(err)
	}

	err = ds.repo.NetworkDisconnect(ctx, cli, net.ID, cntr.ID, true)
	if err != nil {
		return entity.NewErrorResult(err)
	}
	return entity.NewSuccessResult()
}

func (ds dockerService) CreateVolume(ctx context.Context, cli *client.Client, vol command.Volume) entity.Result {
	volConfig := volume.VolumeCreateBody{
		Driver:     vol.Driver,
		DriverOpts: vol.DriverOpts,
		Labels:     vol.Labels,
		Name:       vol.Name,
	}

	_, err := ds.repo.VolumeCreate(ctx, cli, volConfig)
	if err != nil {
		return entity.NewErrorResult(err)
	}

	return entity.NewSuccessResult()
}

func (ds dockerService) RemoveVolume(ctx context.Context, cli *client.Client, name string) entity.Result {
	err := ds.repo.VolumeRemove(ctx, cli, name, true)
	if err != nil {
		return entity.NewErrorResult(err)
	}
	return entity.NewSuccessResult()
}

func (ds dockerService) PlaceFileInContainer(ctx context.Context, cli *client.Client,
	containerName string, file command.IFile) entity.Result {

	cntr, err := ds.aux.GetContainerByName(ctx, cli, containerName)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	rdr, err := file.GetTarReader()
	if err != nil {
		return entity.NewFatalResult(err)
	}
	err = ds.repo.CopyToContainer(ctx, cli, cntr.ID, file.GetDir(), rdr, types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: false,
		CopyUIDGID:                false,
	})
	if err != nil {
		return entity.NewErrorResult(err)
	}
	return entity.NewSuccessResult()
}

func (ds dockerService) PlaceFileInVolume(ctx context.Context, cli *client.Client, volumeName string, file command.IFile) entity.Result {
	//TODO
	return entity.Result{}
}

func (ds dockerService) Emulation(ctx context.Context, cli *client.Client, netem command.Netconf) entity.Result {
	netemImage := "gaiadocker/iproute2:latest"
	errChan := make(chan error)
	defer close(errChan)
	go func() {
		errChan <- ds.aux.EnsureImagePulled(ctx, cli, netemImage)
	}()

	cntr, net, err := ds.getNetworkAndContainerByName(ctx, cli, netem.Network, netem.Container)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	err = <-errChan
	name := cntr.ID + "-" + net.ID
	netemCmd := fmt.Sprintf(
		"tc qdisc add dev $(ip -o addr show to %s | sed -n 's/.*\\(eth[0-9]*\\).*/\\1/p') root netem",
		net.IPAM.Config[0].Subnet)

	if netem.Limit > 0 {
		netemCmd += fmt.Sprintf(" limit %d", netem.Limit)
	}

	if netem.Loss > 0 {
		netemCmd += fmt.Sprintf(" loss %.4f", netem.Loss)
	}

	if netem.Delay > 0 {
		netemCmd += fmt.Sprintf(" delay %dus", netem.Delay)
	}

	if len(netem.Rate) > 0 {
		netemCmd += fmt.Sprintf(" rate %s", netem.Rate)
	}

	if netem.Duplication > 0 {
		netemCmd += fmt.Sprintf(" duplicate %.4f", netem.Duplication)
	}

	if netem.Corrupt > 0 {
		netemCmd += fmt.Sprintf(" corrupt %.4f", netem.Duplication)
	}

	if netem.Reorder > 0 {
		netemCmd += fmt.Sprintf(" reorder %.4f", netem.Reorder)
	}

	config := &container.Config{
		Image:      netemImage,
		Entrypoint: strslice.StrSlice([]string{"/bin/sh", "-c", netemCmd}),
	}

	hostConfig := &container.HostConfig{
		AutoRemove:  true,
		NetworkMode: container.NetworkMode(fmt.Sprintf("container:%s", cntr.ID)),
		CapAdd:      strslice.StrSlice([]string{"NET_ADMIN"}),
	}

	networkConfig := &network.NetworkingConfig{}

	_, err = ds.repo.ContainerCreate(ctx, cli, config, hostConfig, networkConfig, name)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return ds.StartContainer(ctx, cli, name)
}
