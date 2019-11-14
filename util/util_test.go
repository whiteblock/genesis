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
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"testing"
)

func TestGetJSONInt64_Successful(t *testing.T) {
	var test = []struct {
		data        map[string]interface{}
		field       string
		out         int64
		expected    int64
		expectedErr string
	}{
		{data: map[string]interface{}{"field": json.Number("450")}, field: "field", out: 50, expected: 450},
		{data: map[string]interface{}{"int64": json.Number("670")}, field: "int64", out: 40, expected: 670},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			GetJSONInt64(tt.data, tt.field, &tt.out)

			if tt.out != tt.expected {
				t.Errorf("GetJSONInt64 did not extract an int64 from data[field]")
			}
		})
	}
}

func TestGetJSONInt64_Unsuccessful(t *testing.T) {
	var test = []struct {
		data        map[string]interface{}
		field       string
		out         int64
		expectedErr string
	}{
		{data: map[string]interface{}{"field": 450}, field: "field", out: 30, expectedErr: "incorrect type for field"},
		{data: map[string]interface{}{}, field: "field", out: 40, expectedErr: "nil"},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := GetJSONInt64(tt.data, tt.field, &tt.out)
			if err != nil && err.Error() != tt.expectedErr {
				t.Errorf("GetJSONInt64 did not return expected error")
			}
			if err != nil && tt.expectedErr == "nil" {
				t.Errorf("GetJSONInt64 did not return expected error")
			}
		})
	}
}

func TestGetJSONString_Successful(t *testing.T) {
	var test = []struct {
		data     map[string]interface{}
		field    string
		out      string
		expected string
	}{
		{data: map[string]interface{}{"field": "this is a test string"}, field: "field", out: "doesn't matter", expected: "this is a test string"},
		{data: map[string]interface{}{"string": "this is another test"}, field: "string", out: "doesn't matter", expected: "this is another test"},
		{data: map[string]interface{}{"test": "to be extracted"}, field: "test", out: "doesn't matter", expected: "to be extracted"},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			GetJSONString(tt.data, tt.field, &tt.out)

			if tt.out != tt.expected {
				t.Errorf("GetJSONString did not extract a string from data[field]")
			}
		})
	}
}

func TestGetJSONString_Unsuccessful(t *testing.T) {
	var test = []struct {
		data        map[string]interface{}
		field       string
		out         string
		expectedErr string
	}{
		{data: map[string]interface{}{"field": "this is a test string"}, field: "string", out: "doesn't matter", expectedErr: "nil"},
		{data: map[string]interface{}{}, field: "string", out: "doesn't matter", expectedErr: "nil"},
		{data: map[string]interface{}{"test": 40}, field: "test", out: "doesn't matter", expectedErr: "incorrect type for test"},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := GetJSONString(tt.data, tt.field, &tt.out)
			if err != nil && err.Error() != "incorrect type for test" {
				t.Errorf("GetJSONString did not return the correct error")
			}

			if err != nil && tt.expectedErr == "nil" {
				t.Errorf("GETJSONString did not return the correct error")
			}
		})
	}
}

func TestMergeStringMaps(t *testing.T) {
	var test = []struct {
		m1  map[string]interface{}
		m2  map[string]interface{}
		out map[string]interface{}
	}{
		{
			m1:  map[string]interface{}{"one": 1, "two": 2, "three": 3},
			m2:  map[string]interface{}{"four": 4, "five": 5, "six": 6},
			out: map[string]interface{}{"five": 5, "four": 4, "one": 1, "six": 6, "three": 3, "two": 2},
		},
		{
			m1:  map[string]interface{}{"1": "one", "2": "two", "3": "three"},
			m2:  map[string]interface{}{"4": "four", "5": "five", "6": "six"},
			out: map[string]interface{}{"1": "one", "2": "two", "3": "three", "4": "four", "5": "five", "6": "six"},
		},
		{
			m1:  map[string]interface{}{"test": 123},
			m2:  map[string]interface{}{"test": 456},
			out: map[string]interface{}{"test": 456},
		},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(MergeStringMaps(tt.m1, tt.m2), tt.out) {
				t.Errorf("return value of MergeStringMaps does not match expected value")
			}
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

func TestFormatError(t *testing.T) {
	var test = []struct {
		res      string
		err      error
		expected string
	}{
		{res: "testing", err: errors.New("nil"), expected: "testing\nnil"},
		{res: "this is a test string", err: errors.New("test error"), expected: "this is a test string\ntest error"},
		{res: "test", err: errors.New(""), expected: "test\n"},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(FormatError(tt.res, tt.err).Error(), tt.expected) {
				t.Errorf("return value of Format Error does not match expected value")
			}
		})
	}
}

func TestCopyMap(t *testing.T) {
	var test = []struct {
		m map[string]interface{}
	}{
		{m: map[string]interface{}{"1": 1.0, "2": 2.0}},
		{m: map[string]interface{}{"bool": false}},
		{m: map[string]interface{}{"field": "this is a test string"}},
		{m: map[string]interface{}{}},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			m, err := CopyMap(tt.m)
			if err != nil {
				t.Errorf("an error occurred within CopyMap")
			}

			if !reflect.DeepEqual(m, tt.m) {
				t.Errorf("return value of CopyMap did not match expected value")
			}
		})
	}
}
