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
		Name:    "test",
		Network: []string{"testnet"}, //TODO
		Ports:   map[int]int{8888: 8889},
		Volumes: map[string]entity.Volume{}, //TODO
		Image:   "alpine",
		Args:    []string{"test", "test2"},
	}
	testContainer.Cpus = "2.5"
	testContainer.Memory = "5gb"
	raw, err := json.Marshal(testContainer)
	if err != nil {
		panic(err)
	}
	cmd := command.Command{
		ID:        "TEST",
		Timestamp: time.Now().Unix() - 5,
		Timeout:   0,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "createContainer",
			Payload: map[string]interface{}{},
		},
	}
	err = json.Unmarshal(raw, &cmd.Order.Payload)
	if err != nil {
		panic(err)
	}

	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("created a container")
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
	createContainer(dockerUseCase)
}
