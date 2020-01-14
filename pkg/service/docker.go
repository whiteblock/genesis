/*
	Copyright 2019 Whiteblock Inc.
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
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
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
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/system"
	"github.com/sirupsen/logrus"
	"github.com/whiteblock/definition/command"
)

//DockerService provides a intermediate interface between docker and the order from a command
type DockerService interface {

	//CreateContainer attempts to create a docker container
	CreateContainer(ctx context.Context, cli entity.DockerCli,
		container command.Container) entity.Result

	//StartContainer attempts to start an already created docker container
	StartContainer(ctx context.Context, cli entity.DockerCli, sc command.StartContainer) entity.Result

	//RemoveContainer attempts to remove a container
	RemoveContainer(ctx context.Context, cli entity.DockerCli, name string) entity.Result

	//CreateNetwork attempts to create a network
	CreateNetwork(ctx context.Context, cli entity.DockerCli, net command.Network) entity.Result

	//RemoveNetwork attempts to remove a network
	RemoveNetwork(ctx context.Context, cli entity.DockerCli, name string) entity.Result
	AttachNetwork(ctx context.Context, cli entity.DockerCli, cmd command.ContainerNetwork) entity.Result
	DetachNetwork(ctx context.Context, cli entity.DockerCli, network string,
		container string) entity.Result
	CreateVolume(ctx context.Context, cli entity.DockerCli, volume command.Volume) entity.Result
	RemoveVolume(ctx context.Context, cli entity.DockerCli, name string) entity.Result
	PlaceFileInContainer(ctx context.Context, cli entity.DockerCli,
		containerName string, file command.File) entity.Result
	Emulation(ctx context.Context, cli entity.DockerCli, netem command.Netconf) entity.Result
	SwarmCluster(ctx context.Context, cli entity.DockerCli, swarm command.SetupSwarm) entity.Result
	PullImage(ctx context.Context, cli entity.DockerCli, imagePull command.PullImage) entity.Result
	VolumeShare(ctx context.Context, cli entity.DockerCli, vs command.VolumeShare) entity.Result

	//CreateClient creates a new client for connecting to the docker daemon
	CreateClient(host string) (entity.Client, error)
}

var (
	// ErrNoHost is returned when the swarm init command does not contain any hosts
	ErrNoHost = entity.NewFatalResult("no hosts given")
)

const (
	//GlusterContainerName is the name of the gluster container
	GlusterContainerName = "gluster-container"
)

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

func (ds dockerService) errorWhitelistHandler(err error, whitelist ...string) entity.Result {
	if err == nil {
		return entity.NewResult(nil, 1)
	}
	for _, entry := range whitelist {
		if strings.Contains(err.Error(), entry) {
			ds.log.WithField("error", err).Info("ignoring whitelisted error")
			return entity.NewResult(nil, 1).InjectMeta(map[string]interface{}{
				"error": err,
			})
		}
	}
	return entity.NewResult(err, 1)
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

//CreateContainer attempts to create a docker container
func (ds dockerService) CreateContainer(ctx context.Context, cli entity.DockerCli,
	dContainer command.Container) entity.Result {

	ds.withFields(cli, logrus.Fields{"container": dContainer}).Trace("create container")
	errChan := make(chan error)

	go func(image string) {
		errChan <- ds.repo.EnsureImagePulled(ctx, cli, image, "")
	}(dContainer.Image)

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
		return entity.NewFatalResult(err).InjectMeta(map[string]interface{}{
			"given": dContainer.Cpus,
		})
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

	networkConfig := &network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{}}
	if len(dContainer.Network) > 0 {
		networkConfig.EndpointsConfig[dContainer.Network] = &network.EndpointSettings{
			NetworkID: dContainer.Network,
			IPAddress: dContainer.IP,
		}
	}

	err = <-errChan
	if err != nil {
		return entity.NewErrorResult(err)
	}

	_, err = cli.ContainerCreate(ctx, config, hostConfig, networkConfig, dContainer.Name)
	res := ds.errorWhitelistHandler(err, "already in use by container")
	if !res.IsSuccess() {
		res = res.Fatal()
	}
	return res.InjectMeta(map[string]interface{}{
		"image":   dContainer.Image,
		"name":    dContainer.Name,
		"network": dContainer.Network,
		"type":    "CreateContainer",
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
				break
			}
			time.Sleep(1 * time.Second)
		}
	}()

	select {
	case err := <-resChan:
		return ds.errorWhitelistHandler(err, "No such container").InjectMeta(map[string]interface{}{
			"name": sc.Name,
			"type": "StartContainer",
		})
	case <-time.After(sc.Timeout.Duration):
		ds.withFields(cli, logrus.Fields{"name": sc.Name}).Debug("timeout was reached")

	}
	return entity.NewSuccessResult()
}

// RemoveContainer attempts to remove a container
func (ds dockerService) RemoveContainer(ctx context.Context, cli entity.DockerCli,
	name string) entity.Result {

	ds.withFields(cli, logrus.Fields{"name": name}).Debug("removing container")

	err := cli.ContainerRemove(ctx, name, types.ContainerRemoveOptions{
		RemoveVolumes: false,
		RemoveLinks:   false,
		Force:         true,
	})
	return entity.NewResult(err)
}

// CreateNetwork attempts to create a network
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
	ds.withFields(cli, logrus.Fields{"name": net.Name,
		"conf": networkCreate}).Debug("creating a network")
	_, err := cli.NetworkCreate(ctx, net.Name, networkCreate)

	return ds.errorWhitelistHandler(err, "already exists")
}

//RemoveNetwork attempts to remove a network
func (ds dockerService) RemoveNetwork(ctx context.Context, cli entity.DockerCli,
	name string) entity.Result {

	ds.withFields(cli, logrus.Fields{"name": name}).Debug("removing a network")
	return entity.NewResult(cli.NetworkRemove(ctx, name))
}

func generateMacAddress() (string, error) {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	// Set the local bit
	buf[0] |= (buf[0] | 2) & 0xfe
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5]), nil
}

func (ds dockerService) AttachNetwork(ctx context.Context, cli entity.DockerCli,
	cmd command.ContainerNetwork) entity.Result {

	ds.withField(cli, "cmd", cmd).Info("attaching a network")
	macAddress, err := generateMacAddress()
	if err != nil {
		return ds.errorWhitelistHandler(err)
	}
	err = cli.NetworkConnect(ctx, cmd.Network, cmd.Container, &network.EndpointSettings{
		IPAMConfig: &network.EndpointIPAMConfig{
			IPv4Address: cmd.IP,
		},
		MacAddress: macAddress,
	})
	return ds.errorWhitelistHandler(err,
		"is already attached to network",
		"Address already in use")
}

func (ds dockerService) DetachNetwork(ctx context.Context, cli entity.DockerCli,
	networkName string, containerName string) entity.Result {

	err := cli.NetworkDisconnect(ctx, networkName, containerName, true)

	return ds.errorWhitelistHandler(err, "is not connected to the network")
}

func (ds dockerService) CreateVolume(ctx context.Context, ecli entity.DockerCli,
	vol command.Volume) entity.Result {

	if !vol.Global {
		volConfig := volume.VolumeCreateBody{
			Labels: vol.Labels,
			Name:   vol.Name,
		}

		_, err := ecli.VolumeCreate(ctx, volConfig)
		return entity.NewResult(err)
	}
	volConfig := volume.VolumeCreateBody{
		Driver: ds.conf.GlusterDriver,
		Name:   vol.Name,
		DriverOpts: map[string]string{
			"glusteropts": fmt.Sprintf("--volfile-server=localhost --volfile-id=/%s", vol.Name),
		},
	}

	clients := make([]entity.Client, len(vol.Hosts))

	for i, host := range vol.Hosts {
		cli, err := ds.CreateClient(host)
		if err != nil {
			return entity.NewErrorResult(err)
		}
		clients[i] = cli
		ds.withField(ecli, "host", host).Info("created a client for volume share")
	}

	errChan := make(chan error)

	brickDir := fmt.Sprintf("/var/bricks/%s", vol.Name)
	for i := range vol.Hosts {
		go func(i int) { //create the directory for the gluster bricks
			errChan <- ds.repo.Execd(ctx, clients[i], GlusterContainerName, []string{"mkdir", "-p",
				brickDir}, true)
		}(i)
	}

	for range vol.Hosts {
		err := <-errChan
		if err != nil {
			return entity.NewErrorResult(err)
		}
	}

	cmds := []string{"gluster", "volume", "create", vol.Name, "replica", fmt.Sprint(len(vol.Hosts))}
	for _, host := range vol.Hosts {
		cmds = append(cmds, fmt.Sprintf("%s:%s", host, brickDir))
	}
	cmds = append(cmds, "force") //needed because it wants a separate partition by default

	err := ds.repo.Execd(ctx, clients[0], GlusterContainerName, cmds, true) //create the replica volume
	if err != nil {
		return entity.NewErrorResult(err)
	}
	time.Sleep(1 * time.Second)
	err = ds.repo.Execd(ctx, clients[0], GlusterContainerName, []string{"gluster", "volume", "start", vol.Name}, true)
	if err != nil {
		return entity.NewErrorResult(err)
	}
	time.Sleep(1 * time.Second)
	err = ds.repo.Execd(ctx, clients[0], GlusterContainerName, []string{"gluster", "volume",
		"set", vol.Name, "ctime", "off"}, true) //compatibility
	if err != nil {
		return entity.NewErrorResult(err)
	}

	/*err = ds.repo.Execd(ctx, clients[0], GlusterContainerName, []string{"gluster", "volume",
		"set", vol.Name, "auth.allow", strings.Join(vol.Hosts, ",") + ",127.0.0.1"}, true) // restrict access by ip
	if err != nil {
		return entity.NewErrorResult(err)
	}*/

	_, err = ecli.VolumeCreate(ctx, volConfig)
	return entity.NewResult(err)
}

