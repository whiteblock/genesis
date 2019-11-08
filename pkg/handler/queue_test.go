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

package handler

import (
	"encoding/json"
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"
	mocksHandler "github.com/whiteblock/genesis/pkg/handler/mocks"
)

func TestDeliveryHandler_ProcessMessage(t *testing.T) {
	dh := new(mocksHandler.DeliveryHandler)
	dh.On("ProcessMessage", mock.Anything).Return(entity.Result{Error: nil, Type: entity.SuccessType})

	cmd := new(command.Command)
	body, err := json.Marshal(cmd)
	if err != nil {
		t.Error(err)
	}

	res := dh.ProcessMessage(amqp.Delivery{
		Body: body,
	})
	assert.Equal(t, res.Error, nil)
}

func TestDeliveryHandler_GetKickbackMessage(t *testing.T) {
	
}
