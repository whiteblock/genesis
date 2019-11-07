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

package usecase

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	mocks "github.com/whiteblock/genesis/mocks/pkg/service"

	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"
	"testing"
	"time"
)

func TestDockerUseCase_Execute_CreateContainer(t *testing.T) {
	service := new(mocks.DockerService)
	service.On("CreateClient", mock.Anything, mock.Anything).Return(nil, nil)
	service.On("CreateContainer", mock.Anything, mock.Anything, mock.Anything).Return(entity.Result{Type: entity.SuccessType})

	usecase, _ := NewDockerUseCase(entity.DockerConfig{}, service)

	res := usecase.Execute(context.TODO(), command.Command{
		ID:        "TEST",
		Timestamp: 1234567,
		Timeout:   5 * time.Second,
		Target:    command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "createContainer",
			Payload: map[string]interface{}{},
		},
	})
	assert.Equal(t, res.Error, nil)
	assert.True(t, service.AssertNumberOfCalls(t, "CreateContainer", 1))
}
