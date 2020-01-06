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
	"os"

	"github.com/whiteblock/genesis/pkg/config"
	"github.com/whiteblock/genesis/pkg/controller"
	"github.com/whiteblock/genesis/pkg/file"
	"github.com/whiteblock/genesis/pkg/handler"
	handAux "github.com/whiteblock/genesis/pkg/handler/auxillary"
	"github.com/whiteblock/genesis/pkg/repository"
	"github.com/whiteblock/genesis/pkg/service"
	"github.com/whiteblock/genesis/pkg/usecase"

	"github.com/gorilla/mux"
	queue "github.com/whiteblock/amqp"
)

func getRestServer() (controller.RestController, error) {
	conf, err := config.NewConfig()
	if err != nil {
		return nil, err
	}
	dockerRepository := repository.NewDockerRepository()

	return controller.NewRestController(
		conf.GetRestConfig(),
		handler.NewRestHandler(
			usecase.NewDockerUseCase(
				service.NewDockerService(
					dockerRepository,
					conf.Docker,
					file.NewRemoteSources(
						conf.FileHandler,
						conf.GetLogger()),
					conf.GetLogger()),
				conf.GetLogger()),
			conf.GetLogger()),
		mux.NewRouter(),
		conf.GetLogger()), nil
}

func getCommandController() (controller.CommandController, error) {
	conf, err := config.NewConfig()
	if err != nil {
		return nil, err
	}
	complConf, err := conf.CompletionAMQP()
	if err != nil {
		return nil, err
	}

	cmdConf, err := conf.CommandAMQP()
	if err != nil {
		return nil, err
	}

	errConf, err := conf.ErrorsAMQP()
	if err != nil {
		return nil, err
	}
	
	if complConf.QueueName == cmdConf.QueueName {
		panic("completion queue matches command queue")
	}

	if cmdConf.QueueName == errConf.QueueName {
		panic("command queue matches error queue")
	}

	if errConf.QueueName == complConf.QueueName {
		panic("error queue matches command queue")
	}

	cmdConn, err := queue.OpenAMQPConnection(cmdConf.Endpoint)
	if err != nil {
		return nil, err
	}

	complConn, err := queue.OpenAMQPConnection(complConf.Endpoint)
	if err != nil {
		return nil, err
	}

	errConn, err := queue.OpenAMQPConnection(errConf.Endpoint)
	if err != nil {
		return nil, err
	}

	return controller.NewCommandController(
		conf.QueueMaxConcurrency,
		queue.NewAMQPService(cmdConf, queue.NewAMQPRepository(cmdConn), conf.GetLogger()),
		queue.NewAMQPService(cmdConf, queue.NewAMQPRepository(errConn), conf.GetLogger()),
		queue.NewAMQPService(complConf, queue.NewAMQPRepository(complConn), conf.GetLogger()),
		handler.NewDeliveryHandler(
			handAux.NewExecutor(
				conf.Execution,
				usecase.NewDockerUseCase(
					service.NewDockerService(
						repository.NewDockerRepository(),
						conf.Docker,
						file.NewRemoteSources(
							conf.FileHandler,
							conf.GetLogger()),
						conf.GetLogger()),
					conf.GetLogger()),
				conf.GetLogger()),
			queue.NewAMQPMessage(conf.MaxMessageRetries),
			conf.GetLogger()),
		conf.GetLogger())
}

func main() {

	if len(os.Args) == 2 && os.Args[1] == "test" { //Run some basic docker functionality tests
		dockerTest(false)
		os.Exit(0)
	}

	if len(os.Args) == 2 && os.Args[1] == "clean" { //Clean some basic docker functionality tests
		dockerTest(true)
		os.Exit(0)
	}

	restServer, err := getRestServer()
	if err != nil {
		panic(err)
	}

	conf, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	if !conf.LocalMode {
		cmdCntl, err := getCommandController()
		if err != nil {
			panic(err)
		}
		go cmdCntl.Start()
	}

	conf.GetLogger().Info("starting the rest server")
	restServer.Start()
}
