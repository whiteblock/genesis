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

	repo := new(repoMocks.AMQPRepository)
	repo.On("Consume", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Run(
		func(args mock.Arguments) {
			assert.True(t, args[0:len(args)-2].Is(conf.QueueName, conf.Consume.Consumer, conf.Consume.AutoAck,
				conf.Consume.Exclusive, conf.Consume.NoLocal, conf.Consume.NoWait))
		})

	serv, err := NewAMQPService(conf, repo)
	assert.NoError(t, err)

	_, err = serv.Consume()
	assert.NoError(t, err)

	repo.AssertNumberOfCalls(t, "Consume", 1)
}

func TestAMQPService_Requeue(t *testing.T) {
	conf := entity.AMQPConfig{
		QueueName: "test queue",
		Publish: entity.PublishConfig{
			Mandatory: true,
			Immediate: true,
		},
	}

	repo := new(repoMocks.AMQPRepository)
	repo.On("Requeue", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			assert.True(t, args[0:2].Is(conf.Publish.Mandatory, conf.Publish.Immediate))
		})

	serv, err := NewAMQPService(conf, repo)
	assert.NoError(t, err)

	err = serv.Requeue(amqp.Delivery{}, amqp.Publishing{})
	assert.NoError(t, err)

	repo.AssertNumberOfCalls(t, "Requeue", 1)
}

func TestAmqpService_CreateQueue(t *testing.T) {
	conf := entity.AMQPConfig{
		QueueName: "test queue",
		Queue: entity.QueueConfig{
			Durable: true,
			AutoDelete: false,
			Exclusive: false,
			NoWait: false,
			Args: map[string]interface{}{},
		},
	}

	repo := new(repoMocks.AMQPRepository)
	repo.On("CreateQueue", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			assert.True(t, args[0:5].Is(conf.QueueName, conf.Queue.Durable, conf.Queue.AutoDelete, conf.Queue.Exclusive, conf.Queue.NoWait, conf.Queue.Args))
		})

	serv, err := NewAMQPService(conf, repo)
	assert.NoError(t, err)

	err = serv.CreateQueue()
	assert.NoError(t, err)

	repo.AssertNumberOfCalls(t, "CreateQueue", 1)
}
