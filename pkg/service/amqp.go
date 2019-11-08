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
	"github.com/streadway/amqp"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/repository"
)

//AMQPService acts as a simple interface to the command queue
type AMQPService interface {
	//Consume immediately starts delivering queued messages.
	Consume() (<-chan amqp.Delivery, error)
	//Requeue rejects the oldMsg and queues the newMsg in a transaction
	Requeue(oldMsg amqp.Delivery, newMsg amqp.Publishing) error
	//CreateQueue attempts to publish a queue
	CreateQueue() error
}

type amqpService struct {
	repo repository.AMQPRepository
	conf entity.AMQPConfig
}

//NewAMQPService creates a new AMQPService
func NewAMQPService(conf entity.AMQPConfig, repo repository.AMQPRepository) (AMQPService, error) {
	return &amqpService{repo: repo, conf: conf}, nil
}

//Consume immediately starts delivering queued messages.
func (as amqpService) Consume() (<-chan amqp.Delivery, error) {
	return as.repo.Consume(as.conf.QueueName, as.conf.Consume.Consumer, as.conf.Consume.AutoAck,
		as.conf.Consume.Exclusive, as.conf.Consume.NoLocal, as.conf.Consume.NoWait,
		as.conf.Consume.Args)
}

//Requeue rejects the oldMsg and queues the newMsg in a transaction
func (as amqpService) Requeue(oldMsg amqp.Delivery, newMsg amqp.Publishing) error {
	return as.repo.Requeue(as.conf.Publish.Mandatory, as.conf.Publish.Immediate, oldMsg, newMsg)
}

//CreateQueue attempts to publish a queue
func (as amqpService) CreateQueue() error {
	return as.repo.CreateQueue(as.conf.QueueName, as.conf.Queue.Durable, as.conf.Queue.AutoDelete,
		as.conf.Queue.Exclusive, as.conf.Queue.NoWait, as.conf.Queue.Args)
}
