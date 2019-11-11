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
	"github.com/whiteblock/genesis/pkg/service"
	"github.com/whiteblock/utility/utils"
	"golang.org/x/sync/semaphore"
	"sync"
)

//CommandController is a controller which brings in from an AMQP compatible provider
type CommandController interface {
	// Start starts the client. This function should be called only once and does not return
	Start()
}

type consumer struct {
	serv   service.AMQPService
	handle handler.DeliveryHandler
	once   *sync.Once
	sem    *semaphore.Weighted
}

//NewCommandController creates a new CommandController
func NewCommandController(maxConcurreny int64, serv service.AMQPService, handle handler.DeliveryHandler) (CommandController, error) {
	if maxConcurreny < 1 {
		return nil, fmt.Errorf("maxConcurreny must be atleast 1")
	}
	out := &consumer{
		serv:   serv,
		handle: handle,
		once:   &sync.Once{},
		sem:    semaphore.NewWeighted(maxConcurreny),
	}
	err := out.serv.CreateQueue()
	log.WithFields(log.Fields{"err": err}).Debug("attempted to create queue on start")
	return out, nil
}

// Start starts the client. This function should be called only once and does not return
func (c *consumer) Start() {
	c.once.Do(func() { c.loop() })
}

func (c *consumer) kickBackMessage(msg amqp.Delivery) error {
	log.WithFields(log.Fields{"msg": msg}).Info("kicking back message")
	pub, err := c.handle.GetKickbackMessage(msg)
	if err != nil {
		return err
	}
	return c.serv.Requeue(msg, pub)
}

func (c *consumer) handleMessage(msg amqp.Delivery) {
	defer c.sem.Release(1)
	res := c.handle.ProcessMessage(msg)
	if res.IsRequeue() {
		utils.LogError(res.Error)
		c.kickBackMessage(msg)
		return
	} else if res.IsSuccess() {
		msg.Ack(false)
		//TODO: Status report
	} else {
		utils.LogError(msg.Ack(false))
		//TODO: Fatal command error handling
	}
}

func (c *consumer) loop() {
	msgs, err := c.serv.Consume()
	if err != nil {
		log.Fatal(err)
	}
	for msg := range msgs {
		c.sem.Acquire(context.Background(), 1)
		go c.handleMessage(msg)
	}
}
