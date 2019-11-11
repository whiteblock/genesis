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
	"github.com/whiteblock/genesis/pkg/controller"
	"github.com/whiteblock/genesis/pkg/handler"
	"github.com/whiteblock/genesis/pkg/repository"
	"github.com/whiteblock/genesis/pkg/service"
	"github.com/whiteblock/genesis/pkg/usecase"
	"github.com/whiteblock/genesis/util"
)

var conf = util.GetConfig()

func getRestServer() (controller.RestController, error) {
	commandService := service.NewCommandService(repository.NewLocalCommandRepository())
	dockerService, err := service.NewDockerService(repository.NewDockerRepository())
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
	restServer, err := getRestServer()
	if err != nil {
		panic(err)
	}
	log.Info("starting the rest server")
	restServer.Start()
}