func (ds dockerService) RemoveVolume(ctx context.Context, cli entity.DockerCli,
	name string) entity.Result {

	return entity.NewResult(cli.VolumeRemove(ctx, name, true))
}

func (ds dockerService) PlaceFileInContainer(ctx context.Context, cli entity.DockerCli,
	containerName string, file command.File) entity.Result {

	ds.withFields(cli, logrus.Fields{
		"container": containerName,
		"file":      file,
	}).Debug("copying file to container")
	rdr, err := ds.remote.GetTarReader(cli.Labels[command.DefinitionIDKey], file)
	if err != nil {
		return entity.NewErrorResult(err).InjectMeta(map[string]interface{}{
			"labels": cli.Labels,
		})
	}

	srcInfo := archive.CopyInfo{ //appease the Docker Gods
		Path:   file.Meta.Filename,
		Exists: true,
		IsDir:  false,
	}
	dstPath := file.Destination
	if !srcInfo.IsDir && dstPath[len(dstPath)-1] == '/' {
		dstPath += filepath.Base(file.Meta.Filename)
	}

	// Prepare destination copy info by stat-ing the container path.
	dstInfo := archive.CopyInfo{Path: dstPath}

	dstStat, err := cli.ContainerStatPath(ctx, containerName, dstPath)

	// If the destination is a symbolic link, we should evaluate it.
	if err == nil && dstStat.Mode&os.ModeSymlink != 0 {
		linkTarget := dstStat.LinkTarget
		if !system.IsAbs(linkTarget) {
			// Join with the parent directory.
			dstParent, _ := archive.SplitPathDirEntry(dstPath)
			linkTarget = filepath.Join(dstParent, linkTarget)
		}

		dstInfo.Path = linkTarget
		dstStat, err = cli.ContainerStatPath(ctx, containerName, linkTarget)
	}

	if err == nil {
		dstInfo.Exists, dstInfo.IsDir = true, dstStat.Mode.IsDir()
	}
	ds.withFields(cli, logrus.Fields{
		"srcInfo":   srcInfo,
		"dstInfo":   dstInfo,
		"dstStat":   dstStat,
		"dstPath":   dstPath,
		"container": containerName,
	}).Trace("about to prepare the archive copy")
	dstDir, preparedArchive, err := archive.PrepareArchiveCopy(rdr, srcInfo, dstInfo)
	if err != nil {
		return entity.NewErrorResult(err)
	}
	defer preparedArchive.Close()
	ds.withFields(cli, logrus.Fields{
		"destintion": dstDir,
		"container":  containerName,
	}).Debug("got the destination for the file")

	err = cli.CopyToContainer(ctx, containerName, dstDir, preparedArchive, types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
		CopyUIDGID:                false,
	})

	return entity.NewResult(err).InjectMeta(map[string]interface{}{
		"labels":    cli.Labels,
		"container": containerName,
	})
}

