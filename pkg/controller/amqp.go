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

package controller

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/whiteblock/genesis/pkg/handler"
	"github.com/whiteblock/utility/utils"
	"golang.org/x/sync/semaphore"
	"sync"
)

//AMQPController is a controller which brings in from an AMQP compatible provider
type AMQPController interface {
	// Start starts the client. This function should be called only once and does not return
	Start()

	// CreateQueue creates the coresponding queue with the given parameters
	CreateQueue(durable, autoDelete, exclusive, noWait bool, args amqp.Table) error

	// Close cleans up the connections and resources used by this client
	Close()
}

type consumer struct {
	//queue is the name of the queue to consume from
	queue string
	//queueURL is the URL of the queue vhost
	queueURL string
	//handle handles the incoming deliveries
	handle handler.DeliveryHandler
	conn   *amqp.Connection
	once   *sync.Once
	sem    *semaphore.Weighted
}

func NewAMQPController(queue string, queueURL string, maxConcurreny int64, handle handler.DeliveryHandler) (AMQPController, error) {
	out := &consumer{
		queue:    queue,
		queueURL: queueURL,
		handle:   handle,
	}
	return out, out.init(maxConcurreny)
}

// CreateQueue creates the coresponding queue with the given parameters
func (c *consumer) CreateQueue(durable, autoDelete, exclusive, noWait bool, args amqp.Table) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return utils.LogError(err)
	}
	defer ch.Close()
	_, err = ch.QueueDeclare(c.queue, durable, autoDelete, exclusive, noWait, args)
	return utils.LogError(err)
}

// Close cleans up the connections and resources used by this client
func (c *consumer) Close() {
	if c == nil {
		return
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

// Start starts the client. This function should be called only once and does not return
func (c *consumer) Start() {
	c.once.Do(func() { c.loop() })
}

func (c *consumer) kickBackMessage(msg amqp.Delivery) {
	log.WithFields(log.Fields{"msg": msg}).Info("kicking back message")
	pub, err := c.handle.GetKickbackMessage(msg)
	if err != nil {
		utils.LogError(err)
		return
	}
	ch, err := c.conn.Channel()
	if err != nil {
		utils.LogError(err)
		return
	}
	defer ch.Close()
	err = ch.Tx()
	if err != nil {
		utils.LogError(err)
		return
	}

	err = ch.Publish(msg.Exchange, msg.RoutingKey, false, false, pub)
	if err != nil {
		ch.TxRollback()
		utils.LogError(err)
		return
	}
	err = msg.Reject(false)
	if err != nil {
		ch.TxRollback()
		utils.LogError(err)
		return
	}
	ch.TxCommit()
}

func (c *consumer) handleMessage(msg amqp.Delivery) {
	defer c.sem.Release(1)
	res := c.handle.ProcessMessage(msg)
	if res.IsRequeue() {
		utils.LogError(res.Error)
		go c.kickBackMessage(msg)
		return
	} else if res.IsSuccess() {

	} else {
		utils.LogError(msg.Ack(false))
	}

}

func (c *consumer) loop() {
	ch, err := c.conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	msgs, err := ch.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	for msg := range msgs {
		c.sem.Acquire(context.Background(), 1)
		go c.handleMessage(msg)
	}
}

func (c *consumer) init(maxConcurreny int64) (err error) {
	c.once = &sync.Once{}
	if maxConcurreny < 1 {
		return fmt.Errorf("maxConcurreny must be atleast 1")
	}
	c.sem = semaphore.NewWeighted(maxConcurreny)
	c.conn, err = amqp.Dial(c.queueURL)
	return utils.LogError(err)
}
