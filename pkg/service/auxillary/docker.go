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

package auxillary

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	//log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/pkg/repository"
	"io/ioutil"
	"strings"
)

//DockerAuxillary provides extra functions for docker service, which could be placed inside of docker
//service, but would make the testing more difficult
type DockerAuxillary interface {
	//EnsureImagePulled checks if the docker host contains an image and pulls it if it does not
	EnsureImagePulled(ctx context.Context, cli *client.Client, imageName string) error

	//GetContainerByName attempts to find a container with the given name and return information on it.
	GetContainerByName(ctx context.Context, cli *client.Client, containerName string) (types.Container, error)

	//GetNetworkByName attempts to find a network with the given name and return information on it.
	GetNetworkByName(ctx context.Context, cli *client.Client, networkName string) (types.NetworkResource, error)

	//GetVolumeByName attempts to find a volume with the given name and return information on it.
	GetVolumeByName(ctx context.Context, cli *client.Client, volumeName string) (*types.Volume, error)

	//HostHasImage returns true if the docker host has an image matching what was given
	HostHasImage(ctx context.Context, cli *client.Client, image string) (bool, error)
}

type dockerAuxillary struct {
	repo repository.DockerRepository
}

//NewDockerAuxillary creates a new DockerAuxillary instance
func NewDockerAuxillary(repo repository.DockerRepository) DockerAuxillary {
	return &dockerAuxillary{repo: repo}
}

//HostHasImage returns true if the docker host has an image matching what was given
func (da dockerAuxillary) HostHasImage(ctx context.Context, cli *client.Client, image string) (bool, error) {
	imgs, err := da.repo.ImageList(ctx, cli, types.ImageListOptions{All: false})
	if err != nil {
		return false, err
	}
	for _, img := range imgs {
		for _, tag := range img.RepoTags {
			if tag == image {
				return true, nil
			}
		}
		for _, digest := range img.RepoDigests {
			if digest == image {
				return true, nil
			}
		}
	}
	return false, nil
}

//EnsureImagePulled checks if the docker host contains an image and pulls it if it does not
func (da dockerAuxillary) EnsureImagePulled(ctx context.Context, cli *client.Client, imageName string) error {
	exists, err := da.HostHasImage(ctx, cli, imageName)
	if exists || err != nil {
		return err
	}

	rd, err := da.repo.ImagePull(ctx, cli, imageName, types.ImagePullOptions{
		Platform: "Linux", //TODO: pull out to a config
	})
	if err != nil {
		return err
	}

	response, err := da.repo.ImageLoad(ctx, cli, rd, true)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = ioutil.ReadAll(response.Body) //It might get stuck here...
	return err
}

//GetNetworkByName attempts to find a network with the given name and return information on it.
func (da dockerAuxillary) GetNetworkByName(ctx context.Context, cli *client.Client,
	networkName string) (types.NetworkResource, error) {
	nets, err := da.repo.NetworkList(ctx, cli, types.NetworkListOptions{})
	if err != nil {
		return types.NetworkResource{}, err
	}
	for _, net := range nets {
		if net.Name == networkName {
			return net, nil
		}
	}
	return types.NetworkResource{}, fmt.Errorf("could not find the network \"%s\"", networkName)
}

//GetContainerByName attempts to find a container with the given name and return information on it.
func (da dockerAuxillary) GetContainerByName(ctx context.Context, cli *client.Client,
	containerName string) (types.Container, error) {

	cntrs, err := da.repo.ContainerList(ctx, cli, types.ContainerListOptions{})
	if err != nil {
		return types.Container{}, err
	}
	for _, cntr := range cntrs {
		for _, name := range cntr.Names {
			if strings.Trim(name, "/") == strings.Trim(containerName, "/") {
				return cntr, nil
			}
		}
	}
	return types.Container{}, fmt.Errorf("could not find the container \"%s\"", containerName)
}

//GetVolumeByName attempts to find a volume with the given name and return information on it.
func (da dockerAuxillary) GetVolumeByName(ctx context.Context, cli *client.Client, volumeName string) (*types.Volume, error) {
	bdy, err := da.repo.VolumeList(ctx, cli, filters.Args{})
	if err != nil {
		return nil, err
	}

	for _, vol := range bdy.Volumes {
		if vol == nil {
			continue
		}
		if vol.Name == volumeName {
			return vol, nil
		}
	}
	return nil, fmt.Errorf("could not find the volume \"%s\"", volumeName)
}
