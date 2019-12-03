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
	"net/http"
	"sync"

	"github.com/whiteblock/definition/command"
	"github.com/whiteblock/genesis/pkg/usecase"
	util "github.com/whiteblock/utility/utils"

	"github.com/sirupsen/logrus"
)

const maxRetries = 5

//RestHandler handles the REST api calls
type RestHandler interface {
	//AddCommands handles the addition of new commands
	AddCommands(w http.ResponseWriter, r *http.Request)
	//HealthCheck handles the reporting of the current health of this service
	HealthCheck(w http.ResponseWriter, r *http.Request)
}

type commandWrapper struct {
	cmd     command.Command
	retries int
}

type restHandler struct {
	cmdChan chan commandWrapper
	uc      usecase.DockerUseCase
	once    *sync.Once
	log     logrus.Ext1FieldLogger
}

//NewRestHandler creates a new rest handler
func NewRestHandler(uc usecase.DockerUseCase, log logrus.Ext1FieldLogger) RestHandler {
	log.Debug("creating a new rest handler")
	out := &restHandler{
		cmdChan: make(chan commandWrapper, 200),
		uc:      uc,
		once:    &sync.Once{},
		log:     log,
	}
	go out.start()
	return out
}

//AddCommands handles the addition of new commands
func (rH *restHandler) AddCommands(w http.ResponseWriter, r *http.Request) {
	var cmds [][]command.Command
	err := json.NewDecoder(r.Body).Decode(&cmds)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 400)
		return
	}
	for _, commands := range cmds {
		for _, cmd := range commands {
			rH.cmdChan <- commandWrapper{cmd: cmd, retries: 0}
		}
	}
	w.Write([]byte("Success"))
}

//HealthCheck handles the reporting of the current health of this service
func (rH *restHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("OK"))
	if err != nil {
		rH.log.Error(err)
	}
}

//start starts the states inner consume loop
func (rH *restHandler) start() {
	rH.once.Do(func() {
		rH.loop()
	})
}

func (rH *restHandler) runCommand(cmd commandWrapper) {
	res := rH.uc.Run(cmd.cmd)
	entry := rH.log.WithFields(logrus.Fields{"result": res, "count": cmd.retries})

	if res.IsSuccess() {
		entry.Trace("a command executed successfully")
		return
	}
	if res.IsFatal() {
		entry.Error("a command could not execute")
		return
	}
	if res.IsRequeue() {
		go func() {
			if cmd.retries < maxRetries {
				cmd.retries++
				entry.Info("retrying command")
				rH.cmdChan <- cmd

			}
			entry.Error("too many retries for command")
		}()
		return
	}
}

func (rH *restHandler) loop() {
	for {
		cmd, closed := <-rH.cmdChan
		if !closed {
			return
		}
		rH.log.WithField("command", cmd).Trace("attempting to run a command")
		rH.runCommand(cmd)
	}
}
