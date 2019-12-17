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
	"time"

	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/handler/auxillary"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	queue "github.com/whiteblock/amqp"
	"github.com/whiteblock/definition/command"
	"github.com/whiteblock/definition/command/biome"
)

//DeliveryHandler handles the initial processing of a amqp delivery
type DeliveryHandler interface {
	//Process attempts to extract the command and execute it
	Process(msg amqp.Delivery) (out amqp.Publishing, result entity.Result)
}

type deliveryHandler struct {
	msgUtil queue.AMQPMessage
	aux     auxillary.Executor
	log     logrus.Ext1FieldLogger
}

//NewDeliveryHandler creates a new DeliveryHandler which uses the given usecase for
//executing the extracted command
func NewDeliveryHandler(
	aux auxillary.Executor,
	msgUtil queue.AMQPMessage,
	log logrus.Ext1FieldLogger) DeliveryHandler {
	return &deliveryHandler{aux: aux, log: log, msgUtil: msgUtil}
}

func (dh deliveryHandler) sleepy(msg amqp.Delivery) {
	if pub.Headers != nil {
		return
	}
	if _,ok := pub.Headers["retryCount"]; !ok {
		return
	}

	if _, ok := pub.Headers["retryCount"].(int64); !ok {
		return
	}
	if pub.Headers["retryCount"].(int64) > 1 {
		time.Sleep(5 * time.Second)
	}
}

//Process attempts to extract the command and execute it
func (dh deliveryHandler) Process(msg amqp.Delivery) (out amqp.Publishing, result entity.Result) {
	dh.sleepy(msg)

	var allCmds [][]command.Command
	err := json.Unmarshal(msg.Body, &allCmds)
	if err != nil {
		dh.log.WithField("error", err).Errorf("received a malformed command sausage")
		return amqp.Publishing{}, entity.NewFatalResult(err)
	}

	if len(allCmds) == 0 {
		dh.log.Error("recieved an empty command sausage")
		return amqp.Publishing{}, entity.NewFatalResult("nothing to execute")
	}

	result = dh.aux.ExecuteCommands(allCmds[0])
	if result.IsFatal() {
		out, err = dh.msgUtil.CreateMessage(biome.DestroyBiome{
			TestnetID: allCmds[0][0].Target.TestnetID,
		})
		dh.log.WithFields(logrus.Fields{
			"result":  result,
			"error":   result.Error.Error(),
			"testnet": allCmds[0][0].Target.TestnetID,
		}).Error("execution resulted in a fatal error")
		return
	}

	if result.IsTrap() {
		dh.log.Debug("propogating the trap")
		return
	}

	if result.IsSuccess() {
		if len(allCmds) != 1 {
			result = entity.NewRequeueResult()
			dh.log.WithField("remaining", len(allCmds)-1).Debug("creating message for next round")
			out, err = dh.msgUtil.GetNextMessage(msg, allCmds[1:])
		} else {
			dh.log.Debug("creating completion message")
			result = entity.NewAllDoneResult()
			out, err = dh.msgUtil.CreateMessage(biome.DestroyBiome{
				TestnetID: allCmds[0][0].Target.TestnetID,
			})
		}
	} else {
		dh.log.WithField("result", result).Debug("something went wrong, getting kickback message")
		out, err = dh.msgUtil.GetKickbackMessage(msg)
	}
	if err != nil {
		dh.log.WithFields(logrus.Fields{
			"result": result,
			"err":    err}).Error("a fatal error occured, flagging as fatal")
		result = entity.NewFatalResult(err)
	}
	return
}
