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

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	externalsMocks "github.com/whiteblock/genesis/mocks/pkg/externals"
	repoMocks "github.com/whiteblock/genesis/mocks/pkg/repository"
	"github.com/whiteblock/genesis/pkg/entity"
)

func TestNewAMQPService(t *testing.T) {
	conf := entity.AMQPConfig{
		QueueName: "test queue",
	}
	repo := new(repoMocks.AMQPRepository)

	serv, err := NewAMQPService(conf, repo)
	assert.NoError(t, err)

	expectedServ := &amqpService{
		repo: repo,
		conf: conf,
	}

	assert.Equal(t, serv, expectedServ)
}

func TestAMQPService_Consume(t *testing.T) {
	conf := entity.AMQPConfig{
		QueueName: "test queue",
		Consume: entity.ConsumeConfig{
			Consumer:  "test",
			AutoAck:   false,
			Exclusive: false,
			NoLocal:   false,
			NoWait:    false,
			Args:      nil,
		},
	}
	ch := new(externalsMocks.AMQPChannel)

	ch.On("Consume", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Run(
		func(args mock.Arguments) {
			assert.True(t, args[0:len(args)-2].Is(conf.QueueName, conf.Consume.Consumer, conf.Consume.AutoAck,
				conf.Consume.Exclusive, conf.Consume.NoLocal, conf.Consume.NoWait))
		}).Once()
	repo := new(repoMocks.AMQPRepository)
	repo.On("GetChannel").Return(ch, nil).Once()

	serv, err := NewAMQPService(conf, repo)
	assert.NoError(t, err)

	_, err = serv.Consume()
	assert.NoError(t, err)

	repo.AssertExpectations(t)
	ch.AssertExpectations(t)
}

func TestAMQPService_Requeue_Success(t *testing.T) {
	conf := entity.AMQPConfig{
		QueueName: "test queue",
		Publish: entity.PublishConfig{
			Mandatory: true,
			Immediate: true,
		},
	}

	oldMsg := amqp.Delivery{
		Exchange:   "t",
		RoutingKey: "1",
	}
	newMsg := amqp.Publishing{}

	ch := new(externalsMocks.AMQPChannel)
	ch.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 5)
			assert.Equal(t, oldMsg.Exchange, args.Get(0))
			assert.Equal(t, oldMsg.RoutingKey, args.Get(1))
			assert.Equal(t, conf.Publish.Mandatory, args.Get(2))
			assert.Equal(t, conf.Publish.Immediate, args.Get(3))
			assert.Equal(t, newMsg, args.Get(4))
		}).Once()
	ch.On("Tx").Return(nil).Once()
	ch.On("Close").Return(nil).Once()
	ch.On("TxCommit").Return(nil).Once()

	repo := new(repoMocks.AMQPRepository)
	repo.On("GetChannel").Return(ch, nil).Once()
	repo.On("RejectDelivery", mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 2)
			assert.Equal(t, oldMsg, args.Get(0))
			assert.False(t, args.Bool(1))
		})

	serv, err := NewAMQPService(conf, repo)
	assert.NoError(t, err)

	err = serv.Requeue(oldMsg, amqp.Publishing{})
	assert.NoError(t, err)

	repo.AssertExpectations(t)
	ch.AssertExpectations(t)
}

func TestAmqpService_CreateQueue(t *testing.T) {
	conf := entity.AMQPConfig{
		QueueName: "test queue",
		Queue: entity.QueueConfig{
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args:       map[string]interface{}{},
		},
	}

	ch := new(externalsMocks.AMQPChannel)
	ch.On("QueueDeclare", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(amqp.Queue{}, nil).Run(
		func(args mock.Arguments) {
			require.Len(t, args, 6)
			assert.True(t, args[0:5].Is(conf.QueueName, conf.Queue.Durable, conf.Queue.AutoDelete, conf.Queue.Exclusive, conf.Queue.NoWait, conf.Queue.Args))
		}).Once()
	ch.On("Close").Return(nil).Once()
	repo := new(repoMocks.AMQPRepository)
	repo.On("GetChannel").Return(ch, nil).Once()
	serv, err := NewAMQPService(conf, repo)
	assert.NoError(t, err)

	err = serv.CreateQueue()
	assert.NoError(t, err)

	repo.AssertExpectations(t)
	ch.AssertExpectations(t)
}
