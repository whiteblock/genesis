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

package rabbitmq

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/whiteblock/genesis/pkg/handler"
	"golang.org/x/sync/semaphore"
	"sync"
)

type Consumer struct {
	Queue         string
	QueueURL      string
	MaxConcurreny int64
	MaxRetries    int64
	Handle        handler.DeliveryHandler
	conn          *amqp.Connection
	once          *sync.Once
	sem           *semaphore.Weighted
}

func (c *Consumer) Init() error {
	c.once = &sync.Once{}
	if c.MaxConcurreny < 1 {
		return fmt.Errorf("MaxConcurreny must be atleast 1")
	}
	c.sem = semaphore.NewWeighted(c.MaxConcurreny)
	if c.MaxRetries < 1 {
		return fmt.Errorf("MaxRetries must be atleast 1")
	}
	c.callback = callback
	return c.init()
}

// CreateQueue creates the coresponding queue with the given parameters
func (c *Consumer) CreateQueue(durable, autoDelete, exclusive, noWait bool, args amqp.Table) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return utils.LogError(err)
	}
	defer ch.Close()
	_, err = ch.QueueDeclare(c.Queue, durable, autoDelete, exclusive, noWait, args)
	return utils.LogError(err)
}

// Close cleans up the connections and resources used by this client
func (c *Consumer) Close() {
	if c == nil {
		return
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

// Run starts the client. This function should be called only once and does not return
func (c *Consumer) Run() {
	c.once.Do(func() { c.loop() })
}

func (c *Consumer) kickBackMessage(msg amqp.Delivery) {
	pub, err := c.Handle.GetKickbackMessage(msg)
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
		tx.Rollback()
		utils.LogError(err)
		return
	}
	err = msg.Reject(false)
	if err != nil {
		tx.Rollback()
		utils.LogError(err)
		return
	}
	ch.TxCommit()
}

func (c *Consumer) handleMessage(msg amqp.Delivery) {
	defer c.sem.Release(1)
	res := c.ProcessMessage(msg)
	if res.IsRequeue() {
		utils.LogError(err)
		go c.kickBackMessage(msg)
		return
	} else if res.IsSuccess() {

	} else {
		utils.LogError(msg.Ack(false))
	}

}

func (c *Consumer) loop() {
	ch, err := c.conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	msgs, err := ch.Consume(c.Queue, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	for msg := range msgs {
		c.sem.Acquire(context.Background(), 1)
		go c.handleMessage(msg)
	}
}

func (c *Consumer) init() (err error) {
	c.conn, err = amqp.Dial(c.QueueURL)
	return utils.LogError(err)
}
