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
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/whiteblock/genesis/pkg/config"
	"github.com/whiteblock/genesis/pkg/repository"
)

//AMQPService acts as a simple interface to the command queue
type AMQPService interface {
	//Consume immediately starts delivering queued messages.
	Consume() (<-chan amqp.Delivery, error)
	//send places a message into the queue
	Send(pub amqp.Publishing) error
	//Requeue rejects the oldMsg and queues the newMsg in a transaction
	Requeue(oldMsg amqp.Delivery, newMsg amqp.Publishing) error
	//CreateQueue attempts to publish a queue
	CreateQueue() error
}

type amqpService struct {
	repo repository.AMQPRepository
	conf config.AMQP
	log  logrus.Ext1FieldLogger
}

//NewAMQPService creates a new AMQPService
func NewAMQPService(
	conf config.AMQP,
	repo repository.AMQPRepository,
	log logrus.Ext1FieldLogger) AMQPService {

	return &amqpService{repo: repo, conf: conf, log: log}
}

func (as amqpService) Send(pub amqp.Publishing) error {
	ch, err := as.repo.GetChannel()
	if err != nil {
		return err
	}
	defer ch.Close()
	as.log.WithFields(logrus.Fields{
		"exchange": as.conf.Publish.Exchange,
		"queue": as.conf.QueueName,
	}).Debug("publishing a message")
	return ch.Publish(as.conf.Publish.Exchange, "", as.conf.Publish.Mandatory, as.conf.Publish.Immediate, pub)
}

//Consume immediately starts delivering queued messages.
func (as amqpService) Consume() (<-chan amqp.Delivery, error) {
	ch, err := as.repo.GetChannel()
	if err != nil {
		return nil, err
	}
	as.log.WithFields(logrus.Fields{
		"queue":as.conf.QueueName,
		"consumer":as.conf.Consume.Consumer,
	}).Trace("consuming")
	return ch.Consume(as.conf.QueueName, as.conf.Consume.Consumer, as.conf.Consume.AutoAck,
		as.conf.Consume.Exclusive, as.conf.Consume.NoLocal, as.conf.Consume.NoWait,
		as.conf.Consume.Args)
}

//Requeue rejects the oldMsg and queues the newMsg in a transaction
func (as amqpService) Requeue(oldMsg amqp.Delivery, newMsg amqp.Publishing) error {
	ch, err := as.repo.GetChannel()
	if err != nil {
		return err
	}
	defer ch.Close()
	err = ch.Tx()
	if err != nil {
		return err
	}

	err = ch.Publish(oldMsg.Exchange, oldMsg.RoutingKey, as.conf.Publish.Mandatory, as.conf.Publish.Immediate, newMsg)
	if err != nil {
		ch.TxRollback()
		return err
	}

	err = as.repo.RejectDelivery(oldMsg, false)
	if err != nil {
		ch.TxRollback()
		return err
	}
	return ch.TxCommit()
}

//CreateQueue attempts to publish a queue
func (as amqpService) CreateQueue() error {
	ch, err := as.repo.GetChannel()
	if err != nil {
		return err
	}
	defer ch.Close()
	_, err = ch.QueueDeclare(as.conf.QueueName, as.conf.Queue.Durable, as.conf.Queue.AutoDelete,
		as.conf.Queue.Exclusive, as.conf.Queue.NoWait, as.conf.Queue.Args)
	return err
}
