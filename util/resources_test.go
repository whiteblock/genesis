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
	"fmt"
	"strconv"
	"testing"
)

func Test_memconv(t *testing.T) {
	var test = []struct {
		mem string
		expected int64
		hasError bool
	}{
		{mem: "12MB", expected: 12000000, hasError: false},
		{mem: "300KB", expected: 300000, hasError: false},
		{mem: "4GB", expected: 4000000000, hasError: false},
		{mem: "5TB", expected: 5000000000000, hasError: false},
		{mem: "2", expected: 2, hasError: false},
		{mem: "", expected: 0, hasError: true},
		{mem: "3000", expected: -3000, hasError: false},
		{mem: "^MB", expected: -1, hasError: true},
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
		{res: Resources{Cpus: "4", Memory: "5gb", Volumes: []string{"1", "2", "3"}, Ports: []string{"1", "2", "3"}}},
		{res: Resources{Cpus: "10", Memory: "6kb", Volumes: []string{"4", "5", "6"}, Ports: []string{"3", "4", "5"}}},
		{res: Resources{Cpus: "5", Memory: "15mb", Volumes: []string{"5", "6", "7"}, Ports: []string{"4", "5", "6"}}},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if tt.res.Validate() != nil {
				fmt.Println(tt.res.Validate())
				t.Errorf("Validate returns error when nil is expected")
			}
		})
	}
}

func TestValidateAndSetDefaults(t *testing.T) {
	var test = []struct {
		res Resources
	}{
		{res: Resources{Cpus: "4", Memory: "", Volumes: []string{"1", "2", "3"}, Ports: []string{"1", "2", "3"}}},
		{res: Resources{Cpus: "", Memory: "6kb", Volumes: []string{"4", "5", "6"}, Ports: []string{"3", "4", "5"}}},
		{res: Resources{Cpus: "5", Memory: "15mb", Volumes: []string{"5", "6", "7"}, Ports: []string{"4", "5", "6"}}},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if tt.res.ValidateAndSetDefaults() != nil {
				fmt.Println(tt.res.ValidateAndSetDefaults())
				t.Errorf("ValidateAndSetDefaults returns error when nil is expected")
			}
		})
	}
}

