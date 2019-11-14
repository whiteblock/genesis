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

package entity

import (
	"github.com/streadway/amqp"
)

//QueueConfig is the configuration of the queue to be created if the queue does not yet exist
type QueueConfig struct {
	//Durable sets the queue to be persistent
	Durable bool
	//AutoDelete tells the queue to drop messages if there are not any consumers
	AutoDelete bool
	//Exclusive queues are only accessible by the connection that declares them
	Exclusive bool
	//NoWait is true, the queue will assume to be declared on the server
	NoWait bool
	//Args contains addition arguments to be provided
	Args amqp.Table
}

//ConsumeConfig is the configuration of the consumer of messages from the queue
type ConsumeConfig struct {
	//Consumer is the name of the consumer
	Consumer string
	//AutoAck causes the server to acknowledge deliveries to this consumer prior to writing the delivery to the network
	AutoAck bool
	//Exclusive: when true, the server will ensure that this is the sole consumer from this queue
	//This should always be false.
	Exclusive bool
	//NoLocal is not supported by RabbitMQ
	NoLocal bool
	//NoWait: do not wait for the server to confirm the request and immediately begin deliveries
	NoWait bool
	//Args contains addition arguments to be provided
	Args amqp.Table
}

//PublishConfig  is the configuration for the publishing of messages
type PublishConfig struct {
	Mandatory bool
	Immediate bool
}

//AMQPConfig is the configuration for AMQP
type AMQPConfig struct {
	//QueueName the name of the queue to connect to
	QueueName string
	//Queue is the configuration for the queue
	Queue QueueConfig
	//Consume is the configuration of the consumer
	Consume ConsumeConfig
	//Publish is the configuration for the publishing of messages
	Publish PublishConfig
}
