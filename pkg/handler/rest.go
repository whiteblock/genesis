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

package handler

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/whiteblock/definition/command"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/handler/auxillary"
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

type restHandler struct {
	aux auxillary.Executor
	log logrus.Ext1FieldLogger
}

//NewRestHandler creates a new rest handler
func NewRestHandler(aux auxillary.Executor, log logrus.Ext1FieldLogger) RestHandler {
	log.Debug("creating a new rest handler")
	out := &restHandler{
		aux: aux,
		log: log,
	}
	return out
}

//AddCommands handles the addition of new commands
func (rh *restHandler) AddCommands(w http.ResponseWriter, r *http.Request) {
	var cmds command.Instructions
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 400)
		return
	}

	defer r.Body.Close()
	err = json.Unmarshal(data, &cmds)
	if err != nil {
		http.Error(w, util.LogError(err).Error(), 400)
		return
	}
	go rh.run(&cmds)
	w.Write([]byte("Success"))
}

func (rh *restHandler) process(inst *command.Instructions) (result entity.Result) {
	cmds, err := inst.Peek()

	isLastOne := false
	if err != nil {
		if errors.Is(err, command.ErrNoCommands) {
			rh.log.WithField("error", err).Error("ignoring empty message")
			return entity.NewIgnoreResult(err)
		}
		if !errors.Is(err, command.ErrDone) {
			rh.log.Error(err)
			return entity.NewFatalResult(err).InjectMeta(map[string]interface{}{
				"instructions": *inst,
			})
		}
		isLastOne = true
	}

	result = rh.aux.ExecuteCommands(cmds)

	if result.IsFatal() {
		rh.log.WithFields(logrus.Fields{"result": result, "error": result.Error.Error(),
			"testnet": inst.ID}).Error("execution resulted in a fatal error")

		result = result.InjectMeta(map[string]interface{}{
			command.OrgIDKey:        inst.OrgID,
			command.TestIDKey:       inst.ID,
			command.DefinitionIDKey: inst.DefinitionID,
		})
	} else if result.IsTrap() {
		rh.log.WithField("result", result).Debug("propogating the trap")
	} else if isLastOne && result.IsSuccess() {
		if inst.NeverTerminate() {
			return result.Trap()
		}
		rh.log.Debug("creating completion message")
		result = entity.NewAllDoneResult()
	} else if result.IsSuccess() {
		result = entity.NewRequeueResult()
		rh.log.WithField("remaining", len(inst.Commands)).Debug("creating message for next round")
		inst.Next()
	} else if failed, ok := checkPartialFailure(cmds, result); ok {
		rh.log.WithFields(logrus.Fields{
			"failed": failed, "succeeded": len(cmds) - len(failed),
			"result": result,
		}).Warn("something went partially wrong, requeuing only the commands which failed")
		inst.PartialCompletion(failed)
	} else {
		rh.log.WithField("result", result).Debug("something went wrong, getting kickback message")
	}
	return
}

//HealthCheck handles the reporting of the current health of this service
func (rh *restHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("OK"))
	if err != nil {
		rh.log.Error(err)
	}
}

func (rh *restHandler) run(inst *command.Instructions) {
	retries := 0
	for {
		res := rh.process(inst)

		if res.IsAllDone() {
			rh.log.Info("successfully completed")
			return
		}
		if res.IsFatal() {
			rh.log.Error("a command could not execute")
			return
		}

		if res.IsIgnore() {
			rh.log.Error("ignoring a message")
			return
		}
		if res.IsTrap() {
			rh.log.Info("a trap was activated")
			return
		}

		if res.IsRequeue() || !res.IsSuccess() {
			retries++
			if retries > maxRetries {
				rh.log.Error("too many retries for command")
				return
			}
			rh.log.Info("retrying command")
			continue
		} else {
			retries = 0
		}
		continue
	}
}