func (ds dockerService) Emulation(ctx context.Context, cli entity.DockerCli,
	netem command.Netconf) entity.Result {

	netemImage := "gaiadocker/iproute2:latest"
	errChan := make(chan error, 1)
	go func() {
		errChan <- ds.repo.EnsureImagePulled(ctx, cli, netemImage, "")
	}()

	net, err := ds.repo.GetNetworkByName(ctx, cli, netem.Network)
	if err != nil {
		return entity.NewErrorResult(err)
	}

	err = <-errChan
	if err != nil {
		return entity.NewErrorResult(err)
	}

	name := netem.Container + "-" + net.ID
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
		NetworkMode: container.NetworkMode(fmt.Sprintf("container:%s", netem.Container)),
		CapAdd:      strslice.StrSlice([]string{"NET_ADMIN"}),
	}

	networkConfig := &network.NetworkingConfig{}

	_, err = cli.ContainerCreate(ctx, config, hostConfig, networkConfig, name)
	if err != nil {
		return entity.NewErrorResult(err)
	}
	return ds.StartContainer(ctx, cli, command.StartContainer{Name: name})
}

func (ds dockerService) SwarmCluster(ctx context.Context, entryCLI entity.DockerCli,
	dswarm command.SetupSwarm) entity.Result {

	if len(dswarm.Hosts) == 0 {
		return ErrNoHost
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
	}
	return entity.NewResult(err)
}

