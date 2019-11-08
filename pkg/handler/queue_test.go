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
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	mocksHandler "github.com/whiteblock/genesis/mocks/pkg/handler"
	mocksUseCase "github.com/whiteblock/genesis/mocks/pkg/usecase"
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"
)

func TestNewDeliveryHandler(t *testing.T) {
	usecase := new(mocksUseCase.DockerUseCase)

	dh, err := NewDeliveryHandler(usecase)
	if err != nil {
		t.Error(err)
	}

	expectedDH := deliveryHandler{
		usecase: usecase,
	}

	assert.Equal(t, dh, expectedDH)
}

func TestDeliveryHandler_ProcessMessage(t *testing.T) {
	dh := new(mocksHandler.DeliveryHandler)
	dh.On("ProcessMessage", mock.Anything).Return(entity.Result{Error: nil, Type: entity.SuccessType})

	cmd := new(command.Command)
	body, err := json.Marshal(cmd)
	if err != nil {
		t.Error(err)
	}

	res := dh.ProcessMessage(amqp.Delivery{
		Body: body,
	})
	assert.Equal(t, res.Error, nil)
	assert.True(t, dh.AssertNumberOfCalls(t, "ProcessMessage", 1))
}

func TestDeliveryHandler_GetKickbackMessage(t *testing.T) {
	cmd := new(command.Command)
	body, err := json.Marshal(cmd)
	if err != nil {
		t.Error(err)
	}

	msg := amqp.Delivery{
		Body: body,
	}

	pub := amqp.Publishing{
		Headers:         msg.Headers,
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

	dh := new(mocksHandler.DeliveryHandler)
	dh.On("GetKickbackMessage", mock.Anything).Return(pub, nil)

	res, err := dh.GetKickbackMessage(msg)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, res, pub)
	assert.True(t, dh.AssertNumberOfCalls(t, "GetKickbackMessage", 1))
}
