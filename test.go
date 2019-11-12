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
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/repository"
	"github.com/whiteblock/genesis/pkg/service"
	"github.com/whiteblock/genesis/pkg/usecase"
	"time"
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

func createNetwork(dockerUseCase usecase.DockerUseCase) {
	testNetwork := entity.Network{
		Name:   "testnet",
		Global: true,
		Labels: map[string]string{
			"FOO": "BAR",
		},
		Gateway: "10.14.0.1",
		Subnet:  "10.14.0.0/16",
	}
	cmd := mintCommand(testNetwork, "createNetwork")
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("created a network")

}

func removeNetwork(dockerUseCase usecase.DockerUseCase) {
	cmd := mintCommand(map[string]string{"name": "testnet"}, "removeNetwork")
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("removed a network")

}

func createContainer(dockerUseCase usecase.DockerUseCase) {
	testContainer := entity.Container{
		BoundCPUs:  nil, //TODO
		Detach:     false,
		EntryPoint: "/bin/sh",
		Environment: map[string]string{
			"FOO": "BAR",
		},
		Labels: map[string]string{
			"FOO": "BAR",
		},
		Name:    "tester",
		Network: []string{"testnet"}, //TODO
		Ports:   map[int]int{8888: 8889},
		Volumes: map[string]entity.Volume{}, //TODO
		Image:   "alpine",
		Args:    []string{"test", "test2"},
	}
	testContainer.Cpus = "2.5"
	testContainer.Memory = "5gb"
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

func dockerTest() {
	commandService := service.NewCommandService(repository.NewLocalCommandRepository())
	dockerService, err := service.NewDockerService(repository.NewDockerRepository())
	if err != nil {
		panic(err)
	}
	dockerConfig := conf.GetDockerConfig()
	dockerUseCase, err := usecase.NewDockerUseCase(dockerConfig, dockerService, commandService)
	if err != nil {
		panic(err)
	}
	createNetwork(dockerUseCase)
	removeNetwork(dockerUseCase)
	createNetwork(dockerUseCase)
	createContainer(dockerUseCase)
	startContainer(dockerUseCase)
	createVolume(dockerUseCase)
}
