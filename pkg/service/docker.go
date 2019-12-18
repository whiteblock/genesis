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
	"strconv"
	"strings"
	"time"

	"github.com/whiteblock/genesis/pkg/config"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/file"
	"github.com/whiteblock/genesis/pkg/repository"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/whiteblock/definition/command"
)

//DockerService provides a intermediate interface between docker and the order from a command
type DockerService interface {

	//CreateContainer attempts to create a docker container
	CreateContainer(ctx context.Context, cli entity.DockerCli, container command.Container) entity.Result

	//StartContainer attempts to start an already created docker container
	StartContainer(ctx context.Context, cli entity.DockerCli, sc command.StartContainer) entity.Result

	//RemoveContainer attempts to remove a container
	RemoveContainer(ctx context.Context, cli entity.DockerCli, name string) entity.Result

	//CreateNetwork attempts to create a network
	CreateNetwork(ctx context.Context, cli entity.DockerCli, net command.Network) entity.Result

	//RemoveNetwork attempts to remove a network
	RemoveNetwork(ctx context.Context, cli entity.DockerCli, name string) entity.Result
	AttachNetwork(ctx context.Context, cli entity.DockerCli, network string, container string) entity.Result
	DetachNetwork(ctx context.Context, cli entity.DockerCli, network string, container string) entity.Result
	CreateVolume(ctx context.Context, cli entity.DockerCli, volume command.Volume) entity.Result
	RemoveVolume(ctx context.Context, cli entity.DockerCli, name string) entity.Result
	PlaceFileInContainer(ctx context.Context, cli entity.DockerCli,
		containerName string, file command.File) entity.Result
	Emulation(ctx context.Context, cli entity.DockerCli, netem command.Netconf) entity.Result
	SwarmCluster(ctx context.Context, cli entity.DockerCli, swarm command.SetupSwarm) entity.Result
	PullImage(ctx context.Context, cli entity.DockerCli, imagePull command.PullImage) entity.Result

	//CreateClient creates a new client for connecting to the docker daemon
	CreateClient(host string) (entity.Client, error)

	// GetNetworkingConfig determines the proper networking config based
	// on the docker hosts state and the networks
	GetNetworkingConfig(ctx context.Context, cli entity.DockerCli,
		networks strslice.StrSlice) (*network.NetworkingConfig, error)
}

type dockerService struct {
	repo   repository.DockerRepository
	conf   config.Docker
	log    logrus.Ext1FieldLogger
	remote file.RemoteSources
}

//NewDockerService creates a new DockerService
func NewDockerService(
	repo repository.DockerRepository,
	conf config.Docker,
	remote file.RemoteSources,
	log logrus.Ext1FieldLogger) DockerService {

	return dockerService{
		conf:   conf,
		repo:   repo,
		remote: remote,
		log:    log}
}

//CreateClient creates a new client for connecting to the docker daemon
func (ds dockerService) CreateClient(host string) (entity.Client, error) {
	if ds.conf.LocalMode {
		return client.NewClientWithOpts(
			client.WithAPIVersionNegotiation(),
		)
	}
	return client.NewClientWithOpts(
		client.WithAPIVersionNegotiation(),
		client.WithHost("tcp://"+host+":"+ds.conf.DaemonPort),
		ds.repo.WithTLSClientConfig(ds.conf.CACertPath, ds.conf.CertPath, ds.conf.KeyPath),
	)
}

func (ds dockerService) withFields(cli entity.DockerCli, fields logrus.Fields) *logrus.Entry {
	for key, value := range cli.Labels {
		fields[key] = value
	}
	return ds.log.WithFields(fields)
}

func (ds dockerService) withField(cli entity.DockerCli, key string, value interface{}) *logrus.Entry {
	return ds.withFields(cli, logrus.Fields{key: value})
}

