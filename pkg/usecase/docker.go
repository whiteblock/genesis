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

package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/service"
	"github.com/whiteblock/genesis/util"
)

type DockerUseCase interface {
	Execute(ctx context.Context, cmd command.Command) entity.Result
}

type dockerUseCase struct {
	conf    entity.DockerConfig
	service service.DockerService
}

func NewDockerUseCase(conf entity.DockerConfig, service service.DockerService) (DockerUseCase, error) {
	return dockerUseCase{conf: conf, service: service}, nil
}

func (duck dockerUseCase) Execute(ctx context.Context, cmd command.Command) entity.Result {
	cli, err := client.NewClientWithOpts(
		client.WithAPIVersionNegotiation(),
		client.WithHost(cmd.Target.IP),
		client.WithTLSClientConfig(duck.conf.CACertPath, duck.conf.CertPath, duck.conf.KeyPath),
	)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	log.WithField("client", cli).Debug("created a client")
	switch cmd.Order.Type {
	case "createContainer":
		return duck.createContainerShim(ctx, cli, cmd)
	case "startContainer":
		return duck.startContainerShim(ctx, cli, cmd)
	case "removeContainer":
		return duck.removeContainerShim(ctx, cli, cmd)
	case "createNetwork":
		return duck.createNetworkShim(ctx, cli, cmd)
	case "attachNetwork":
		return duck.attachNetworkShim(ctx, cli, cmd)
	case "createVolume":
		return duck.removeVolumeShim(ctx, cli, cmd)
	case "removeVolume":
		return duck.removeVolumeShim(ctx, cli, cmd)
	case "putFile":
		return duck.putFileShim(ctx, cli, cmd)
	case "putFileInContainer":
		return duck.putFileInContainerShim(ctx, cli, cmd)
	case "emulation":
		return duck.emulationShim(ctx, cli, cmd)
	}
	return entity.NewFatalResult(fmt.Errorf("unknown command type: %s", cmd.Order.Type))
}

func (duc dockerUseCase) createContainerShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	raw, err := json.Marshal(cmd.Order.Payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	var container entity.Container
	err = json.Unmarshal(raw, &container)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.CreateContainer(ctx, cli, container)
}

func (duc dockerUseCase) startContainerShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	iName, exists := cmd.Order.Payload["name"]
	if !exists {
		return entity.NewFatalResult(fmt.Errorf("missing field \"name\""))
	}
	name, isString := iName.(string)
	if !isString {
		return entity.NewFatalResult(fmt.Errorf("field \"name\" is expected to be a string"))
	}
	return duc.service.StartContainer(ctx, cli, name)
}

func (duc dockerUseCase) removeContainerShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	var name string
	err := util.GetJSONString(cmd.Order.Payload, "name", &name)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.RemoveContainer(ctx, cli, name)
}

func (duc dockerUseCase) createNetworkShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	raw, err := json.Marshal(cmd.Order.Payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	var net entity.Network
	err = json.Unmarshal(raw, &net)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.CreateNetwork(ctx, cli, net)
}

func (duc dockerUseCase) attachNetworkShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	var networkName string
	var containerName string
	err := util.GetJSONString(cmd.Order.Payload, "network", &networkName)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	err = util.GetJSONString(cmd.Order.Payload, "container", &containerName)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.AttachNetwork(ctx, cli, networkName, containerName)
}

func (duc dockerUseCase) createVolumeShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	raw, err := json.Marshal(cmd.Order.Payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	var volume entity.Volume
	err = json.Unmarshal(raw, &volume)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.CreateVolume(ctx, cli, volume)
}

func (duc dockerUseCase) removeVolumeShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	var name string
	err := util.GetJSONString(cmd.Order.Payload, "name", &name)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.RemoveVolume(ctx, cli, name)
}

func (duc dockerUseCase) putFileShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	var volumeName string
	err := util.GetJSONString(cmd.Order.Payload, "volume", &volumeName)
	if err != nil {
		return entity.NewFatalResult(err)
	}

	_, hasField := cmd.Order.Payload["file"]
	if !hasField {
		return entity.NewFatalResult(fmt.Errorf("missing file field"))
	}

	raw, err := json.Marshal(cmd.Order.Payload["file"])
	if err != nil {
		return entity.NewFatalResult(err)
	}
	var file entity.File
	err = json.Unmarshal(raw, &file)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.PlaceFileInVolume(ctx, cli, volumeName, file)
}

func (duc dockerUseCase) putFileInContainerShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	var containerName string
	err := util.GetJSONString(cmd.Order.Payload, "container", &containerName)
	if err != nil {
		return entity.NewFatalResult(err)
	}

	_, hasField := cmd.Order.Payload["file"]
	if !hasField {
		return entity.NewFatalResult(fmt.Errorf("missing file field"))
	}

	raw, err := json.Marshal(cmd.Order.Payload["file"])
	if err != nil {
		return entity.NewFatalResult(err)
	}
	var file entity.File
	err = json.Unmarshal(raw, &file)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.PlaceFileInContainer(ctx, cli, containerName, file)
}

func (duc dockerUseCase) emulationShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	raw, err := json.Marshal(cmd.Order.Payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	var netem entity.Netconf
	err = json.Unmarshal(raw, &netem)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.Emulation(ctx, cli, netem)
}
