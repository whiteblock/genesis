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

package main

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/repository"
	"github.com/whiteblock/genesis/pkg/service"
	"github.com/whiteblock/genesis/pkg/service/auxillary"
	"github.com/whiteblock/genesis/pkg/usecase"
)

/*FUNCTIONALITY TESTS*/
/*NOTE: this should be replaced with an integration test*/

func mintCommand(i interface{}, orderType string) command.Command {
	raw, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	cmd := command.Command{
		ID:        "TEST",
		Timestamp: time.Now().Unix() - 5,
		Timeout:   0,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type: orderType,
		},
	}
	err = json.Unmarshal(raw, &cmd.Order.Payload)
	if err != nil {
		panic(err)
	}
	return cmd
}

func createVolume(dockerUseCase usecase.DockerUseCase) {
	vol := entity.Volume{
		Name: "test_volume",
		Labels: map[string]string{
			"FOO": "BAR",
		},
	}

	cmd := mintCommand(vol, "createVolume")
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("created a volume")
}

func removeVolume(dockerUseCase usecase.DockerUseCase) {
	cmd := mintCommand(map[string]string{
		"name": "test_volume",
	}, "removeVolume")
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("removed a volume")
}

func removeContainer(dockerUseCase usecase.DockerUseCase) {
	cmd := mintCommand(map[string]string{
		"name": "tester",
	}, "removeContainer")
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("removed a container")
}

func createNetwork(dockerUseCase usecase.DockerUseCase, name string, num int) {
	testNetwork := entity.Network{
		Name:   name,
		Global: true,
		Labels: map[string]string{
			"FOO": "BAR",
		},
		Gateway: fmt.Sprintf("10.%d.0.1", num),
		Subnet:  fmt.Sprintf("10.%d.0.0/16", num),
	}
	cmd := mintCommand(testNetwork, "createNetwork")
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("created a network")
}

func attachNetwork(dockerUseCase usecase.DockerUseCase, networkName string, containerName string) {
	cmd := mintCommand(map[string]string{
		"container": "tester",
		"network":   networkName,
	}, "attachnetwork")
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("attached a network")
}

func detachNetwork(dockerUseCase usecase.DockerUseCase, networkName string, containerName string) {
	cmd := mintCommand(map[string]string{
		"container": "tester",
		"network":   networkName,
	}, "detachnetwork")
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("detached a network")
}

func removeNetwork(dockerUseCase usecase.DockerUseCase, name string) {
	cmd := mintCommand(map[string]string{"name": name}, "removeNetwork")
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("removed a network")
}

func createContainer(dockerUseCase usecase.DockerUseCase) {
	testContainer := entity.Container{
		BoundCPUs: nil, //TODO
		Detach:    false,
		Environment: map[string]string{
			"FOO": "BAR",
		},
		Labels: map[string]string{
			"FOO": "BAR",
		},
		Name:    "tester",
		Network: []string{"testnet"},
		Ports:   map[int]int{8888: 8889},
		Volumes: []entity.Mount{}, //TODO
		Image:   "nginx:latest",
	}
	testContainer.Cpus = "1"
	testContainer.Memory = "1gb"
	cmd := mintCommand(testContainer, "createContainer")
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("created a container")
}

func startContainer(dockerUseCase usecase.DockerUseCase) {
	cmd := mintCommand(map[string]interface{}{
		"name": "tester",
	}, "startContainer")
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("started a container")
}

func dockerTest(clean bool) {
	commandService := service.NewCommandService(repository.NewLocalCommandRepository())
	dockerRepository := repository.NewDockerRepository()
	dockerAux := auxillary.NewDockerAuxillary(dockerRepository)
	dockerService, err := service.NewDockerService(dockerRepository, dockerAux)
	if err != nil {
		panic(err)
	}

	dockerConfig := conf.GetDockerConfig()
	dockerUseCase, err := usecase.NewDockerUseCase(dockerConfig, dockerService, commandService)
	if err != nil {
		panic(err)
	}
	if clean {
		removeContainer(dockerUseCase)
		removeVolume(dockerUseCase)
		time.Sleep(2 * time.Second)
		removeNetwork(dockerUseCase, "testnet")
		removeNetwork(dockerUseCase, "testnet2")
		return
	}

	createVolume(dockerUseCase)
	createNetwork(dockerUseCase, "testnet", 14)
	createContainer(dockerUseCase)
	startContainer(dockerUseCase)
	createNetwork(dockerUseCase, "testnet2", 15)
	attachNetwork(dockerUseCase, "testnet2", "tester")
	detachNetwork(dockerUseCase, "testnet", "tester")
}
