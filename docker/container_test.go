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

package docker

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/util"
)

func TestNewNodeContainer(t *testing.T) {
	testNode := new(db.Node)

	var tests = []struct {
		node *db.Node
		env map[string]string
		resources util.Resources
		SubnetID int
		expected ContainerDetails
	}{
		{
			node: testNode,
			env: map[string]string{},
			resources: util.Resources{},
			SubnetID: 4,
			expected: ContainerDetails{
				Environment: map[string]string{},
				Image: testNode.Image,
				Node: testNode.LocalID,
				Resources: util.Resources{},
				SubnetID: 4,
				NetworkIndex: 0,
				Type: ContainerType(0),
			},
		},
		{
			node: testNode,
			env: map[string]string{},
			resources: util.Resources{
				Cpus: "5",
				Memory: "5GB",
				Volumes: []string{},
				Ports: []string{},
			},
			SubnetID: 16,
			expected: ContainerDetails{
				Environment: map[string]string{},
				Image: testNode.Image,
				Node: testNode.LocalID,
				Resources: util.Resources{
					Cpus: "5",
					Memory: "5GB",
					Volumes: []string{},
					Ports: []string{},
				},
				SubnetID: 16,
				NetworkIndex: 0,
				Type: ContainerType(0),
			},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(NewNodeContainer(tt.node, tt.env, tt.resources, tt.SubnetID), &tt.expected) {
				t.Error("return value of New Node Container does not match expected value")
			}
		})
	}
}

func BenchmarkNewNodeContainer(b *testing.B) {
	for n := 0; n < b.N; n++ {
		NewNodeContainer(new(db.Node), map[string]string{}, util.Resources{}, 4)
	}
}
