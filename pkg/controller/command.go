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
	"sync"

	"github.com/whiteblock/genesis/pkg/handler"
	"github.com/whiteblock/genesis/pkg/service"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"golang.org/x/sync/semaphore"
)

// CommandController is a controller which brings in from an AMQP compatible provider
type CommandController interface {
	// Start starts the client. This function should be called only once and does not return
	Start()
}

type consumer struct {
	completion service.AMQPService
	cmds       service.AMQPService
	handle     handler.DeliveryHandler
	log        logrus.Ext1FieldLogger
	once       *sync.Once
	sem        *semaphore.Weighted
}

// NewCommandController creates a new CommandController
func NewCommandController(
	maxConcurreny int64,
	cmds service.AMQPService,
	completion service.AMQPService,
	handle handler.DeliveryHandler,
	log logrus.Ext1FieldLogger) (CommandController, error) {

	if maxConcurreny < 1 {
		return nil, fmt.Errorf("maxConcurreny must be atleast 1")
	}
	out := &consumer{
		log:        log,
		completion: completion,
		cmds:       cmds,
		handle:     handle,
		once:       &sync.Once{},
		sem:        semaphore.NewWeighted(maxConcurreny),
	}
	err := out.cmds.CreateQueue()
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Debug("failed attempt to create queue on start")
	}

	err = out.completion.CreateQueue()
	if err != nil {
		log.WithFields(logrus.Fields{"err": err}).Debug("failed attempt to create queue on start")
	}

	return out, nil
}

// Start starts the client. This function should be called only once and does not return
func (c *consumer) Start() {
	c.once.Do(func() { c.loop() })
}

func (c *consumer) handleMessage(msg amqp.Delivery) {
	defer c.sem.Release(1)

	pub, res := c.handle.Process(msg)
	if res.IsRequeue() {
		err := c.cmds.Requeue(msg, pub)
		if err != nil {
			c.log.WithField("err", err).Error("failed to re-queue")
		}
		return
	}
	if res.IsAllDone() {
		err := c.completion.Send(pub)
		if err != nil {
			c.log.WithField("err", err).Error("failed to send to the completion queue")
			return
		}
	}
	msg.Ack(false)
}

func (c *consumer) loop() {
	msgs, err := c.cmds.Consume()
	if err != nil {
		c.log.Fatal(err)
	}
	for msg := range msgs {
		c.log.Trace("recieved a message")
		c.sem.Acquire(context.Background(), 1)
		go c.handleMessage(msg)
	}
}
