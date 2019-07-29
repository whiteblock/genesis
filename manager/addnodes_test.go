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

package manager

import (
	"strconv"
	"testing"

	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/util"
)

// TODO this doesn't work
func TestAddNodes(t *testing.T) {
	var tests = []struct {
		details *db.DeploymentDetails
		testnetID string
	}{
		{
			details: &db.DeploymentDetails{
				ID: "10",
				Servers: []int{},
				Blockchain: "eos",
				Nodes: 3,
				Images: []string{},
				Params: map[string]interface{}{},
				Resources: []util.Resources{},
				Environments: []map[string]string{},
				Files: []map[string]string{},
				Logs: []map[string]string{},
				Extras: map[string]interface{}{},
			},
			testnetID: "10",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if AddNodes(tt.details, tt.testnetID) != nil {
				t.Error("return value of AddNodes does not match expected value")
			}
		})
	}
}