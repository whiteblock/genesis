/*
	Copyright 2019 Whiteblock Inc.
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
	"github.com/stretchr/testify/assert"
	"reflect"
	"strconv"
	"testing"
)

func TestGetJSONString_Successful(t *testing.T) {
	var test = []struct {
		data     map[string]interface{}
		field    string
		out      string
		expected string
	}{
		{data: map[string]interface{}{"field": "this is a test string"}, field: "field", out: "doesn't matter", expected: "this is a test string"},
		{data: map[string]interface{}{"test": "this is another test"}, field: "test", out: "doesn't matter", expected: "this is another test"},
		{data: map[string]interface{}{"test": "to be extracted"}, field: "test", out: "doesn't matter", expected: "to be extracted"},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := GetJSONString(tt.data, tt.field, &tt.out)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, tt.out)
		})
	}
}

func TestGetJSONString_Unsuccessful(t *testing.T) {
	var test = []struct {
		data  map[string]interface{}
		field string
		out   string
	}{
		{data: map[string]interface{}{"field": "this is a test string"}, field: "string"},
		{data: map[string]interface{}{}, field: "string"},
		{data: map[string]interface{}{"test": 40}, field: "test"},
		{data: map[string]interface{}{"test": nil}, field: "test"},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := GetJSONString(tt.data, tt.field, &tt.out)
			assert.Error(t, err)
		})
	}
}

func TestConvertToStringMap(t *testing.T) {
	var test = []struct {
		data map[string]interface{}
		out  map[string]string
	}{
		{data: map[string]interface{}{"1": 1, "2": 2, "3": 3}, out: map[string]string{"1": "1", "2": "2", "3": "3"}},
		{data: map[string]interface{}{"bool": false}, out: map[string]string{"bool": "false"}},
		{data: map[string]interface{}{"float": 10.74}, out: map[string]string{"float": "10.74"}},
		{data: map[string]interface{}{}, out: map[string]string{}},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(ConvertToStringMap(tt.data), tt.out) {
				t.Errorf("return value of ConvertToStringMap does not match expected value")
			}
		})
	}
}
