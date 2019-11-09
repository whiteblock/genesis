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

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/usecase"
)

//DeliveryHandler handles the initial processing of a amqp delivery
type DeliveryHandler interface {
	//ProcessMessage attempts to extract the command and execute it
	ProcessMessage(msg amqp.Delivery) entity.Result
	//GetKickbackMessage takes the delivery and creates a message from it
	//for requeuing on non-fatal error
	GetKickbackMessage(msg amqp.Delivery) (amqp.Publishing, error)
}

type deliveryHandler struct {
	usecase usecase.DockerUseCase
}

//NewDeliveryHandler creates a new DeliveryHandler which uses the given usecase for
//executing the extracted command
func NewDeliveryHandler(usecase usecase.DockerUseCase) (DeliveryHandler, error) {
	return deliveryHandler{usecase: usecase}, nil
}

//ProcessMessage attempts to extract the command and execute it
func (dh deliveryHandler) ProcessMessage(msg amqp.Delivery) entity.Result {
	var cmd command.Command //TODO: add validation check
	err := json.Unmarshal(msg.Body, &cmd)
	if err != nil {
		return entity.Result{Error: err}
	}
	log.WithFields(log.Fields{"cmd": cmd}).Trace("finished processing a command from amqp")
	return dh.usecase.Run(cmd)
}

//GetKickbackMessage takes the delivery and creates a message from it
//for requeuing on non-fatal error
func (dh deliveryHandler) GetKickbackMessage(msg amqp.Delivery) (amqp.Publishing, error) {
	pub := amqp.Publishing{
		Headers: msg.Headers,
		// Properties
		ContentType:     msg.ContentType,
		ContentEncoding: msg.ContentEncoding,
		DeliveryMode:    msg.DeliveryMode,
		Priority:        msg.Priority,
		CorrelationId:   msg.CorrelationId,
		ReplyTo:         msg.ReplyTo,
		Expiration:      msg.Expiration,
		MessageId:       msg.MessageId,
		Timestamp:       msg.Timestamp,
		Type:            msg.Type,
	}

	var cmd command.Command
	err := json.Unmarshal(msg.Body, &cmd)
	if err != nil {
		return pub, err
	}

	if cmd.ID == "unit_test" {
		cmd = cmd.GetRetryCommand(int64(5))
	} else {
		cmd = cmd.GetRetryCommand(time.Now().Unix())
	}

	body, err := json.Marshal(cmd)
	if err != nil {
		return pub, err
	}
	pub.Body = body
	return pub, nil
}
