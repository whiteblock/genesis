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

func TestResources_GetMemory_Successful(t *testing.T) {
	var tests = []struct {
		res      Resources
		expected int64
	}{
		{res: Resources{
			Cpus:    "",
			Memory:  "45",
			Volumes: []string{},
			Ports:   []string{},
		}, expected: int64(45)},
		{res: Resources{
			Cpus:    "",
			Memory:  "1",
			Volumes: []string{},
			Ports:   []string{},
		}, expected: int64(1)},
		{res: Resources{
			Cpus:    "",
			Memory:  "92233720368547",
			Volumes: []string{},
			Ports:   []string{},
		}, expected: int64(92233720368547)},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			num, _ := tt.res.GetMemory()

			if num != tt.expected {
				t.Error("return value of GetMemory does not match expected value")
			}
		})
	}
}

func TestResources_GetMemory_Unsuccessful(t *testing.T) {
	var tests = []struct {
		res Resources
	}{
		{res: Resources{
			Cpus:    "",
			Memory:  "45.46",
			Volumes: []string{},
			Ports:   []string{},
		}},
		{res: Resources{
			Cpus:    "",
			Memory:  "35273409857203948572039458720349857",
			Volumes: []string{},
			Ports:   []string{},
		}},
		{res: Resources{
			Cpus:    "",
			Memory:  "s",
			Volumes: []string{},
			Ports:   []string{},
		}},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := tt.res.GetMemory()

			if err == nil {
				t.Error("return error of GetMemory does not match expected error")
			}
		})
	}
}

func TestResources_Validate_Successful(t *testing.T) {
	var tests = []struct {
		res Resources
	}{
		{res: Resources{
			Cpus:    "",
			Memory:  "",
			Volumes: []string{},
			Ports:   []string{},
		}},
		{res: Resources{
			Cpus:    "2",
			Memory:  "39458720349857",
			Volumes: []string{},
			Ports:   []string{},
		}},
		{res: Resources{
			Cpus:    "7",
			Memory:  "45",
			Volumes: []string{},
			Ports:   []string{},
		}},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if tt.res.Validate() != nil {
				t.Error("return value of Validate does not match expected value")
			}
		})
	}
}

func TestResources_Validate_Unsuccessful(t *testing.T) {
	var tests = []struct {
		res Resources
	}{
		{res: Resources{
			Cpus:    "f",
			Memory:  "sasdfa;",
			Volumes: []string{},
			Ports:   []string{},
		}},
		{res: Resources{
			Cpus:    "2984572304958234",
			Memory:  "-234875923485.8",
			Volumes: []string{},
			Ports:   []string{},
		}},
		{res: Resources{
			Cpus:    "fifteen",
			Memory:  "45",
			Volumes: []string{},
			Ports:   []string{},
		}},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if tt.res.Validate() == nil {
				t.Error("return value of Validate does not match expected value")
			}
		})
	}
}

func TestResources_ValidateAndSetDefaults_Successful(t *testing.T) {
	var tests = []struct {
		res Resources
	}{
		{res: Resources{
			Cpus:    "4",
			Memory:  "4",
			Volumes: []string{},
			Ports:   []string{},
		}},
		{res: Resources{
			Cpus:    "6",
			Memory:  "49824",
			Volumes: []string{},
			Ports:   []string{},
		}},
		{res: Resources{
			Cpus:    "",
			Memory:  "",
			Volumes: []string{},
			Ports:   []string{},
		}},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if tt.res.ValidateAndSetDefaults() != nil {
				t.Error("return value does not match expected value")
			}
		})
	}
}

func TestResources_GetParsedPortMappings(t *testing.T) {
	var tests = []struct {
		res Resources
		out map[string]string
	}{
		{res: Resources{
			Cpus:    "",
			Memory:  "",
			Volumes: []string{},
			Ports:   []string{"192.168.123.132:3333 192.168.123.132:7937293847 3333:4000 192.168.123.132:300"},
		}, out: map[string]string{"300": "192.168.123.132", "3333": "192.168.123.132", "4000": "3333", "7937293847": "192.168.123.132"}},
		{res: Resources{
			Cpus:    "",
			Memory:  "",
			Volumes: []string{},
			Ports:   []string{"333:333"},
		}, out: map[string]string{"333": "333"}},
		{res: Resources{
			Cpus:    "",
			Memory:  "",
			Volumes: []string{},
			Ports:   []string{"blahblah:3000 4000:7635 6000", "3333"},
		}, out: map[string]string{"3000": "blahblah", "7635": "4000"}},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(tt.res.GetParsedPortMappings(0), tt.out) {
				t.Error("return value of GetParsedPortMappings does not match expected value")
			}
		})
	}
}
