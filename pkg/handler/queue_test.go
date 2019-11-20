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
	usecaseMocks "github.com/whiteblock/genesis/mocks/pkg/usecase"
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"
)

func TestNewDeliveryHandler(t *testing.T) {
	duc := new(usecaseMocks.DockerUseCase)

	expected := &deliveryHandler{
		usecase: duc,
	}

	dh := NewDeliveryHandler(duc)

	assert.Equal(t, expected, dh)
}

func TestDeliveryHandler_ProcessMessage_Successful(t *testing.T) {
	duc := new(usecaseMocks.DockerUseCase)
	duc.On("Run", mock.Anything).Return(entity.Result{Type: entity.SuccessType}).Once()

	dh := NewDeliveryHandler(duc)

	cmd := new(command.Command)
	cmd.Order.Type = "createContainer"
	cmd.Order.Payload = map[string]interface{}{}
	cmd.Target.IP = "127.0.0.1"

	body, err := json.Marshal(cmd)
	if err != nil {
		t.Error(err)
	}

	res := dh.ProcessMessage(amqp.Delivery{Body: body})
	if res.Error != nil {
		t.Error("expected return value of ProcessMessage does not match expected value: ", err)
	}

	assert.NoError(t, res.Error)
	duc.AssertExpectations(t)
}

func TestDeliveryHandler_ProcessMessage_Unsuccessful(t *testing.T) {
	dh := NewDeliveryHandler(new(usecaseMocks.DockerUseCase))

	body := []byte("should be a failure")

	res := dh.ProcessMessage(amqp.Delivery{Body: body})
	assert.Error(t, res.Error)
}

func TestDeliveryHandler_GetKickbackMessage_Successful(t *testing.T) {
	duc := new(usecaseMocks.DockerUseCase)
	duc.On("TimeSupplier").Return(int64(5))

	dh := NewDeliveryHandler(duc)

	cmd := new(command.Command)
	cmd.ID = "unit_test"
	cmd.Retry = uint8(1)

	body, err := json.Marshal(cmd)
	assert.NoError(t, err)

	msg := amqp.Delivery{
		Body: body,
	}

	res, err := dh.GetKickbackMessage(msg)
	assert.NoError(t, err)

	var resCmd command.Command
	err = json.Unmarshal(res.Body, &resCmd)
	assert.NoError(t, err)

	assert.Exactly(t, resCmd.Timestamp, int64(5))
	assert.Exactly(t, resCmd.Retry, uint8(2))
	assert.Exactly(t, resCmd.ID, "unit_test")
}

func TestDeliveryHandler_GetKickbackMessage_Unsuccessful(t *testing.T) {
	duc := new(usecaseMocks.DockerUseCase)

	dh := NewDeliveryHandler(duc)

	body := []byte("supposed to fail")

	msg := amqp.Delivery{
		Body: body,
	}

	_, err := dh.GetKickbackMessage(msg)
	assert.Error(t, err)
}
