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
	"errors"
	"fmt"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/service"
	"strings"
	"time"
)

const (
	numberOfRetries = 4
	waitBeforeRetry = 10
)

var (
	statusTooSoon = entity.Result{Type: entity.TooSoonType, Error: fmt.Errorf("command ran too soon")}
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
	service    service.DockerService
	cmdService service.CommandService
}

//NewDockerUseCase creates a DockerUseCase arguments given the proper dep injections
func NewDockerUseCase(service service.DockerService,
	cmdService service.CommandService) (DockerUseCase, error) {
	return &dockerUseCase{service: service, cmdService: cmdService}, nil
}

// TimeSupplier supplies the time as a unix timestamp
func (duc dockerUseCase) TimeSupplier() int64 {
	return time.Now().Unix()
}

// Run is equivalent to Execute, except it generates context based on the given command
func (duc dockerUseCase) Run(cmd command.Command) entity.Result {
	stat, ok := duc.dependencyCheck(cmd)
	if !ok {
		return stat
	}
	log.WithField("command", cmd).Trace("running command")
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
	log.WithField("client", cli).Trace("created a client")
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

func (duc dockerUseCase) dependencyCheck(cmd command.Command) (stat entity.Result, ok bool) {
	ok = false
	if duc.TimeSupplier() < cmd.Timestamp {
		stat = statusTooSoon
		return
	}

	ready, err := duc.cmdService.CheckDependenciesExecuted(cmd)
	if err != nil {
		stat = entity.NewErrorResult(fmt.Errorf("error checking dependencies"))
		return
	}

	if !ready {
		stat = statusTooSoon
		return
	}
	ok = true
	return
}

func (duc dockerUseCase) createContainerShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	raw, err := json.Marshal(cmd.Order.Payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	var container command.Container
	err = json.Unmarshal(raw, &container)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.CreateContainer(ctx, cli, container)
}

func (duc dockerUseCase) startContainerShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	raw, err := json.Marshal(cmd.Order.Payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	var sc command.StartContainer
	err = json.Unmarshal(raw, &sc)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	if len(sc.Name) == 0 {
		return entity.NewFatalResult(fmt.Errorf("empty field \"name\""))
	}
	return duc.service.StartContainer(ctx, cli, sc)
}

func (duc dockerUseCase) removeContainerShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	if payload, ok := cmd.Order.Payload.(command.SimpleName); ok {
		if payload.Name == "" {
			return entity.NewFatalResult(fmt.Errorf("empty field \"name\""))
		}
		return duc.service.RemoveContainer(ctx, cli, payload.Name)
	}
	return entity.NewFatalResult(fmt.Errorf("missing field \"name\""))
}

func (duc dockerUseCase) createNetworkShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	raw, err := json.Marshal(cmd.Order.Payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	var net command.Network
	err = json.Unmarshal(raw, &net)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.CreateNetwork(ctx, cli, net)
}

func (duc dockerUseCase) attachNetworkShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	if payload, ok := cmd.Order.Payload.(command.ContainerNetwork); ok {
		if payload.ContainerName == "" {
			return entity.NewFatalResult(fmt.Errorf("empty field \"container\""))
		}
		if payload.Network == "" {
			return entity.NewFatalResult(fmt.Errorf("empty field \"network\""))
		}
		return duc.service.AttachNetwork(ctx, cli, payload.Network, payload.ContainerName)
	}
	return entity.NewFatalResult(errors.New("Invalid payload"))
}

func (duc dockerUseCase) detachNetworkShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	if payload, ok := cmd.Order.Payload.(command.ContainerNetwork); ok {
		return duc.service.DetachNetwork(ctx, cli, payload.Network, payload.ContainerName)
	}
	return entity.NewFatalResult(errors.New("Invalid payload"))
}

func (duc dockerUseCase) removeNetworkShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	if payload, ok := cmd.Order.Payload.(command.SimpleName); ok {
		if payload.Name == "" {
			return entity.NewFatalResult(fmt.Errorf("empty field \"name\""))
		}
		return duc.service.RemoveNetwork(ctx, cli, payload.Name)
	}
	return entity.NewFatalResult(fmt.Errorf("missing field \"name\""))
}
func (duc dockerUseCase) createVolumeShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	if payload, ok := cmd.Order.Payload.(command.Volume); ok {
		return duc.service.CreateVolume(ctx, cli, payload)
	}
	return entity.NewFatalResult(fmt.Errorf("Invalid payload"))
}

func (duc dockerUseCase) removeVolumeShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	if payload, ok := cmd.Order.Payload.(command.SimpleName); ok {
		if payload.Name == "" {
			return entity.NewFatalResult(fmt.Errorf("empty field \"name\""))
		}
		return duc.service.RemoveVolume(ctx, cli, payload.Name)
	}
	return entity.NewFatalResult(fmt.Errorf("missing field \"name\""))
}
func (duc dockerUseCase) putFileInContainerShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	if payload, ok := cmd.Order.Payload.(command.FileAndContainer); ok {
		return duc.service.PlaceFileInContainer(ctx, cli, payload.ContainerName, payload.File)
	}
	return entity.NewFatalResult(errors.New("Invalid payload"))
}

func (duc dockerUseCase) emulationShim(ctx context.Context, cli *client.Client, cmd command.Command) entity.Result {
	if payload, ok := cmd.Order.Payload.(command.Netconf); ok {
		return duc.service.Emulation(ctx, cli, payload)
	}
	return entity.NewFatalResult(errors.New("Invalid payload"))
}
