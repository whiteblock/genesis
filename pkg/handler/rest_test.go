/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/whiteblock/definition/command"
	auxMocks "github.com/whiteblock/genesis/mocks/pkg/handler/auxillary"
	"github.com/whiteblock/genesis/pkg/entity"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var testCommands = command.Instructions{Commands: [][]command.Command{{
	command.Command{
		ID:     "TEST",
		Target: command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "createContainer",
			Payload: map[string]interface{}{},
		},
	},
	command.Command{
		ID:     "TEST2",
		Target: command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "createContainer",
			Payload: map[string]interface{}{},
		},
	},
}}}

func TestRestHandler(t *testing.T) {

	data, err := json.Marshal(testCommands)
	assert.NoError(t, err)
	req, err := http.NewRequest("POST", "/commands", bytes.NewReader(data))

	assert.NoError(t, err)

	runChan := make(chan []command.Command)

	aux := new(auxMocks.Executor)
	aux.On("ExecuteCommands", mock.Anything).Return(entity.NewSuccessResult()).Run(func(args mock.Arguments) {
		cmds, ok := args.Get(0).([]command.Command)
		assert.True(t, ok)
		runChan <- cmds
	}).Times(len(testCommands.Commands))

	rh := NewRestHandler(aux, logrus.New())

	recorder := httptest.NewRecorder()
	go rh.AddCommands(recorder, req)

	for i := range testCommands.Commands {
		select {
		case <-runChan:
		case <-time.After(5 * time.Second):
			t.Fatal(fmt.Sprintf("Report did not happen within 5 seconds: %d/%d", i, len(testCommands.Commands)))
		}
	}

	aux.AssertExpectations(t)
}

func TestRestHandler_Requeue(t *testing.T) {

	data, err := json.Marshal(testCommands)
	assert.NoError(t, err)
	req, err := http.NewRequest("POST", "/commands", bytes.NewReader(data))

	assert.NoError(t, err)

	runChan := make(chan []command.Command)

	aux := new(auxMocks.Executor)
	aux.On("ExecuteCommands", mock.Anything).Return(entity.NewErrorResult("err")).Run(func(args mock.Arguments) {
		cmds, ok := args.Get(0).([]command.Command)
		assert.True(t, ok)
		runChan <- cmds

	}).Times(len(testCommands.Commands) * (maxRetries + 1))

	rh := NewRestHandler(aux, logrus.New())

	recorder := httptest.NewRecorder()
	go rh.AddCommands(recorder, req)

	for i := 0; i < len(testCommands.Commands)*(maxRetries+1); i++ {
		select {
		case <-runChan:
		case <-time.After(5 * time.Second):
			t.Fatal(fmt.Sprintf("Report did not happen within 5 seconds: %d/%d", i,
				len(testCommands.Commands)*(maxRetries)))
		}
	}
	aux.AssertExpectations(t)
}

func TestRestHandler_Fatal(t *testing.T) { //testCommands
	data, err := json.Marshal(testCommands)
	assert.NoError(t, err)
	req, err := http.NewRequest("POST", "/commands", bytes.NewReader(data))

	assert.NoError(t, err)

	runChan := make(chan []command.Command)

	aux := new(auxMocks.Executor)
	aux.On("ExecuteCommands", mock.Anything).Return(entity.NewFatalResult("err")).Run(func(args mock.Arguments) {
		t.Log("called run")
		cmds, ok := args.Get(0).([]command.Command)
		assert.True(t, ok)
		runChan <- cmds

	}).Times(len(testCommands.Commands))

	rh := NewRestHandler(aux, logrus.New())

	recorder := httptest.NewRecorder()
	rh.AddCommands(recorder, req)

	for range testCommands.Commands {
		select {
		case <-runChan:
		case <-time.After(5 * time.Second):
			t.Fatal("Report did not happen within 5 seconds")
		}
	}
	aux.AssertExpectations(t)
}

func TestRestHandler_HealthCheck(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", bytes.NewReader([]byte{}))
	assert.NoError(t, err)

	rh := NewRestHandler(nil, logrus.New())
	recorder := httptest.NewRecorder()
	rh.HealthCheck(recorder, req)

	assert.Equal(t, "OK", recorder.Body.String())
}
