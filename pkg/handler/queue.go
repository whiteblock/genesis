/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package handler

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/whiteblock/genesis/pkg/config"
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
	conf       config.Config
}

// NewDeliveryHandler creates a new DeliveryHandler which uses the given usecase for
// executing the extracted command
func NewDeliveryHandler(
	aux auxillary.Executor,
	conf config.Config,
	maxRetries int64,
	log logrus.Ext1FieldLogger) DeliveryHandler {
	return &deliveryHandler{aux: aux, conf: conf, log: log, maxRetries: maxRetries}
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

func checkPartialFailure(cmds []command.Command, result entity.Result) ([]string, bool) {
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
	if result.IsDelayed() {
		inst.Next()
		out, err = queue.GetNextMessage(msg, inst)
	} else if result.IsFatal() {
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
	} else if failed, ok := checkPartialFailure(cmds, result); ok {
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
		result = result.Fatal().InjectMeta(map[string]interface{}{
			"secondaryError": err,
		})
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
	if dh.conf.Execution.DebugMode && result.IsFatal() {
		dh.log.Info("trapping fatal error due to debug mode")
		result = result.Trap()
	}

	if result.IsAllDone() || result.IsTrap() || result.IsFatal() || result.IsIgnore() {
		stat.Finished = true
		stat.StepsLeft = 0
	}
	if !result.IsSuccess() {
		stat.Message = result.Error.Error()
	}
	if result.IsDelayed() {
		dh.log.WithFields(logrus.Fields{
			"result": result,
		}).Info("adding the delay field to the header")
		out.Headers["x-delay"] = result.Delay.Milliseconds()
	}

	status, err = queue.CreateMessage(stat)
	if err != nil {
		dh.log.WithField("error", err).Error("malformed status generated")
	}
	return
}
