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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	externalsMock "github.com/whiteblock/genesis/mocks/pkg/externals"
	"testing"
)

func TestNewAMQPRepository(t *testing.T) {
	conn := new(externalsMock.AMQPConnection)
	repo, err := NewAMQPRepository(conn)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestAMQPRepository_GetChannel(t *testing.T) {
	conn := new(externalsMock.AMQPConnection)
	conn.On("Channel").Return(nil, nil).Once()
	repo, err := NewAMQPRepository(conn)
	assert.NoError(t, err)

	ch, err := repo.GetChannel()
	assert.NoError(t, err)
	assert.Nil(t, ch)
}

func TestAMQPRepository_RejectDelivery(t *testing.T) {
	msg := new(externalsMock.AMQPDelivery)
	msg.On("Reject", mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 1)
			assert.Equal(t, true, args.Get(0))
		}).Once()
	conn := new(externalsMock.AMQPConnection)
	repo, err := NewAMQPRepository(conn)
	assert.NoError(t, err)
	require.NotNil(t, repo)

	err = repo.RejectDelivery(msg, true)
	assert.NoError(t, err)
	msg.AssertExpectations(t)

}
