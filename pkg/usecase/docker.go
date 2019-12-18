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
	"strings"
	"time"

	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/service"
	"github.com/whiteblock/genesis/pkg/validator"

	"github.com/imdario/mergo"
	"github.com/sirupsen/logrus"
	"github.com/whiteblock/definition/command"
)

//DockerUseCase is the usecase for executing the commands in docker
type DockerUseCase interface {
	// Run is equivalent to Execute, except it generates context based on the given command
	Run(cmd command.Command) entity.Result
	// Execute executes the command with the given context
	Execute(ctx context.Context, cmd command.Command) entity.Result
}

var (
	ErrEmptyFieldName      = entity.NewFatalResult("empty field \"name\"")
	ErrEmptyFieldContainer = entity.NewFatalResult("empty field \"container\"")
	ErrEmptyFieldImage     = entity.NewFatalResult("empty field \"image\"")
	ErrEmptyFieldHosts     = entity.NewFatalResult("empty field \"hosts\"")
	ErrEmptyFieldNetwork   = entity.NewFatalResult("empty field \"network\"")
)

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

func (duc dockerUseCase) withFields(cmd command.Command, fields logrus.Fields) *logrus.Entry {
	fields["command"] = cmd.ID
	return duc.log.WithFields(fields)
}

func (duc dockerUseCase) withField(cmd command.Command, key string, value interface{}) *logrus.Entry {
	return duc.withFields(cmd, logrus.Fields{key: value})
}

// Run is equivalent to Execute, except it generates context based on the given command
func (duc dockerUseCase) Run(cmd command.Command) entity.Result {
	stat, ok := duc.validationCheck(cmd)
	if !ok {
		return stat
	}
	duc.withField(cmd, "command", cmd).Trace("running command")
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancelFn()
	return duc.Execute(ctx, cmd)
}

// Execute executes the command with the given context
func (duc dockerUseCase) Execute(ctx context.Context, cmd command.Command) entity.Result {
	cli, err := duc.service.CreateClient(cmd.Target.IP)
	if err != nil {
		duc.withField(cmd, "dest", cmd.Target.IP).Error("failed to create a client")
		return entity.NewFatalResult(err)
	}
	duc.withField(cmd, "client", cli).Trace("created a client")
	duc.withField(cmd, "type", cmd.Order.Type).Trace("routing a command")
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
	case command.SwarmInit:
		return duc.swarmSetupShim(ctx, cli, cmd)
	case command.Pullimage:
		return duc.pullImageShim(ctx, cli, cmd)
	}
	return entity.NewFatalResult(fmt.Errorf("unknown command type: %s", cmd.Order.Type))
}

func (duc dockerUseCase) validationCheck(cmd command.Command) (result entity.Result, ok bool) {
	ok = false
	if len(cmd.Target.IP) == 0 || cmd.Target.IP == "0.0.0.0" {
		result = entity.NewFatalResult(fmt.Sprintf(`invalid target ip "%s"`, cmd.Target.IP))
		return
	}
	ok = true
	return
}

func (duc dockerUseCase) injectLabels(cli entity.Client, cmd command.Command) entity.DockerCli {
	out := entity.DockerCli{Client: cli, Labels: map[string]string{}}
	out.Labels["testRun"] = cmd.Target.TestnetID
	duc.withField(cmd, "meta", cmd.Meta).Trace("got the meta from the command")
	mergo.Map(&out.Labels, cmd.Meta)

	return out
}

func (duc dockerUseCase) createContainerShim(ctx context.Context, cli entity.Client, cmd command.Command) entity.Result {
	var container command.Container
	err := cmd.ParseOrderPayloadInto(&container)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	err = duc.valid.ValidateContainer(container)
	if err != nil {
		return entity.NewFatalResult(err)
	}

	docker := duc.injectLabels(cli, cmd)
	err = mergo.Map(&docker.Labels, container.Labels)
	if err != nil {
		return entity.NewFatalResult(err)
	}

	docker.Labels["name"] = container.Name
	duc.withField(cmd, "labels", docker.Labels).Trace("got the labels for the container")
	return duc.service.CreateContainer(ctx, docker, container)
}

func (duc dockerUseCase) startContainerShim(ctx context.Context, cli entity.Client, cmd command.Command) entity.Result {
	var sc command.StartContainer
	err := cmd.ParseOrderPayloadInto(&sc)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	if len(sc.Name) == 0 {
		return ErrEmptyFieldName
	}
	return duc.service.StartContainer(ctx, duc.injectLabels(cli, cmd), sc)
}

