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

package handler

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/service"
	"github.com/whiteblock/genesis/pkg/usecase"
	util "github.com/whiteblock/utility/utils"
	"net/http"
	"sync"
)

//RestHandler handles the REST api calls
type RestHandler interface {
	//AddCommands handles the addition of new commands
	AddCommands(w http.ResponseWriter, r *http.Request)
	//HealthCheck handles the reporting of the current health of this service
	HealthCheck(w http.ResponseWriter, r *http.Request)
}

type restHandler struct {
	cmdChan chan command.Command
	serv    service.CommandService
	uc      usecase.DockerUseCase
	once    *sync.Once
}

//NewRestHandler creates a new rest handler
func NewRestHandler(uc usecase.DockerUseCase, serv service.CommandService) RestHandler {
	log.Debug("creating a new rest handler")
	out := &restHandler{
		cmdChan: make(chan command.Command),
		serv:    serv,
		uc:      uc,
		once:    &sync.Once{},
	}
	go out.start()
	return out
}

//AddCommands handles the addition of new commands
func (rH *restHandler) AddCommands(w http.ResponseWriter, r *http.Request) {
	var commands []command.Command
	err := json.NewDecoder(r.Body).Decode(&commands)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 400)
		return
	}
	for _, cmd := range commands {
		rH.cmdChan <- cmd
	}
	w.Write([]byte("Success"))
}

//HealthCheck handles the reporting of the current health of this service
func (rH *restHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Fatal(err)
	}
}

//TODO: move this logic out of the rest handler

//start starts the states inner consume loop
func (rH *restHandler) start() {
	rH.once.Do(func() {
		rH.loop()
	})
}

func (rH *restHandler) runCommand(cmd command.Command) {
	res := rH.uc.Run(cmd)
	log.WithFields(log.Fields{"result": res}).Debug("got a result")
	if res.IsRequeue() {
		rH.cmdChan <- cmd
	} else {
		rH.serv.ReportCommandResult(cmd, res)
	}
}

func (rH *restHandler) loop() {
	for {
		cmd, closed := <-rH.cmdChan
		if !closed {
			return
		}
		log.WithFields(log.Fields{"command": cmd}).Trace("attempting to run a command")
		go rH.runCommand(cmd)
	}
}
