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
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package utility

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

//AMQPMessage contains utilities for manipulating AMQP messages
type AMQPMessage interface {
	CreateMessage(body interface{}) (amqp.Publishing, error)
	//GetKickbackMessage takes the delivery and creates a message from it
	//for requeuing on non-fatal error
	GetKickbackMessage(msg amqp.Delivery) (amqp.Publishing, error)
	GetNextMessage(msg amqp.Delivery, body interface{}) (amqp.Publishing, error)
}

type amqpMessage struct {
	maxRetries int64
}

//NewAMQPMessage creates a new AMQPMessage
func NewAMQPMessage(maxRetries int64) AMQPMessage {
	return &amqpMessage{maxRetries: maxRetries}
}

func (dh amqpMessage) CreateMessage(body interface{}) (amqp.Publishing, error) {
	rawBody, err := json.Marshal(body)
	if err != nil {
		return amqp.Publishing{}, err
	}

	pub := amqp.Publishing{
		Headers: map[string]interface{}{
			"retryCount": int64(0),
		},
		Body: rawBody,
	}
	return pub, nil
}

// GetNextMessage is similar to GetKickbackMessage but takes in a new body, and does not increment the
// retry count
func (dh amqpMessage) GetNextMessage(msg amqp.Delivery, body interface{}) (amqp.Publishing, error) {
	rawBody, err := json.Marshal(body)
	if err != nil {
		return amqp.Publishing{}, err
	}
	pub := amqp.Publishing{
		Headers: msg.Headers,
		// Properties
		ContentType:     msg.ContentType,
		ContentEncoding: msg.ContentEncoding,
		DeliveryMode:    msg.DeliveryMode,
		Type:            msg.Type,
		Body:            rawBody,
	}
	if pub.Headers == nil {
		pub.Headers = map[string]interface{}{}
	}
	pub.Headers["retryCount"] = int64(0) //reset retry count

	return pub, nil
}

// GetKickbackMessage takes the delivery and creates a message from it
// for requeuing on non-fatal error. It returns an error if
func (am amqpMessage) GetKickbackMessage(msg amqp.Delivery) (amqp.Publishing, error) {
	pub := amqp.Publishing{
		Headers: msg.Headers,
		// Properties
		ContentType:     msg.ContentType,
		ContentEncoding: msg.ContentEncoding,
		DeliveryMode:    msg.DeliveryMode,
		Priority:        msg.Priority,
		CorrelationId:   msg.CorrelationId,
		ReplyTo:         msg.ReplyTo,
		Expiration:      msg.Expiration,
		MessageId:       msg.MessageId,
		Timestamp:       msg.Timestamp,
		Type:            msg.Type,
		Body:            msg.Body,
	}
	if pub.Headers == nil {
		pub.Headers = map[string]interface{}{}
	}
	_, exists := pub.Headers["retryCount"]
	if !exists {
		pub.Headers["retryCount"] = int64(0)
	}
	if pub.Headers["retryCount"].(int64) > am.maxRetries {
		return amqp.Publishing{}, errors.New("too many retries")
	}
	pub.Headers["retryCount"] = pub.Headers["retryCount"].(int64) + 1
	return pub, nil
}
