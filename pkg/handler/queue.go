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
	"github.com/streadway/amqp"
	"github.com/whiteblock/genesis/pkg/usecase"
	"github.com/whiteblock/genesis/pkg/command"
	"time"
)

type DeliveryHandler interface {
	ProcessMessage(msg amqp.Delivery) entity.Result
	GetKickbackMessage(msg amqp.Delivery) amqp.Publishing
}

type deliveryHandler struct {
	usecase usecase.CommandUseCase
}

func NewDeliveryHandler(usecase usecase.CommandUseCase) (DeliveryHandler,error) {
	return deliveryHandler{usecase:usecase},nil
}

func (dh deliveryHandler) ProcessMessage(msg amqp.Delivery) entity.Result {
	var cmd Command
	err := json.Unmarshal(msg.Body,&cmd)
	if err != nil {
		return entity.Result{Error:err}
	}
	return usecase.Run(cmd)
}

func (dh deliveryHandler) GetKickbackMessage(msg amqp.Delivery) (amqp.Publishing,error) {
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

	var cmd Command
	err := json.Unmarshal(msg.Body,&cmd)
	if err != nil {
		return pub,err
	}
	cmd = cmd.GetRetryCommand(time.Now().Unix())

	body,err := json.Marshal(cmd)
	if err != nil {
		return pub,err
	}
	pub.Body = body
	return pub,nil
}