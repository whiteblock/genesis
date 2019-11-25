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

	auxMocks "github.com/whiteblock/genesis/mocks/pkg/handler/auxillary"
	utilityMocks "github.com/whiteblock/genesis/mocks/pkg/utility"
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewDeliveryHandler(t *testing.T) {
	assert.NotNil(t, NewDeliveryHandler(nil, nil, nil))
}

func TestDeliveryHandler_Process_Successful(t *testing.T) {
	aux := new(auxMocks.Executor)
	aux.On("ExecuteCommands", mock.Anything).Return(entity.NewSuccessResult()).Once()
	util := new(utilityMocks.AMQPMessage)
	util.On("CreateMessage", mock.Anything).Return(amqp.Publishing{}, nil).Once()

	dh := NewDeliveryHandler(aux, util, logrus.New())

	cmd := [][]command.Command{[]command.Command{command.Command{
		Order: command.Order{
			Type:    "createContainer",
			Payload: map[string]interface{}{},
		},
		Target: command.Target{
			IP: "127.0.0.1",
		},
	}}}

	body, err := json.Marshal(cmd)
	assert.NoError(t, err)

	_, res := dh.Process(amqp.Delivery{Body: body})
	assert.NoError(t, res.Error)

	aux.AssertExpectations(t)
	util.AssertExpectations(t)
}

func TestDeliveryHandler_Process_Unsuccessful(t *testing.T) {
	aux := new(auxMocks.Executor)

	dh := NewDeliveryHandler(aux, nil, logrus.New())

	body := []byte("should be a failure")

	_, res := dh.Process(amqp.Delivery{Body: body})
	assert.Error(t, res.Error)

	aux.AssertExpectations(t)
}

func TestDeliveryHandler_Process_NoCmds_Failures(t *testing.T) {
	dh := NewDeliveryHandler(nil, nil, logrus.New())

	cmd := [][]command.Command{}

	body, err := json.Marshal(cmd)
	assert.NoError(t, err)

	_, res := dh.Process(amqp.Delivery{Body: body})
	assert.Error(t, res.Error)
}

func TestDeliveryHandler_Process_Multiple_Commands_Successful(t *testing.T) {
	aux := new(auxMocks.Executor)
	aux.On("ExecuteCommands", mock.Anything).Return(entity.NewSuccessResult()).Once()
	util := new(utilityMocks.AMQPMessage)
	util.On("GetNextMessage", mock.Anything, mock.Anything).Return(amqp.Publishing{}, nil).Once()

	dh := NewDeliveryHandler(aux, util, logrus.New())

	cmd := [][]command.Command{
		[]command.Command{
			command.Command{
				Order: command.Order{
					Type:    "createContainer",
					Payload: map[string]interface{}{},
				},
				Target: command.Target{
					IP: "127.0.0.1",
				},
			},
		},
		[]command.Command{
			command.Command{
				Order: command.Order{
					Type:    "createContainer",
					Payload: map[string]interface{}{},
				},
				Target: command.Target{
					IP: "127.0.0.1",
				},
			},
		},
	}

	body, err := json.Marshal(cmd)
	assert.NoError(t, err)

	_, res := dh.Process(amqp.Delivery{Body: body})
	assert.NoError(t, res.Error)

	aux.AssertExpectations(t)
	util.AssertExpectations(t)
}

func TestDeliveryHandler_Process_Execute_Nonfatal_Failure(t *testing.T) {
	aux := new(auxMocks.Executor)
	aux.On("ExecuteCommands", mock.Anything).Return(entity.NewErrorResult("err")).Once()
	util := new(utilityMocks.AMQPMessage)
	util.On("GetKickbackMessage", mock.Anything).Return(amqp.Publishing{}, nil).Once()

	dh := NewDeliveryHandler(aux, util, logrus.New())

	cmd := [][]command.Command{
		[]command.Command{
			command.Command{
				Order: command.Order{
					Type:    "createContainer",
					Payload: map[string]interface{}{},
				},
				Target: command.Target{
					IP: "127.0.0.1",
				},
			},
		},
	}

	body, err := json.Marshal(cmd)
	assert.NoError(t, err)

	_, res := dh.Process(amqp.Delivery{Body: body})
	assert.Error(t, res.Error)

	aux.AssertExpectations(t)
	util.AssertExpectations(t)
}

func TestDeliveryHandler_Process_Execute_Fatal_Failure(t *testing.T) {
	aux := new(auxMocks.Executor)
	aux.On("ExecuteCommands", mock.Anything).Return(entity.NewFatalResult("err")).Once()
	util := new(utilityMocks.AMQPMessage)

	dh := NewDeliveryHandler(aux, util, logrus.New())

	cmd := [][]command.Command{
		[]command.Command{
			command.Command{
				Order: command.Order{
					Type:    "createContainer",
					Payload: map[string]interface{}{},
				},
				Target: command.Target{
					IP: "127.0.0.1",
				},
			},
		},
	}

	body, err := json.Marshal(cmd)
	assert.NoError(t, err)

	_, res := dh.Process(amqp.Delivery{Body: body})
	assert.Error(t, res.Error)

	aux.AssertExpectations(t)
	util.AssertExpectations(t)
}
