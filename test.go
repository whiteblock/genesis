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

package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/whiteblock/genesis/pkg/config"
	"github.com/whiteblock/genesis/pkg/file"
	"github.com/whiteblock/genesis/pkg/repository"
	"github.com/whiteblock/genesis/pkg/service"
	"github.com/whiteblock/genesis/pkg/usecase"

	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/definition/command"
)

/*FUNCTIONALITY TESTS*/
/*NOTE: this should be replaced with an integration test*/

func mintCommand(i interface{}, orderType command.OrderType) command.Command {
	raw, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	cmd := command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "127.0.0.1"},
		Order: command.Order{
			Type: orderType,
		},
		Meta: map[string]string{
			"org": "543",
		},
	}
	err = json.Unmarshal(raw, &cmd.Order.Payload)
	if err != nil {
		panic(err)
	}
	return cmd
}

func createVolume(dockerUseCase usecase.DockerUseCase, name string) {
	vol := command.Volume{
		Name: name,
		Labels: map[string]string{
			"FOO": "BAR",
		},
	}

	cmd := mintCommand(vol, command.Createvolume)
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("created a volume")
}

func removeVolume(dockerUseCase usecase.DockerUseCase, name string) {
	cmd := mintCommand(map[string]string{
		"name": name,
	}, command.Removevolume)
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("removed a volume")
}

func removeContainer(dockerUseCase usecase.DockerUseCase, name string) {
	cmd := mintCommand(command.SimpleName{
		Name: name,
	}, command.Removecontainer)
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("removed a container")
}

func createNetwork(dockerUseCase usecase.DockerUseCase, name string, num int) {
	testNetwork := command.Network{
		Name:   name,
		Global: true,
		Labels: map[string]string{
			"FOO": "BAR",
		},
		Gateway: fmt.Sprintf("10.%d.0.1", num),
		Subnet:  fmt.Sprintf("10.%d.0.0/16", num),
	}
	cmd := mintCommand(testNetwork, command.Createnetwork)
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("created a network")
}

func attachNetwork(dockerUseCase usecase.DockerUseCase,
	networkName string, containerName string, ip string) {

	cmd := mintCommand(command.ContainerNetwork{
		ContainerName: containerName,
		Network:       networkName,
		IP:            ip,
	}, command.Attachnetwork)
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("attached a network")
}

func detachNetwork(dockerUseCase usecase.DockerUseCase, networkName string, containerName string) {
	cmd := mintCommand(map[string]string{
		"container": "tester",
		"network":   networkName,
	}, command.Detachnetwork)
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("detached a network")
}

func removeNetwork(dockerUseCase usecase.DockerUseCase, name string) {
	cmd := mintCommand(map[string]string{"name": name}, command.Removenetwork)
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("removed a network")
}

func pullImage(dockerUseCase usecase.DockerUseCase, image string) {
	cmd := mintCommand(command.PullImage{
		Image: image,
	}, command.Pullimage)

	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("pulled an image")
}

func createContainer(dockerUseCase usecase.DockerUseCase, name string,
	args []string, ports map[int]int) {
	testContainer := command.Container{
		BoundCPUs: nil, //TODO
		Environment: map[string]string{
			"FOO": "BAR",
		},
		Labels: map[string]string{
			"FOO": "BAR",
		},
		Name:    name,
		Network: "testnet",
		Ports:   ports,
		Volumes: []command.Mount{command.Mount{
			Name:      "test_volume",
			Directory: "/foo/bar",
			ReadOnly:  false,
		}},
		Image:      "nettools/ubuntools",
		EntryPoint: "ping",
		Args:       args,
	}
	testContainer.Cpus = "1"
	testContainer.Memory = "1gb"
	cmd := mintCommand(testContainer, "createContainer")
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("created a container")
}

func startContainer(dockerUseCase usecase.DockerUseCase, name string, attach bool) {
	cmd := mintCommand(map[string]interface{}{
		"name":    name,
		"attach":  attach,
		"timeout": 3 * time.Minute,
	}, command.Startcontainer)
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("started a container")
}

func emulate(dockerUseCase usecase.DockerUseCase, containerName string, networkName string) {
	cmd := mintCommand(command.Netconf{
		Container: containerName,
		Network:   networkName,
		Delay:     100000,
	}, command.Emulation)
	res := dockerUseCase.Run(cmd)
	log.WithFields(log.Fields{"res": res}).Info("applied emulation")
}

func dockerTest(clean bool) {
	conf, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	lvl, err := log.ParseLevel(conf.Verbosity)
	if err != nil {
		panic(err)
	}
	log.SetLevel(lvl)

	dockerUseCase := usecase.NewDockerUseCase(
		service.NewDockerService(
			repository.NewDockerRepository(),
			conf.Docker,
			file.NewRemoteSources(
				conf.FileHandler,
				conf.GetLogger()),
			conf.GetLogger()),
		conf.GetLogger())

	if clean {
		removeContainer(dockerUseCase, "tester")
		removeContainer(dockerUseCase, "tester2")
		removeContainer(dockerUseCase, "tester3")
		removeVolume(dockerUseCase, "test_volume")
		time.Sleep(2 * time.Second)
		removeNetwork(dockerUseCase, "testnet")
		removeNetwork(dockerUseCase, "testnet2")
		return
	}
	pullImage(dockerUseCase, "nettools/ubuntools")

	createVolume(dockerUseCase, "test_volume")
	createNetwork(dockerUseCase, "testnet", 14)
	createContainer(dockerUseCase, "tester", []string{"localhost"}, map[int]int{
		8765: 8755,
	})
	startContainer(dockerUseCase, "tester", false)
	createContainer(dockerUseCase, "tester2", []string{"localhost"}, nil)
	startContainer(dockerUseCase, "tester2", false)

	createContainer(dockerUseCase, "tester3", []string{"-c", "30", "localhost"}, nil)
	startContainer(dockerUseCase, "tester3", true)

	createNetwork(dockerUseCase, "testnet2", 15)
	attachNetwork(dockerUseCase, "testnet2", "tester", "10.15.0.2")
	attachNetwork(dockerUseCase, "testnet2", "tester2", "10.15.0.3")
	detachNetwork(dockerUseCase, "testnet", "tester")
	emulate(dockerUseCase, "tester", "testnet2")
	emulate(dockerUseCase, "tester2", "testnet2")
}
