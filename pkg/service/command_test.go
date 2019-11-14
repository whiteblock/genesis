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
	along with this program. If not, see <https://www.gnu.org/licenses/>.
*/

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	repoMocks "github.com/whiteblock/genesis/mocks/pkg/repository"
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"
)

func TestNewCommandService(t *testing.T) {
	repo := new(repoMocks.CommandRepository)

	expected := &commandService{
		repo: repo,
	}

	assert.Equal(t, expected, NewCommandService(repo))
}

func TestCommandService_CheckDependenciesExecuted(t *testing.T) {
	repo := new(repoMocks.CommandRepository)
	repo.On("HasCommandExecuted", "test1").Return(true, nil)
	repo.On("HasCommandExecuted", "test2").Return(false, nil)

	serv := NewCommandService(repo)

	ready, err := serv.CheckDependenciesExecuted(command.Command{
		ID:      "TEST",
		Timeout: 0,
		Target:  command.Target{IP: "0.0.0.0"},
		Order: command.Order{
			Type:    "createContainer",
			Payload: map[string]interface{}{},
		},
		Dependencies: []string{"test1", "test2"},
	})
	assert.NoError(t, err)
	assert.False(t, ready)
}

func TestCommandService_ReportCommandResult(t *testing.T) {
	repo := new(repoMocks.CommandRepository)

	cmd := command.Command{
		ID: "test",
	}
	res := entity.Result{
		Error: nil,
		Type:  entity.SuccessType,
	}

	repo.On("ReportCommandFinished", mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 2)
			assert.Equal(t, cmd.ID, args.Get(0))
			assert.Equal(t, res, args.Get(1))
		})

	serv := NewCommandService(repo)

	err := serv.ReportCommandResult(cmd, res)
	assert.NoError(t, err)

	assert.True(t, repo.AssertNumberOfCalls(t, "ReportCommandFinished", 1))
}
