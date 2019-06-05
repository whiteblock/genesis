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

package netconf

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/whiteblock/genesis/util"
)

func TestCreateLinks(t *testing.T) {
	var test = []struct {
		pnts []util.Point
		c *Calculator
		expected [][]Link
	}{
		{
			pnts: []util.Point{{X: 0, Y: 2}, {X: 3, Y: 2}},
			c: &Calculator{
				Loss: func(float64) float64 {return 3.5},
				Delay: func(float64) int {return 2},
				Rate: func(float64) string {return "0"},
				Duplication: func(float64) float64 {return 0},
				Corrupt: func(float64) float64 {return 0.2},
				Reorder: func(float64) float64 {return 0.1},
			},
			expected: [][]Link{
				{{0, 0, 3.5, 2, "0", 0, 0.2, 0.1}, {0, 1, 3.5, 2, "0", 0, 0.2, 0.1}},
				{{1, 0, 3.5, 2, "0", 0, 0.2, 0.1}, {1, 1, 3.5, 2, "0", 0, 0.2, 0.1}},
			},
		},
		{
			pnts: []util.Point{{X: 5, Y: 4}, {X: 0, Y: 0}},
			c: &Calculator{
				Loss: func(float64) float64 {return 1.5},
				Delay: func(float64) int {return 1},
				Rate: func(float64) string {return "2"},
				Duplication: func(float64) float64 {return 0.5},
				Corrupt: func(float64) float64 {return 1.5},
				Reorder: func(float64) float64 {return 0},
			},
			expected: [][]Link{
				{{0, 0, 1.5, 1, "2", 0.5, 1.5, 0}, {0, 1, 1.5, 1, "2", 0.5, 1.5, 0}},
				{{1, 0, 1.5, 1, "2", 0.5, 1.5, 0}, {1, 1, 1.5,1, "2", 0.5, 1.5, 0}},
			},
		},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(CreateLinks(tt.pnts, tt.c), tt.expected) {
				t.Errorf("return value of CreateLinks does not match expected value")
			}
		})
	}
}