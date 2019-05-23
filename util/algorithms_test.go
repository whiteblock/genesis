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
	"strconv"
	"testing"
)

func TestGetUniqueStrings(t *testing.T) {
	var test = []struct {
		in       []string
		expected []string
	}{
		{[]string{"0", "4", "4", "7", "9", "3", "8", "0"}, []string{"0", "4", "7", "9", "3", "8"}},
		{[]string{"3", "3", "2"}, []string{"3", "2"}},
		{[]string{"1", "1", "1"}, []string{"1"}},
		{[]string{"get", "test", "go", "test"}, []string{"get", "test", "go"}},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(GetUniqueStrings(tt.in), tt.expected) {
				t.Errorf("return value from GetUniqueStrings does not match expected value")
			}
		})
	}
}
