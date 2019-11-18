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

package externals

import (
	"github.com/streadway/amqp"
)

//AMQPConnection represents the needed functionality from a amqp.Connection
type AMQPConnection interface {
	Channel() (*amqp.Channel, error)
	Close() error
}

type amqpConnection struct {
	ch *amqp.Channel
}
