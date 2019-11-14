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

package repository

import (
	"github.com/stretchr/testify/assert"
	"github.com/whiteblock/genesis/pkg/entity"
	"testing"
)

func TestCommandRepository_Local(t *testing.T) {
	repo := NewLocalCommandRepository()
	err := repo.ReportCommandFinished("test", entity.NewSuccessResult())
	assert.NoError(t, err)
	err = repo.ReportCommandFinished("test2", entity.NewSuccessResult())
	assert.NoError(t, err)
	err = repo.ReportCommandFinished("test3", entity.NewSuccessResult())
	assert.NoError(t, err)

	executed, err := repo.HasCommandExecuted("test")
	assert.NoError(t, err)
	assert.True(t, executed)

	executed, err = repo.HasCommandExecuted("test3")
	assert.NoError(t, err)
	assert.True(t, executed)

	executed, err = repo.HasCommandExecuted("test4")
	assert.NoError(t, err)
	assert.False(t, executed)
}
