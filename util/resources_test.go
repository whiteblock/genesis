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
	"strconv"
	"testing"
)

func Test_memconv(t *testing.T) {
	var test = []struct {
		mem string
		expected int64
		hasError bool
	}{
		{"12MB", 12000000, false},
		{"300KB", 300000, false},
		{"4GB", 4000000000, false},
		{"5TB", 5000000000000, false},
		{"2", 2, false},
		{"", 0, true},
		{"3000", 3000, false},
		{"^MB", -1, true},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			out, err := memconv(tt.mem)
			if out != tt.expected {
				t.Errorf("return value of memconv does not match expected value")
			}

			if tt.hasError && err == nil {
				t.Errorf("did not return expected error")
			}
		})
	}
}

func TestValidate(t *testing.T) {
	var test = []struct {
		res Resources
	}{
		{Resources{"4TB", "10", []string{"1", "2", "3"}, []string{"1", "2", "3"}}},
		{Resources{"10MB", "15", []string{"4", "5", "6"}, []string{"3", "4", "5"}}},
		{Resources{"5GB", "20", []string{"5", "6", "7"}, []string{"4", "5", "6"}}},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {

		})
	}
}

func
