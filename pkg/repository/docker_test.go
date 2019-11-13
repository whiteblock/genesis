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

package repository

import (
	"testing"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	entityMock "github.com/whiteblock/genesis/mocks/pkg/entity"
)

func TestDockerRepository_VolumeList(t *testing.T) {
	cli := new(entityMock.Client)
	testFilters := filters.Args{}
	result := volume.VolumeListOKBody{}

	cli.On("VolumeList", mock.Anything, mock.Anything).Return(result, nil).Run(func(args mock.Arguments) {
		require.Len(t, args, 2)
		assert.Nil(t, args.Get(0))
		assert.Equal(t, testFilters, args.Get(1))
	})

	repo := NewDockerRepository()
	res, err := repo.VolumeList(nil, cli, testFilters)
	assert.NoError(t, err)
	assert.Equal(t, result, res)

	cli.AssertExpectations(t)
}

func TestDockerRepository_VolumeRemove(t *testing.T) {
	cli := new(entityMock.Client)
	volumeID := "volume1"
	isForced := true

	cli.On("VolumeRemove", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		require.Len(t, args, 3)
		assert.Nil(t, args.Get(0))
		assert.Equal(t, volumeID, args.String(1))
		assert.Equal(t, isForced, args.Bool(2))
	}).Once()

	repo := NewDockerRepository()

	err := repo.VolumeRemove(nil, cli, volumeID, isForced)
	assert.NoError(t, err)
	cli.AssertExpectations(t)
}
