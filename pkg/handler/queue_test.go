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

func TestDeliveryHandler_ProcessMessage(t *testing.T) {
	//service := new(mocks.DockerService)
	//service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	//service.On("CreateContainer", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	uc := new(usecaseMocks.DockerUseCase)
	uc.On("Run", mock.Anything).Return(entity.Result{Type: entity.SuccessType})
	uc.On("CreateContainer", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})
	//cmdService := new(mocks.CommandService)
	//cmdService.On("CheckDependenciesExecuted", mock.Anything).Return(true)

	dh, err := NewDeliveryHandler(uc)
	if err != nil {
		t.Error(err)
	}

	cmd := new(command.Command)
	cmd.Order.Type = "createContainer"
	cmd.Order.Payload = map[string]interface{}{}
	cmd.Target.IP = "0.0.0.0"

	body, err := json.Marshal(cmd)
	if err != nil {
		t.Error(err)
	}

	res := dh.ProcessMessage(amqp.Delivery{Body: body})
	if res.Error != nil {
		t.Error("expected return value of ProcessMessage does not match expected value: ", err)
	}

	assert.Equal(t, res.Error, nil)
	assert.True(t, uc.AssertNumberOfCalls(t, "Run", 1))
	assert.True(t, uc.AssertNumberOfCalls(t, "CreateContainer", 1))
}

func TestDeliveryHandler_GetKickbackMessage(t *testing.T) {
	duc := new(usecaseMocks.DockerUseCase)

	dh, err := NewDeliveryHandler(duc)
	if err != nil {
		t.Error(err)
	}

	cmd := new(command.Command)
	cmd.ID = "unit_test"
	cmd.Retry = uint8(1)

	body, err := json.Marshal(cmd)
	if err != nil {
		t.Error(err)
	}

	msg := amqp.Delivery{
		Body: body,
	}

	res, err := dh.GetKickbackMessage(msg)
	if err != nil {
		t.Error(err)
	}

	var resCmd command.Command
	err = json.Unmarshal(res.Body, &resCmd)
	if err != nil {
		t.Error(err)
	}

	assert.Exactly(t, resCmd.Timestamp, int64(5))
	assert.Exactly(t, resCmd.Retry, uint8(2))
}
