/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package handler

import (
	"encoding/json"
	"testing"

	auxMocks "github.com/whiteblock/genesis/mocks/pkg/handler/auxillary"
	"github.com/whiteblock/genesis/pkg/config"
	"github.com/whiteblock/genesis/pkg/entity"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/whiteblock/definition/command"
)

func TestNewDeliveryHandler(t *testing.T) {
	assert.NotNil(t, NewDeliveryHandler(nil, config.Config{}, 1, nil))
}

func TestDeliveryHandler_Process_Successful(t *testing.T) {
	aux := new(auxMocks.Executor)
	aux.On("ExecuteCommands", mock.Anything).Return(entity.NewSuccessResult()).Once()

	dh := NewDeliveryHandler(aux, config.Config{}, 1, logrus.New())

	cmd := command.Instructions{Commands: [][]command.Command{{command.Command{
		Order: command.Order{
			Type:    "createContainer",
			Payload: map[string]interface{}{},
		},
		Target: command.Target{
			IP: "127.0.0.1",
		},
	}}}}

	body, err := json.Marshal(cmd)
	require.NoError(t, err)

	_, _, res := dh.Process(amqp.Delivery{Body: body})
	assert.NoError(t, res.Error)

	aux.AssertExpectations(t)

}

func TestDeliveryHandler_Process_Unsuccessful(t *testing.T) {
	aux := new(auxMocks.Executor)

	dh := NewDeliveryHandler(aux, config.Config{}, 1, logrus.New())

	body := []byte("should be a failure")

	_, _, res := dh.Process(amqp.Delivery{Body: body})
	assert.Error(t, res.Error)

	aux.AssertExpectations(t)
}

func TestDeliveryHandler_Process_NoCmds_Failures(t *testing.T) {
	dh := NewDeliveryHandler(nil, config.Config{}, 1, logrus.New())

	cmd := command.Instructions{}

	body, err := json.Marshal(cmd)
	assert.NoError(t, err)

	_, _, res := dh.Process(amqp.Delivery{Body: body})
	assert.Error(t, res.Error)
}

func TestDeliveryHandler_Process_Multiple_Commands_Successful(t *testing.T) {
	aux := new(auxMocks.Executor)
	aux.On("ExecuteCommands", mock.Anything).Return(entity.NewSuccessResult()).Once()

	dh := NewDeliveryHandler(aux, config.Config{}, 1, logrus.New())

	cmd := command.Instructions{Commands: [][]command.Command{
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
	}}

	body, err := json.Marshal(cmd)
	assert.NoError(t, err)

	_, _, res := dh.Process(amqp.Delivery{Body: body})
	assert.NoError(t, res.Error)

	aux.AssertExpectations(t)

}

func TestDeliveryHandler_Process_Execute_Nonfatal_Failure(t *testing.T) {
	aux := new(auxMocks.Executor)
	aux.On("ExecuteCommands", mock.Anything).Return(entity.NewErrorResult("err")).Once()
	dh := NewDeliveryHandler(aux, config.Config{}, 1, logrus.New())

	cmd := command.Instructions{Commands: [][]command.Command{
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
	}}

	body, err := json.Marshal(cmd)
	assert.NoError(t, err)

	_, _, res := dh.Process(amqp.Delivery{Body: body})
	assert.Error(t, res.Error)

	aux.AssertExpectations(t)

}

func TestDeliveryHandler_Process_Execute_Fatal_Failure(t *testing.T) {
	aux := new(auxMocks.Executor)
	aux.On("ExecuteCommands", mock.Anything).Return(entity.NewFatalResult("err")).Once()
	dh := NewDeliveryHandler(aux, config.Config{}, 1, logrus.New())

	cmd := command.Instructions{Commands: [][]command.Command{
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
	}}

	body, err := json.Marshal(cmd)
	assert.NoError(t, err)

	_, _, res := dh.Process(amqp.Delivery{Body: body})
	assert.Error(t, res.Error)

	aux.AssertExpectations(t)

}
