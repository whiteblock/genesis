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

package command

import (
	"encoding/json"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"testing"
)

func TestDeserSerRoundtripCommand(t *testing.T) {
	command := Command{
		ID:           "",
		Timestamp:    0,
		Timeout:      0,
		Retry:        0,
		Target:       Target{},
		Dependencies: nil,
		Order: Order{
			Type:    Startcontainer,
			Payload: SimpleName{Name: "test"},
		},
	}
	bytes, err := json.Marshal(command)
	if err != nil {
		t.Fatal(err)
	}
	read := Command{}
	err = json.Unmarshal(bytes, &read)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(command, read) {
		t.Fatal("cannot read back command")
	}

	payload := SimpleName{}
	err = mapstructure.Decode(read.Order.Payload, &payload)
	if err != nil {
		t.Fatal(err)
	}
	if payload.Name != "test" {
		t.Fatal("cannot read back payload name")
	}
}
