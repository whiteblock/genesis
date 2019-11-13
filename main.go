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
	"github.com/whiteblock/genesis/config"
	"github.com/whiteblock/genesis/pkg/controller"
	"github.com/whiteblock/genesis/pkg/handler"
	"github.com/whiteblock/genesis/pkg/repository"
	"github.com/whiteblock/genesis/pkg/service"
	"github.com/whiteblock/genesis/pkg/service/auxillary"
	"github.com/whiteblock/genesis/pkg/usecase"
	"os"
)

var conf = config.GetConfig()

func getRestServer() (controller.RestController, error) {
	commandService := service.NewCommandService(repository.NewLocalCommandRepository())
	dockerRepository := repository.NewDockerRepository()
	dockerAux := auxillary.NewDockerAuxillary(dockerRepository)
	dockerService, err := service.NewDockerService(dockerRepository, dockerAux)
	if err != nil {
		return nil, err
	}
	dockerConfig := conf.GetDockerConfig()
	dockerUseCase, err := usecase.NewDockerUseCase(dockerConfig, dockerService, commandService)
	if err != nil {
		return nil, err
	}

	restHandler := handler.NewRestHandler(dockerUseCase, commandService)
	restRouter := mux.NewRouter()
	restConfig := conf.GetRestConfig()
	restServer := controller.NewRestController(restConfig, restHandler, restRouter)
	return restServer, nil
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "test" { //Run some basic docker functionality tests
		dockerTest()
		os.Exit(0)
	}

	restServer, err := getRestServer()
	if err != nil {
		panic(err)
	}
	log.Info("starting the rest server")
	restServer.Start()
}
