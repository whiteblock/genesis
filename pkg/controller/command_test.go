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

	queue "github.com/whiteblock/genesis/mocks/amqp"
	handler "github.com/whiteblock/genesis/mocks/pkg/handler"
	"github.com/whiteblock/genesis/pkg/entity"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewCommandController_Failure(t *testing.T) {
	ctl, err := NewCommandController(0, nil, nil, nil, logrus.New())
	assert.Nil(t, ctl)
	assert.Error(t, err)
}

func TestNewCommandController_Ignore_CreateQueueFailure(t *testing.T) {
	serv := new(queue.AMQPService)
	serv2 := new(queue.AMQPService)
	serv.On("CreateQueue").Return(fmt.Errorf("err")).Once()
	serv2.On("CreateQueue").Return(fmt.Errorf("err")).Once()

	control, err := NewCommandController(2, serv, serv2, nil, logrus.New())
	assert.NotNil(t, control)
	assert.NoError(t, err)

	serv.AssertExpectations(t)
	serv2.AssertExpectations(t)
}

func TestCommandController_Consumption(t *testing.T) {
	items := 10

	processedChan := make(chan bool, items)
	deliveryChan := make(chan amqp.Delivery, items)
	serv := new(queue.AMQPService)
	serv2 := new(queue.AMQPService)
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

func TestCommandController_ConsumptionAllDone(t *testing.T) {
	items := 10

	processedChan := make(chan bool, items)
	deliveryChan := make(chan amqp.Delivery, items)
	serv := new(queue.AMQPService)
	serv2 := new(queue.AMQPService)
	serv.On("Consume").Return((<-chan amqp.Delivery)(deliveryChan), nil).Once()
	serv.On("CreateQueue").Return(nil).Once()
	serv2.On("CreateQueue").Return(nil).Once()
	serv2.On("Send", mock.Anything).Return(nil).Times(items).Run(func(_ mock.Arguments) {
		processedChan <- true
	})

	hand := new(handler.DeliveryHandler)
	hand.On("Process", mock.Anything).Return(amqp.Publishing{}, entity.NewAllDoneResult()).Times(items)

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

func TestCommandController_ConsumptionAllDone_Send_Err(t *testing.T) {
	items := 10

	processedChan := make(chan bool, items)
	deliveryChan := make(chan amqp.Delivery, items)
	serv := new(queue.AMQPService)
	serv2 := new(queue.AMQPService)
	serv.On("Consume").Return((<-chan amqp.Delivery)(deliveryChan), nil).Once()
	serv.On("CreateQueue").Return(nil).Once()
	serv2.On("CreateQueue").Return(nil).Once()
	serv2.On("Send", mock.Anything).Return(fmt.Errorf("err")).Times(items).Run(func(_ mock.Arguments) {
		processedChan <- true
	})

	hand := new(handler.DeliveryHandler)
	hand.On("Process", mock.Anything).Return(amqp.Publishing{}, entity.NewAllDoneResult()).Times(items)

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
	serv := new(queue.AMQPService)
	serv.On("Consume").Return((<-chan amqp.Delivery)(deliveryChan), nil).Once()
	serv.On("CreateQueue").Return(nil).Once()
	serv.On("Requeue", mock.Anything, mock.Anything).Run(func(_ mock.Arguments) {
		processedChan <- true
	}).Return(nil).Times(items)
	serv2 := new(queue.AMQPService)
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
