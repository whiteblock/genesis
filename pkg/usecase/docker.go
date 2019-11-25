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
	"fmt"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/service"
	"github.com/whiteblock/genesis/pkg/validator"
	"strings"
	"time"
)

//DockerUseCase is the usecase for executing the commands in docker
type DockerUseCase interface {
	// Run is equivalent to Execute, except it generates context based on the given command
	Run(cmd command.Command) entity.Result
	// TimeSupplier supplies the time as a unix timestamp
	TimeSupplier() int64
	// Execute executes the command with the given context
	Execute(ctx context.Context, cmd command.Command) entity.Result
}

type dockerUseCase struct {
	valid   validator.OrderValidator
	service service.DockerService
	log     logrus.Ext1FieldLogger
}

//NewDockerUseCase creates a DockerUseCase arguments given the proper dep injections
func NewDockerUseCase(
	service service.DockerService,
	valid validator.OrderValidator,
	log logrus.Ext1FieldLogger) DockerUseCase {
	return &dockerUseCase{service: service, valid: valid, log: log}
}

// TimeSupplier supplies the time as a unix timestamp
func (duc dockerUseCase) TimeSupplier() int64 {
	return time.Now().Unix()
}

// Run is equivalent to Execute, except it generates context based on the given command
func (duc dockerUseCase) Run(cmd command.Command) entity.Result {
	stat, ok := duc.validationCheck(cmd)
	if !ok {
		return stat
	}
	duc.log.WithField("command", cmd).Trace("running command")
	if cmd.Timeout == 0 {
		return duc.Execute(context.Background(), cmd)
	}
	ctx, cancelFn := context.WithTimeout(context.Background(), cmd.Timeout)
	defer cancelFn()
	return duc.Execute(ctx, cmd)
}

// Execute executes the command with the given context
func (duc dockerUseCase) Execute(ctx context.Context, cmd command.Command) entity.Result {
	cli, err := duc.service.CreateClient(cmd.Target.IP)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	duc.log.WithField("client", cli).Trace("created a client")
	switch command.OrderType(strings.ToLower(string(cmd.Order.Type))) {
	case command.Createcontainer:
		return duc.createContainerShim(ctx, cli, cmd)
	case command.Startcontainer:
		return duc.startContainerShim(ctx, cli, cmd)
	case command.Removecontainer:
		return duc.removeContainerShim(ctx, cli, cmd)
	case command.Createnetwork:
		return duc.createNetworkShim(ctx, cli, cmd)
	case command.Attachnetwork:
		return duc.attachNetworkShim(ctx, cli, cmd)
	case command.Detachnetwork:
		return duc.detachNetworkShim(ctx, cli, cmd)
	case command.Removenetwork:
		return duc.removeNetworkShim(ctx, cli, cmd)
	case command.Createvolume:
		return duc.createVolumeShim(ctx, cli, cmd)
	case command.Removevolume:
		return duc.removeVolumeShim(ctx, cli, cmd)
	case command.Putfileincontainer:
		return duc.putFileInContainerShim(ctx, cli, cmd)
	case command.Emulation:
		return duc.emulationShim(ctx, cli, cmd)
	}
	return entity.NewFatalResult(fmt.Errorf("unknown command type: %s", cmd.Order.Type))
}

func (duc dockerUseCase) validationCheck(cmd command.Command) (result entity.Result, ok bool) {
	ok = false
	if len(cmd.Target.IP) == 0 || cmd.Target.IP == "0.0.0.0" {
		result = entity.NewFatalResult("invalid target ip")
		return
	}
	ok = true
	return
}

func (duc dockerUseCase) createContainerShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	var container command.Container
	err := cmd.ParseOrderPayloadInto(&container)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	err = duc.valid.ValidateContainer(container)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.CreateContainer(ctx, cli, container)
}

func (duc dockerUseCase) startContainerShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	var sc command.StartContainer
	err := cmd.ParseOrderPayloadInto(&sc)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	if len(sc.Name) == 0 {
		return entity.NewFatalResult("empty field \"name\"")
	}
	return duc.service.StartContainer(ctx, cli, sc)
}

func (duc dockerUseCase) removeContainerShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	var payload command.SimpleName
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	if payload.Name == "" {
		return entity.NewFatalResult("empty field \"name\"")
	}
	return duc.service.RemoveContainer(ctx, cli, payload.Name)
}

func (duc dockerUseCase) createNetworkShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	var net command.Network
	err := cmd.ParseOrderPayloadInto(&net)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.CreateNetwork(ctx, cli, net)
}

func (duc dockerUseCase) attachNetworkShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	var payload command.ContainerNetwork
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewErrorResult(err)
	}
	if len(payload.ContainerName) == 0 {
		return entity.NewFatalResult(fmt.Errorf("empty field \"container\""))
	}
	if len(payload.Network) == 0 {
		return entity.NewFatalResult(fmt.Errorf("empty field \"network\""))
	}
	return duc.service.AttachNetwork(ctx, cli, payload.Network, payload.ContainerName)
}

func (duc dockerUseCase) detachNetworkShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	var payload command.ContainerNetwork
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewErrorResult(err)
	}
	return duc.service.DetachNetwork(ctx, cli, payload.Network, payload.ContainerName)
}

func (duc dockerUseCase) removeNetworkShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	var payload command.SimpleName
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	if payload.Name == "" {
		return entity.NewFatalResult("empty field \"name\"")
	}
	return duc.service.RemoveNetwork(ctx, cli, payload.Name)
}
func (duc dockerUseCase) createVolumeShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	var payload command.Volume
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.CreateVolume(ctx, cli, payload)
}

func (duc dockerUseCase) removeVolumeShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	var payload command.SimpleName
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	if payload.Name == "" {
		return entity.NewFatalResult("empty field \"name\"")
	}
	return duc.service.RemoveVolume(ctx, cli, payload.Name)
}
func (duc dockerUseCase) putFileInContainerShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	var payload command.FileAndContainer
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	if len(payload.ContainerName) == 0 {
		return entity.NewFatalResult("missing field \"container\"")
	}
	return duc.service.PlaceFileInContainer(ctx, cli, payload.ContainerName, payload.File)
}

func (duc dockerUseCase) emulationShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	var payload command.Netconf
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.Emulation(ctx, cli, payload)
}
