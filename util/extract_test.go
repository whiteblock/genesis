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

package util

import (
	"reflect"
	"testing"
)

func TestExtractStringMapUnsuccessful(t *testing.T) {
	in := map[string]interface{}{
		"test":   []int{1, 2, 3},
		"key":    3,
		"node":   1,
		"string": "string",
	}

	for i, _ := range in {
		t.Run(i, func(t *testing.T) {
			out, ok := ExtractStringMap(in, i)
			if reflect.DeepEqual(out, in[i]) {
				t.Errorf("return value of ExtractStringMap does not equal expected value")
			}
			if ok == true {
				t.Errorf("return bool of ExtractStringMap does not equal expected bool")
			}
		})
	}
}

func TestExtractStringMapSuccessful(t *testing.T) {
	in := map[string]interface{}{
		"interface": map[string]interface{}{"sub": 3},
		"peers":     map[string]interface{}{"node1": 1, "node2": 2},
	}

	for i, _ := range in {
		t.Run(i, func(t *testing.T) {
			out, ok := ExtractStringMap(in, i)
			if !reflect.DeepEqual(out, in[i]) {
				t.Errorf("return value of ExtractStringMap does not equal expected value")
			}
			if ok != true {
				t.Errorf("return bool of ExtractStringMap does not equal expected bool")
			}
		})
	}
}
