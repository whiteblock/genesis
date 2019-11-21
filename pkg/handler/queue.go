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
	"fmt"

	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/handler/auxillary"
	"github.com/whiteblock/genesis/pkg/utility"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

//DeliveryHandler handles the initial processing of a amqp delivery
type DeliveryHandler interface {
	//Process attempts to extract the command and execute it
	Process(msg amqp.Delivery) (out amqp.Publishing, result entity.Result)
}

type deliveryHandler struct {
	msgUtil utility.AMQPMessage
	aux     auxillary.Executor
	log     logrus.Ext1FieldLogger
}

//NewDeliveryHandler creates a new DeliveryHandler which uses the given usecase for
//executing the extracted command
func NewDeliveryHandler(
	aux auxillary.Executor,
	msgUtil utility.AMQPMessage,
	log logrus.Ext1FieldLogger) DeliveryHandler {
	return &deliveryHandler{aux: aux, log: log, msgUtil: msgUtil}
}

//Process attempts to extract the command and execute it
func (dh deliveryHandler) Process(msg amqp.Delivery) (out amqp.Publishing, result entity.Result) {
	var allCmds [][]command.Command
	err := json.Unmarshal(msg.Body, &allCmds)
	if err != nil {
		return amqp.Publishing{}, entity.Result{Error: err}
	}

	if len(allCmds) == 0 {
		dh.log.Error("recieved an empty command sausage")
		return amqp.Publishing{}, entity.NewFatalResult(fmt.Errorf("nothing to execute"))
	}

	result = dh.aux.ExecuteCommands(allCmds[0])
	if result.IsFatal() {
		return
	}

	if result.IsSuccess() {
		if len(allCmds) != 1 {
			out, err = dh.msgUtil.GetNextMessage(msg, allCmds[1:])
		} else {
			out, err = dh.msgUtil.CreateMessage(map[string]string{
				"testnetId": allCmds[0][0].Target.TestnetID,
			})
		}
	} else {
		out, err = dh.msgUtil.GetKickbackMessage(msg)
	}
	if err != nil {
		result = entity.NewFatalResult(msg)
	}
	return
}