func (ds dockerService) mkConfigs() (*container.Config, *container.HostConfig, *network.NetworkingConfig, string) {
	return &container.Config{
			Hostname:   GlusterContainerName,
			Domainname: GlusterContainerName,
			Image:      ds.conf.GlusterImage,
			Entrypoint: strslice.StrSlice([]string{"glusterd", "--no-daemon"}),
		},
		&container.HostConfig{
			AutoRemove:  true,
			NetworkMode: container.NetworkMode("host"),
			CapAdd:      strslice.StrSlice([]string{"NET_ADMIN", "CAP_SYS_ADMIN"}),
		}, &network.NetworkingConfig{}, GlusterContainerName
}

func (ds dockerService) VolumeShare(ctx context.Context, ecli entity.DockerCli, vs command.VolumeShare) entity.Result {

	if len(vs.Hosts) == 0 {
		return entity.NewFatalResult("given an empty volume share command")
	}

	clients := make([]entity.Client, len(vs.Hosts))

	for i, host := range vs.Hosts {
		cli, err := ds.CreateClient(host)
		if err != nil {
			return entity.NewErrorResult(err)
		}
		clients[i] = cli
		ds.withField(ecli, "host", host).Info("created a client for volume share")
	}
	errChan := make(chan error)

	config, hostConfig, networkConfig, name := ds.mkConfigs()

	for i := range vs.Hosts {
		go func(i int) {
			_, err := clients[i].ContainerCreate(ctx, config, hostConfig, networkConfig, name)
			errChan <- err
		}(i)
	}

	for range vs.Hosts {
		err := <-errChan
		if err != nil {
			return entity.NewErrorResult(err)
		}
	}

	for i := range vs.Hosts {
		go func(i int) {
			errChan <- clients[i].ContainerStart(ctx, GlusterContainerName, types.ContainerStartOptions{})
		}(i)
	}

	for i := range vs.Hosts {
		err := <-errChan
		if err != nil {
			return entity.NewErrorResult(err).InjectMeta(map[string]interface{}{
				"host": vs.Hosts[i],
				"type": "StartContainer",
			})
		}
	}

	for _, host := range vs.Hosts[1:] {
		err := ds.repo.Execd(ctx, clients[0], GlusterContainerName, []string{"gluster", "peer", "probe", host}, true)
		if err != nil {
			return entity.NewErrorResult(err)
		}
	}
	return entity.NewSuccessResult()
}
