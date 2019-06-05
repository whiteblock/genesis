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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/util"
)

//TODO idk how to test this one
func Test_validateResources(t *testing.T) {
	details := &db.DeploymentDetails{
		ID:           "123",
		Servers:      []int{4, 5, 6},
		Blockchain:   "eos",
		Nodes:        2,
		Images:       []string{},
		Params:       map[string]interface{}{},
		Resources:    []util.Resources{{Cpus: " ", Memory: " ", Volumes: []string{}, Ports: []string{}}},
		Environments: []map[string]string{},
		Files:        []map[string]string{},
		Logs:         []map[string]string{},
		Extras:       map[string]interface{}{},
	}

	validateResources(details)

	fmt.Println()
}

func Test_validateNumOfNodes(t *testing.T) {
	var test = []struct {
		details  *db.DeploymentDetails
		expected error
	}{
		{
			details: &db.DeploymentDetails{
				ID:           "123",
				Servers:      []int{4, 5, 6},
				Blockchain:   "eos",
				Nodes:        100,
				Images:       []string{},
				Params:       map[string]interface{}{},
				Resources:    []util.Resources{{Cpus: " ", Memory: " ", Volumes: []string{}, Ports: []string{}}},
				Environments: []map[string]string{},
				Files:        []map[string]string{},
				Logs:         []map[string]string{},
				Extras:       map[string]interface{}{},
			},
			expected: nil,
		},
		{
			details: &db.DeploymentDetails{
				ID:           "123",
				Servers:      []int{4, 5, 6},
				Blockchain:   "eos",
				Nodes:        1000,
				Images:       []string{},
				Params:       map[string]interface{}{},
				Resources:    []util.Resources{{Cpus: " ", Memory: " ", Volumes: []string{}, Ports: []string{}}},
				Environments: []map[string]string{},
				Files:        []map[string]string{},
				Logs:         []map[string]string{},
				Extras:       map[string]interface{}{},
			},
			expected: errors.New("too many nodes: max of 200 nodes"),
		},
		{
			details: &db.DeploymentDetails{
				ID:           "123",
				Servers:      []int{4, 5, 6},
				Blockchain:   "eos",
				Nodes:        0,
				Images:       []string{},
				Params:       map[string]interface{}{},
				Resources:    []util.Resources{{Cpus: " ", Memory: " ", Volumes: []string{}, Ports: []string{}}},
				Environments: []map[string]string{},
				Files:        []map[string]string{},
				Logs:         []map[string]string{},
				Extras:       map[string]interface{}{},
			},
			expected: errors.New("must have at least 1 node"),
		},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(validateNumOfNodes(tt.details), tt.expected) {
				t.Errorf("returned error did not match expected error")
			}
		})
	}
}
