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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	service "github.com/whiteblock/genesis/mocks/pkg/service"
	usecase "github.com/whiteblock/genesis/mocks/pkg/usecase"
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
	uc.On("Run", mock.Anything).Return(entity.NewSuccessResult())

	serv := new(service.CommandService)
	serv.On("ReportCommandResult", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		cmd, ok := args.Get(0).(command.Command)
		assert.True(t, ok)
		runChan <- cmd
	}).Return(nil)

	rh := NewRestHandler(uc, serv)

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
