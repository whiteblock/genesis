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
	"fmt"
	"testing"
	"time"

	handler "github.com/whiteblock/genesis/mocks/pkg/handler"
	service "github.com/whiteblock/genesis/mocks/pkg/service"
	"github.com/whiteblock/genesis/pkg/entity"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCommandController_Consumption(t *testing.T) {
	items := 10

	processedChan := make(chan bool, items)
	deliveryChan := make(chan amqp.Delivery, items)
	serv := new(service.AMQPService)
	serv2 := new(service.AMQPService)
	serv.On("Consume").Return((<-chan amqp.Delivery)(deliveryChan), nil).Once()
	serv.On("CreateQueue").Return(nil).Once()
	serv2.On("CreateQueue").Return(nil).Once()

	hand := new(handler.DeliveryHandler)
	hand.On("Process", mock.Anything).Run(func(_ mock.Arguments) {
		processedChan <- true
	}).Return(amqp.Publishing{}, entity.NewSuccessResult()).Times(items)

	control, err := NewCommandController(2, serv, serv2, hand, logrus.New())
	assert.Equal(t, err, nil)
	go control.Start()

	for i := 0; i < items; i++ {
		deliveryChan <- amqp.Delivery{}
	}
	for i := 0; i < items; i++ {
		select {
		case <-processedChan:
		case <-time.After(5 * time.Second):
			t.Fatal("messages were not consumed within 5 seconds")
		}
	}

	close(deliveryChan)
	hand.AssertExpectations(t)
	serv.AssertExpectations(t)
	serv2.AssertExpectations(t)
}

func TestCommandController_Requeue(t *testing.T) {
	items := 10

	processedChan := make(chan bool, items)
	deliveryChan := make(chan amqp.Delivery, items)
	serv := new(service.AMQPService)
	serv.On("Consume").Return((<-chan amqp.Delivery)(deliveryChan), nil).Once()
	serv.On("CreateQueue").Return(nil).Once()
	serv.On("Requeue", mock.Anything, mock.Anything).Run(func(_ mock.Arguments) {
		processedChan <- true
	}).Return(nil).Times(items)
	serv2 := new(service.AMQPService)
	serv2.On("CreateQueue").Return(nil).Once()

	hand := new(handler.DeliveryHandler)
	hand.On("Process", mock.Anything).Return(amqp.Publishing{},
		entity.NewErrorResult(fmt.Errorf("some non-fatal error"))).Times(items)

	control, err := NewCommandController(2, serv, serv2, hand, logrus.New())
	assert.Equal(t, err, nil)
	go control.Start()

	for i := 0; i < items; i++ {
		deliveryChan <- amqp.Delivery{}
	}
	for i := 0; i < items; i++ {
		select {
		case <-processedChan:
		case <-time.After(5 * time.Second):
			t.Fatal("messages were not consumed within 5 seconds")
		}
	}

	close(deliveryChan)
	hand.AssertExpectations(t)
	serv.AssertExpectations(t)
}
