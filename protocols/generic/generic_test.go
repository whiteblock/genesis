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

package generic

import (
	"fmt"
	"strconv"
	"testing"
)

func TestCreatingNetworkTopology(t *testing.T) {
	var tests = []struct {
		currentNodeIndex int
		peerIds          map[int]string
		networkTopology  topology
		expected         string
	}{
		{
			currentNodeIndex: 0,
			peerIds:          map[int]string{0: "a", 1: "b", 2: "c"},
			networkTopology:  all,
			expected:         "bc",
		},
		{
			currentNodeIndex: 2,
			peerIds:          map[int]string{0: "a", 1: "b", 2: "c"},
			networkTopology:  all,
			expected:         "ab",
		},
		{
			currentNodeIndex: 1,
			peerIds:          map[int]string{0: "a", 1: "b", 2: "c"},
			networkTopology:  all,
			expected:         "ac",
		},
		{
			currentNodeIndex: 0,
			peerIds:          map[int]string{0: "a", 1: "b", 2: "c"},
			networkTopology:  sequence,
			expected:         "b",
		},
		{
			currentNodeIndex: 2,
			peerIds:          map[int]string{0: "a", 1: "b", 2: "c"},
			networkTopology:  sequence,
			expected:         "",
		},
		{
			currentNodeIndex: 1,
			peerIds:          map[int]string{0: "a", 1: "b", 2: "c"},
			networkTopology:  sequence,
			expected:         "c",
		},
		{
			currentNodeIndex: 0,
			peerIds:          map[int]string{0: "a", 1: "b", 2: "c"},
			networkTopology:  randomTwo,
			expected:         "",
		},
		{
			currentNodeIndex: 2,
			peerIds:          map[int]string{0: "a", 1: "b", 2: "c"}, // TODO this one is peering to itself. Bad boi? I just fixed it tho.
			networkTopology:  randomTwo,
			expected:         "ba",
		},
		{
			currentNodeIndex: 1,
			peerIds:          map[int]string{0: "a", 1: "b", 2: "c"},
			networkTopology:  randomTwo,
			expected:         "",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			params, err := createPeers(tt.currentNodeIndex, tt.peerIds, tt.networkTopology)
			if err != nil {
				t.Errorf("could not create peers")
			}

			fmt.Println(params)
			fmt.Println(tt.expected)

			if params != tt.expected {
				t.Errorf("return value of createPeers does not match expected value")
			}
		})
	}
}
