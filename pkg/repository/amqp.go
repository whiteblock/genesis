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
	"github.com/streadway/amqp"
)

//AMQPRepository represents functions available via a connection to a AMQP provider
type AMQPRepository interface {
	//Consume immediately starts delivering queued messages.
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	//Requeue rejects the oldMsg and queues the newMsg in a transaction
	Requeue(mandatory bool, immediate bool, oldMsg amqp.Delivery, newMsg amqp.Publishing) error
	//CreateQueue attempts to publish a queue
	CreateQueue(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) error
}

type amqpRepository struct {
	conn *amqp.Connection
}

//NewAMQPRepository creates a new AMQPRepository
func NewAMQPRepository(conn *amqp.Connection) (AMQPRepository, error) {
	return &amqpRepository{conn: conn}, nil
}

//Consume immediately starts delivering queued messages.
func (ar amqpRepository) Consume(queue, consumer string, autoAck, exclusive,
	noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	ch, err := ar.conn.Channel()
	if err != nil {
		return nil, err
	}
	msgs, err := ch.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	return msgs, err
}

//Requeue rejects the oldMsg and queues the newMsg in a transaction
func (ar amqpRepository) Requeue(mandatory bool, immediate bool, oldMsg amqp.Delivery, newMsg amqp.Publishing) error {
	ch, err := ar.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	err = ch.Tx()
	if err != nil {
		return err
	}

	err = ch.Publish(oldMsg.Exchange, oldMsg.RoutingKey, mandatory, immediate, newMsg)
	if err != nil {
		ch.TxRollback()
		return err
	}

	err = oldMsg.Reject(false)
	if err != nil {
		ch.TxRollback()
		return err
	}
	return ch.TxCommit()
}

//CreateQueue attempts to publish a queue
func (ar amqpRepository) CreateQueue(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) error {
	ch, err := ar.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	_, err = ch.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
	return err
}
