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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	usecase "github.com/whiteblock/genesis/mocks/pkg/usecase"
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRestHandler(t *testing.T) {
	commands := []command.Command{
		command.Command{
			ID:        "TEST",
			Timestamp: 5,
			Timeout:   0,
			Target:    command.Target{IP: "0.0.0.0"},
			Order: command.Order{
				Type:    "createContainer",
				Payload: map[string]interface{}{},
			},
		},
		command.Command{
			ID:        "TEST2",
			Timestamp: 5,
			Timeout:   0,
			Target:    command.Target{IP: "0.0.0.0"},
			Order: command.Order{
				Type:    "createContainer",
				Payload: map[string]interface{}{},
			},
		},
	}
	data, err := json.Marshal(commands)
	assert.NoError(t, err)
	req, err := http.NewRequest("POST", "/commands", bytes.NewReader(data))

	assert.NoError(t, err)

	runChan := make(chan command.Command)

	uc := new(usecase.DockerUseCase)
	uc.On("Run", mock.Anything).Return(entity.NewSuccessResult()).Run(func(args mock.Arguments) {
		cmd, ok := args.Get(0).(command.Command)
		assert.True(t, ok)
		runChan <- cmd
	})

	rh := NewRestHandler(uc, logrus.New())

	recorder := httptest.NewRecorder()
	rh.AddCommands(recorder, req)

	for range commands {
		select {
		case <-runChan:
		case <-time.After(5 * time.Second):
			t.Fatal("Report did not happen within 5 seconds")
		}
	}
	uc.AssertNumberOfCalls(t, "Run", len(commands))
	close(rh.(*restHandler).cmdChan)
}

func TestRestHandler_Requeue(t *testing.T) {
	commands := []command.Command{
		command.Command{
			ID:        "TEST",
			Timestamp: 5,
			Timeout:   0,
			Target:    command.Target{IP: "0.0.0.0"},
			Order: command.Order{
				Type:    "createContainer",
				Payload: map[string]interface{}{},
			},
		},
		command.Command{
			ID:        "TEST2",
			Timestamp: 5,
			Timeout:   0,
			Target:    command.Target{IP: "0.0.0.0"},
			Order: command.Order{
				Type:    "createContainer",
				Payload: map[string]interface{}{},
			},
		},
	}
	data, err := json.Marshal(commands)
	assert.NoError(t, err)
	req, err := http.NewRequest("POST", "/commands", bytes.NewReader(data))

	assert.NoError(t, err)

	runChan := make(chan command.Command)

	uc := new(usecase.DockerUseCase)
	uc.On("Run", mock.Anything).Return(entity.NewErrorResult("err")).Run(func(args mock.Arguments) {
		cmd, ok := args.Get(0).(command.Command)
		assert.True(t, ok)
		runChan <- cmd
	}).Times(len(commands) * (maxRetries + 1))

	rh := NewRestHandler(uc, logrus.New())

	recorder := httptest.NewRecorder()
	rh.AddCommands(recorder, req)

	for i := 0; i < len(commands)*(maxRetries+1); i++ {
		select {
		case <-runChan:
		case <-time.After(5 * time.Second):
			t.Fatal("Report did not happen within 5 seconds")
		}
	}
	uc.AssertExpectations(t)
	close(rh.(*restHandler).cmdChan)
}

func TestRestHandler_Fatal(t *testing.T) {
	commands := []command.Command{
		command.Command{
			ID:        "TEST",
			Timestamp: 5,
			Timeout:   0,
			Target:    command.Target{IP: "0.0.0.0"},
			Order: command.Order{
				Type:    "createContainer",
				Payload: map[string]interface{}{},
			},
		},
		command.Command{
			ID:        "TEST2",
			Timestamp: 5,
			Timeout:   0,
			Target:    command.Target{IP: "0.0.0.0"},
			Order: command.Order{
				Type:    "createContainer",
				Payload: map[string]interface{}{},
			},
		},
	}
	data, err := json.Marshal(commands)
	assert.NoError(t, err)
	req, err := http.NewRequest("POST", "/commands", bytes.NewReader(data))

	assert.NoError(t, err)

	runChan := make(chan command.Command)

	uc := new(usecase.DockerUseCase)
	uc.On("Run", mock.Anything).Return(entity.NewFatalResult("err")).Run(func(args mock.Arguments) {
		cmd, ok := args.Get(0).(command.Command)
		assert.True(t, ok)
		runChan <- cmd
	}).Times(len(commands))

	rh := NewRestHandler(uc, logrus.New())

	recorder := httptest.NewRecorder()
	rh.AddCommands(recorder, req)

	for range commands {
		select {
		case <-runChan:
		case <-time.After(5 * time.Second):
			t.Fatal("Report did not happen within 5 seconds")
		}
	}
	uc.AssertExpectations(t)
	close(rh.(*restHandler).cmdChan)
}

func TestRestHandler_HealthCheck(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", bytes.NewReader([]byte{}))
	assert.NoError(t, err)

	rh := NewRestHandler(nil, logrus.New())
	recorder := httptest.NewRecorder()
	rh.HealthCheck(recorder, req)

	assert.Equal(t, "OK", recorder.Body.String())
}