// GetNetworkingConfig determines the proper networking
// config based on the docker hosts state and the networks
func (ds dockerService) GetNetworkingConfig(ctx context.Context, cli entity.DockerCli,
	networks strslice.StrSlice) (*network.NetworkingConfig, error) {

	resourceChan := make(chan types.NetworkResource, len(networks))
	errChan := make(chan error, len(networks))

	for _, net := range networks {
		go func(net string) {
			resource, err := ds.repo.GetNetworkByName(ctx, cli, net)
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
func (ds dockerService) CreateContainer(ctx context.Context, cli entity.DockerCli,
	dContainer command.Container) entity.Result {

	ds.withFields(cli, logrus.Fields{"container": dContainer}).Trace("create container")
	errChan := make(chan error)
	netConfChan := make(chan *network.NetworkingConfig)

	go func(image string) {
		errChan <- ds.repo.EnsureImagePulled(ctx, cli, image, "")
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
		Labels:       cli.Labels,
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
		LogConfig: container.LogConfig{
			Type: ds.conf.LogDriver,
			Config: map[string]string{
				"labels": ds.conf.LogLabels,
			},
		},
		Mounts: dContainer.GetMounts(),
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

	_, err = cli.ContainerCreate(ctx, config, hostConfig, networkConfig, dContainer.Name)
	if err != nil {
		if strings.Contains(err.Error(), "already in use by container") {
			ds.withFields(cli, logrus.Fields{"name": dContainer.Name,
				"error": err}).Error("duplicate container error")
			return entity.NewSuccessResult()
		}
		return entity.NewFatalResult(err).InjectMeta(map[string]interface{}{
			"image":  dContainer.Image,
			"name":   dContainer.Name,
			"errMsg": err.Error(),
			"type":   "CreateContainer",
		})
	}

	return entity.NewSuccessResult().InjectMeta(map[string]interface{}{
		"networks": dContainer.Network,
		"name":     dContainer.Name,
	})
}

//StartContainer attempts to start an already created docker container
func (ds dockerService) StartContainer(ctx context.Context, cli entity.DockerCli,
	sc command.StartContainer) entity.Result {

	ds.withFields(cli, logrus.Fields{"name": sc.Name}).Trace("starting container")
	opts := types.ContainerStartOptions{}

	err := cli.ContainerStart(ctx, sc.Name, opts)
	if err != nil {
		return entity.NewErrorResult(err).InjectMeta(map[string]interface{}{
			"name": sc.Name,
			"type": "StartContainer",
		})
	}

	if !sc.Attach {
		return entity.NewSuccessResult()
	}
	if sc.Timeout.IsInfinite() { // Trap to stop any further execution
		return entity.NewTrapResult().InjectMeta(map[string]interface{}{
			"name": sc.Name,
		})
	}

	resChan := make(chan error)
	ctx2, cancelFn := context.WithTimeout(context.Background(), sc.Timeout.Duration)
	defer cancelFn()

	go func() {
		startTime := time.Now().Add(sc.Timeout.Duration)
		for time.Now().Unix() < startTime.Unix() {
			ds.withFields(cli, logrus.Fields{
				"name": sc.Name}).Trace("checking container status")
			_, err := cli.ContainerInspect(ctx2, sc.Name)
			if err != nil {
				ds.withFields(cli, logrus.Fields{
					"name":  sc.Name,
					"error": err.Error()}).Info("container finished execution")
				resChan <- err
			}
			time.Sleep(1 * time.Second)
		}
	}()

	select {
	case err := <-resChan:
		if err != nil {
			if strings.Contains(err.Error(), "No such container") {
				return entity.NewSuccessResult()
			}
			return entity.NewErrorResult(err).InjectMeta(map[string]interface{}{
				"name":  sc.Name,
				"type":  "StartContainer",
				"error": err.Error(),
			})
		}
	case <-time.After(sc.Timeout.Duration):
		ds.withFields(cli, logrus.Fields{"name": sc.Name}).Debug("timeout was reached")

	}
	return entity.NewSuccessResult()
}

//RemoveContainer attempts to remove a container
func (ds dockerService) RemoveContainer(ctx context.Context, cli entity.DockerCli,
	name string) entity.Result {

	ds.withFields(cli, logrus.Fields{"name": name}).Debug("removing container")
	cntr, err := ds.repo.GetContainerByName(ctx, cli, name)
	if err != nil {
		return entity.NewErrorResult(err)
	}
	err = cli.ContainerRemove(ctx, cntr.ID, types.ContainerRemoveOptions{
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
func (ds dockerService) CreateNetwork(ctx context.Context, cli entity.DockerCli,
	net command.Network) entity.Result {

	networkCreate := types.NetworkCreate{
		CheckDuplicate: true,
		Attachable:     true,
		Ingress:        false,
		Internal:       false,
		Labels:         cli.Labels,
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
	ds.withFields(cli, logrus.Fields{"name": net.Name, "conf": networkCreate}).Debug("creating a network")
	_, err := cli.NetworkCreate(ctx, net.Name, networkCreate)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			ds.withFields(cli, logrus.Fields{"name": net.Name, "error": err}).Error("duplicate error")
			return entity.NewSuccessResult()
		}
		return entity.NewErrorResult(err)
	}
	return entity.NewSuccessResult()
}

//RemoveNetwork attempts to remove a network
func (ds dockerService) RemoveNetwork(ctx context.Context, cli entity.DockerCli, name string) entity.Result {
	ds.withFields(cli, logrus.Fields{"name": name}).Debug("removing a network")
	net, err := ds.repo.GetNetworkByName(ctx, cli, name)
	if err != nil {
		return entity.NewErrorResult(err)
	}
	err = cli.NetworkRemove(ctx, net.ID)
	if err != nil {
		return entity.NewErrorResult(err)
	}
	return entity.NewSuccessResult()
}

func (ds dockerService) getNetworkAndContainerByName(ctx context.Context, cli entity.DockerCli,
	networkName string, containerName string) (container types.Container,
	network types.NetworkResource, err error) {

	errChan := make(chan error, 2)
	netChan := make(chan types.NetworkResource, 1)
	cntrChan := make(chan types.Container, 1)

	go func(networkName string) {
		net, err := ds.repo.GetNetworkByName(ctx, cli, networkName)
		errChan <- err
		netChan <- net
	}(networkName)

	go func(containerName string) {
		cntr, err := ds.repo.GetContainerByName(ctx, cli, containerName)
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

func (ds dockerService) AttachNetwork(ctx context.Context, cli entity.DockerCli, networkName string,
	containerName string) entity.Result {

	err := cli.NetworkConnect(ctx, networkName, containerName, &network.EndpointSettings{})
	if err != nil {
		if strings.Contains(err.Error(), "is already attached to network") {
			ds.withField(cli, "error", err).Info("ignoring failure on deplicate network attach command")
			return entity.NewSuccessResult().InjectMeta(map[string]interface{}{
				"failure": "ignored",
			})
		}
		return entity.NewErrorResult(err)
	}
	return entity.NewSuccessResult()
}

func (ds dockerService) DetachNetwork(ctx context.Context, cli entity.DockerCli,
	networkName string, containerName string) entity.Result {

	err := cli.NetworkDisconnect(ctx, networkName, containerName, true)
	if err != nil {
		if strings.Contains(err.Error(), "is not connected to the network") {
			ds.withField(cli, "error", err).Info("ignoring failure on detach command")
			return entity.NewSuccessResult().InjectMeta(map[string]interface{}{
				"failure": "ignored",
			})
		}
		return entity.NewErrorResult(err)
	}
	return entity.NewSuccessResult()
}

func (ds dockerService) CreateVolume(ctx context.Context, cli entity.DockerCli,
	vol command.Volume) entity.Result {

	volConfig := volume.VolumeCreateBody{
		Driver:     vol.Driver,
		DriverOpts: vol.DriverOpts,
		Labels:     vol.Labels,
		Name:       vol.Name,
	}

	_, err := cli.VolumeCreate(ctx, volConfig)
	if err != nil {
		return entity.NewErrorResult(err)
	}

	return entity.NewSuccessResult()
}

func (ds dockerService) RemoveVolume(ctx context.Context, cli entity.DockerCli, name string) entity.Result {
	err := cli.VolumeRemove(ctx, name, true)
	if err != nil {
		return entity.NewErrorResult(err)
	}
	return entity.NewSuccessResult()
}

func (ds dockerService) PlaceFileInContainer(ctx context.Context, cli entity.DockerCli,
	containerName string, file command.File) entity.Result {

	cntr, err := ds.repo.GetContainerByName(ctx, cli, containerName)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	rdr, err := ds.remote.GetTarReader(cli.Labels["org"], file)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	err = cli.CopyToContainer(ctx, cntr.ID, file.Destination, rdr, types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: false,
		CopyUIDGID:                false,
	})
	if err != nil {
		return entity.NewErrorResult(err)
	}
	return entity.NewSuccessResult()
}

func (ds dockerService) Emulation(ctx context.Context, cli entity.DockerCli,
	netem command.Netconf) entity.Result {

	netemImage := "gaiadocker/iproute2:latest"
	errChan := make(chan error, 1)
	go func() {
		errChan <- ds.repo.EnsureImagePulled(ctx, cli, netemImage, "")
	}()

	cntr, net, err := ds.getNetworkAndContainerByName(ctx, cli, netem.Network, netem.Container)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	err = <-errChan
	if err != nil {
		return entity.NewFatalResult(err)
	}

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

	_, err = cli.ContainerCreate(ctx, config, hostConfig, networkConfig, name)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return ds.StartContainer(ctx, cli, command.StartContainer{Name: name})
}

func (ds dockerService) SwarmCluster(ctx context.Context, entryCLI entity.DockerCli,
	dswarm command.SetupSwarm) entity.Result {

	if len(dswarm.Hosts) == 0 {
		return entity.NewFatalResult("no hosts given")
	}
	cli, err := ds.CreateClient(dswarm.Hosts[0])
	if err != nil {
		ds.withField(entryCLI, "error", err).Error("creating the manager client")
		return entity.NewErrorResult(err)
	}
	token, err := cli.SwarmInit(ctx, swarm.InitRequest{
		ListenAddr:      fmt.Sprintf("0.0.0.0:%d", ds.conf.SwarmPort),
		AdvertiseAddr:   fmt.Sprintf("%s:%d", dswarm.Hosts[0], ds.conf.SwarmPort),
		ForceNewCluster: true,
		Availability:    swarm.NodeAvailabilityActive,
	})
	if err != nil {
		ds.withField(entryCLI, "error", err).Error("error with docker swarm init")
		return entity.NewErrorResult(err)
	}

	details, err := cli.SwarmInspect(ctx)
	if err != nil {
		ds.withField(entryCLI, "error", err).Error("error with docker swarm inspect")
		return entity.NewErrorResult(err)
	}

	ds.withField(entryCLI, "manager", token).Info("initializing docker swarm")
	if len(dswarm.Hosts) == 1 {
		return entity.NewSuccessResult()
	}

	for _, host := range dswarm.Hosts[1:] {
		cli, err := ds.CreateClient(host)
		if err != nil {
			return entity.NewErrorResult(err)
		}
		ds.withField(entryCLI, "token", details.JoinTokens.Worker).Info("adding worker to swarm")
		err = cli.SwarmJoin(ctx, swarm.JoinRequest{
			ListenAddr:    fmt.Sprintf("0.0.0.0:%d", ds.conf.SwarmPort),
			AdvertiseAddr: fmt.Sprintf("%s:%d", host, ds.conf.SwarmPort),
			RemoteAddrs:   []string{fmt.Sprintf("%s:%d", dswarm.Hosts[0], ds.conf.SwarmPort)},
			JoinToken:     details.JoinTokens.Worker,
			Availability:  swarm.NodeAvailabilityActive,
		})
		if err != nil {
			return entity.NewErrorResult(err)
		}
	}
	return entity.NewSuccessResult()
}

func (ds dockerService) PullImage(ctx context.Context, cli entity.DockerCli,
	imagePull command.PullImage) entity.Result {

	ds.withFields(cli, logrus.Fields{
		"image":     imagePull.Image,
		"usingAuth": imagePull.RegistryAuth != "",
	}).Debug("pre-emptively pulling an image if it doesn't exist")

	err := ds.repo.EnsureImagePulled(ctx, cli, imagePull.Image, imagePull.RegistryAuth)
	if err != nil {
		ds.withFields(cli, logrus.Fields{
			"image": imagePull.Image,
			"error": err,
		}).Error("unable to pull an image")
		return entity.NewErrorResult(err)
	}
	return entity.NewSuccessResult()
}