func (duc dockerUseCase) removeContainerShim(ctx context.Context, cli entity.Client, cmd command.Command) entity.Result {
	var payload command.SimpleName
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	if payload.Name == "" {
		return ErrEmptyFieldName
	}
	return duc.service.RemoveContainer(ctx, duc.injectLabels(cli, cmd), payload.Name)
}

func (duc dockerUseCase) createNetworkShim(ctx context.Context, cli entity.Client, cmd command.Command) entity.Result {
	var net command.Network
	err := cmd.ParseOrderPayloadInto(&net)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	docker := duc.injectLabels(cli, cmd)
	err = mergo.Map(&docker.Labels, net.Labels)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.CreateNetwork(ctx, docker, net)
}

func (duc dockerUseCase) attachNetworkShim(ctx context.Context, cli entity.Client, cmd command.Command) entity.Result {
	var payload command.ContainerNetwork
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewErrorResult(err)
	}
	if len(payload.ContainerName) == 0 {
		return ErrEmptyFieldContainer
	}
	if len(payload.Network) == 0 {
		return ErrEmptyFieldNetwork
	}
	return duc.service.AttachNetwork(ctx, duc.injectLabels(cli, cmd), payload.Network, payload.ContainerName)
}

func (duc dockerUseCase) detachNetworkShim(ctx context.Context,
	cli entity.Client, cmd command.Command) entity.Result {

	var payload command.ContainerNetwork
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewErrorResult(err)
	}
	if len(payload.ContainerName) == 0 {
		return ErrEmptyFieldContainer
	}
	if len(payload.Network) == 0 {
		return ErrEmptyFieldNetwork
	}
	return duc.service.DetachNetwork(ctx, duc.injectLabels(cli, cmd),
		payload.Network, payload.ContainerName)
}

func (duc dockerUseCase) removeNetworkShim(ctx context.Context, cli entity.Client, cmd command.Command) entity.Result {
	var payload command.SimpleName
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	if payload.Name == "" {
		return ErrEmptyFieldName
	}
	return duc.service.RemoveNetwork(ctx, duc.injectLabels(cli, cmd), payload.Name)
}
func (duc dockerUseCase) createVolumeShim(ctx context.Context, cli entity.Client, cmd command.Command) entity.Result {
	var payload command.Volume
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	docker := duc.injectLabels(cli, cmd)
	err = mergo.Map(&docker.Labels, payload.Labels)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.CreateVolume(ctx, docker, payload)
}

func (duc dockerUseCase) removeVolumeShim(ctx context.Context, cli entity.Client, cmd command.Command) entity.Result {
	var payload command.SimpleName
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	if payload.Name == "" {
		return ErrEmptyFieldName
	}
	return duc.service.RemoveVolume(ctx, duc.injectLabels(cli, cmd), payload.Name)
}
func (duc dockerUseCase) putFileInContainerShim(ctx context.Context, cli entity.Client, cmd command.Command) entity.Result {
	var payload command.FileAndContainer
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	if len(payload.ContainerName) == 0 {
		return ErrEmptyFieldContainer
	}
	return duc.service.PlaceFileInContainer(ctx, duc.injectLabels(cli, cmd),
		payload.ContainerName, payload.File)
}

func (duc dockerUseCase) emulationShim(ctx context.Context, cli entity.Client, cmd command.Command) entity.Result {
	var payload command.Netconf
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	return duc.service.Emulation(ctx, duc.injectLabels(cli, cmd), payload)
}

func (duc dockerUseCase) swarmSetupShim(ctx context.Context, cli entity.Client, cmd command.Command) entity.Result {
	var payload command.SetupSwarm
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	if len(payload.Hosts) == 0 {
		return ErrEmptyFieldHosts
	}
	return duc.service.SwarmCluster(ctx, duc.injectLabels(cli, cmd), payload)
}

func (duc dockerUseCase) pullImageShim(ctx context.Context, cli entity.Client, cmd command.Command) entity.Result {
	var payload command.PullImage
	err := cmd.ParseOrderPayloadInto(&payload)
	if err != nil {
		return entity.NewFatalResult(err)
	}
	if len(payload.Image) == 0 {
		return ErrEmptyFieldImage
	}
	return duc.service.PullImage(ctx, duc.injectLabels(cli, cmd), payload)
}
