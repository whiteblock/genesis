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
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/pkg/config"
	"github.com/whiteblock/genesis/pkg/controller"
	"github.com/whiteblock/genesis/pkg/handler"
	handAux "github.com/whiteblock/genesis/pkg/handler/auxillary"
	"github.com/whiteblock/genesis/pkg/repository"
	"github.com/whiteblock/genesis/pkg/service"
	"github.com/whiteblock/genesis/pkg/service/auxillary"
	"github.com/whiteblock/genesis/pkg/usecase"
	"github.com/whiteblock/genesis/pkg/utility"
	"os"
)

func getRestServer() (controller.RestController, error) {
	conf, err := config.NewConfig()
	if err != nil {
		return nil, err
	}

	/*logger,err := conf.GetLogger()
	if err != nil {
		return nil, err
	}*/
	dockerRepository := repository.NewDockerRepository()

	return controller.NewRestController(
		conf.GetRestConfig(),
		handler.NewRestHandler(
			usecase.NewDockerUseCase(
				service.NewDockerService(
					dockerRepository,
					auxillary.NewDockerAuxillary(dockerRepository),
					conf.GetDockerConfig()))),
		mux.NewRouter()), nil
}

func getCommandController() (controller.CommandController, error) {
	conf, err := config.NewConfig()
	if err != nil {
		return nil, err
	}

	logger, err := conf.GetLogger()
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

	cmdConn, err := utility.OpenAMQPConnection(cmdConf.Endpoint)
	if err != nil {
		return nil, err
	}

	complConn, err := utility.OpenAMQPConnection(complConf.Endpoint)
	if err != nil {
		return nil, err
	}

	return controller.NewCommandController(
		conf.QueueMaxConcurrency,
		service.NewAMQPService(cmdConf, repository.NewAMQPRepository(cmdConn)),
		service.NewAMQPService(complConf, repository.NewAMQPRepository(complConn)),
		handler.NewDeliveryHandler(
			handAux.NewExecutor(
				usecase.NewDockerUseCase(
					service.NewDockerService(
						repository.NewDockerRepository(),
						auxillary.NewDockerAuxillary(
							repository.NewDockerRepository()),
						conf.GetDockerConfig())),
				logger),
			utility.NewAMQPMessage(conf.MaxMessageRetries),
			logger),
		logger)
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
	log.Info("starting the rest server")
	restServer.Start()
}
