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
	"time"

	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/handler/auxillary"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	queue "github.com/whiteblock/amqp"
	"github.com/whiteblock/definition/command"
	"github.com/whiteblock/definition/command/biome"
)

// DeliveryHandler handles the initial processing of a amqp delivery
type DeliveryHandler interface {
	// Process attempts to extract the command and execute it
	Process(msg amqp.Delivery) (amqp.Publishing, amqp.Publishing, entity.Result)
}

type deliveryHandler struct {
	maxRetries int64
	aux        auxillary.Executor
	log        logrus.Ext1FieldLogger
}

// NewDeliveryHandler creates a new DeliveryHandler which uses the given usecase for
// executing the extracted command
func NewDeliveryHandler(
	aux auxillary.Executor,
	maxRetries int64,
	log logrus.Ext1FieldLogger) DeliveryHandler {
	return &deliveryHandler{aux: aux, log: log, maxRetries: maxRetries}
}

func (dh deliveryHandler) sleepy(msg amqp.Delivery) {
	if msg.Headers != nil {
		return
	}
	if _, ok := msg.Headers["retryCount"]; !ok {
		return
	}

	if _, ok := msg.Headers["retryCount"].(int64); !ok {
		return
	}
	if msg.Headers["retryCount"].(int64) > 0 {
		time.Sleep(5 * time.Second)
	}
}

func (dh deliveryHandler) checkPartialFailure(cmds []command.Command, result entity.Result) ([]string, bool) {
	if _, hasFailed := result.Meta["failed"]; !hasFailed {
		return nil, false
	}
	failed, ok := result.Meta["failed"].([]string)
	if !ok || failed == nil {
		return nil, false
	}
	return failed, len(failed) != len(cmds)
}

func (dh deliveryHandler) destructMsg(inst *command.Instructions) amqp.Publishing {
	out, err := queue.CreateMessage(biome.DestroyBiome{
		TestID: inst.ID,
	})
	if err != nil {
		dh.log.Error(err)
	}
	return out
}

func (dh deliveryHandler) process(msg amqp.Delivery,
	inst *command.Instructions) (out amqp.Publishing, result entity.Result) {

	cmds, err := inst.Peek()

	isLastOne := false
	if err != nil {
		if errors.Is(err, command.ErrNoCommands) {
			dh.log.WithField("error", err).Error("ignoring empty message")
			return amqp.Publishing{}, entity.NewIgnoreResult(err)
		}
		if !errors.Is(err, command.ErrDone) {
			dh.log.Error(err)
			return dh.destructMsg(inst), entity.NewFatalResult(err).InjectMeta(map[string]interface{}{
				"instructions": *inst,
			})
		}
		isLastOne = true
	}

	result = dh.aux.ExecuteCommands(cmds)

	if result.IsFatal() {
		dh.log.WithFields(logrus.Fields{"result": result, "error": result.Error.Error(),
			"testnet": inst.ID}).Error("execution resulted in a fatal error")

		out = dh.destructMsg(inst)
		result = result.InjectMeta(map[string]interface{}{
			command.OrgIDKey:        inst.OrgID,
			command.TestIDKey:       inst.ID,
			command.DefinitionIDKey: inst.DefinitionID,
		})
	} else if result.IsTrap() {
		dh.log.WithField("result", result).Debug("propogating the trap")
	} else if isLastOne && result.IsSuccess() {
		if inst.NeverTerminate() {
			result = result.Trap()
			return
		}
		dh.log.Debug("creating completion message")
		result = entity.NewAllDoneResult()
		out, err = queue.CreateMessage(biome.DestroyBiome{
			TestID: inst.ID,
		})
	} else if result.IsSuccess() {
		result = entity.NewRequeueResult()
		dh.log.WithField("remaining", len(inst.Commands)).Debug("creating message for next round")
		inst.Next()
		out, err = queue.GetNextMessage(msg, inst)
	} else if failed, ok := dh.checkPartialFailure(cmds, result); ok {
		dh.log.WithFields(logrus.Fields{
			"failed": failed, "succeeded": len(cmds) - len(failed),
			"result": result,
		}).Warn("something went partially wrong, requeuing only the commands which failed")
		inst.PartialCompletion(failed)
		out, err = queue.GetNextMessage(msg, inst)
	} else {
		dh.log.WithField("result", result).Debug("something went wrong, getting kickback message")
		out, err = queue.GetKickbackMessage(dh.maxRetries, msg)
	}

	if err != nil {
		dh.log.WithFields(logrus.Fields{
			"result": result,
			"err":    err}).Error("a fatal error occured, flagging as fatal")
		result = result.Fatal(err)
		out = dh.destructMsg(inst)
	}
	return
}

//Process attempts to extract the command and execute it
func (dh deliveryHandler) Process(msg amqp.Delivery) (out amqp.Publishing,
	status amqp.Publishing, result entity.Result) {
	dh.sleepy(msg)

	var inst command.Instructions
	err := json.Unmarshal(msg.Body, &inst)
	if err != nil {
		dh.log.WithField("error", err).Error("received malformed instructions")
		return dh.destructMsg(&inst), amqp.Publishing{},
			entity.NewFatalResult(err).InjectMeta(map[string]interface{}{
				"data": msg.Body,
			})
	}
	out, result = dh.process(msg, &inst)

	stat := inst.Status()

	if result.IsAllDone() || result.IsTrap() || result.IsFatal() || result.IsIgnore() {
		stat.Finished = true
	}
	if !result.IsSuccess() {
		stat.Message = result.Error.Error()
	}

	status, err = queue.CreateMessage(stat)
	if err != nil {
		dh.log.WithField("error", err).Error("malformed status generated")
	}
	return
}
