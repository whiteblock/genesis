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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

func TestHTTPRequest_Successful(t *testing.T) {
	var test = []struct {
		method string
		url string
		bodyData string
	}{
		{method: "", url: "https://www.wikipedia.org/", bodyData: ""},
		{method: "", url: "https://www.google.com/", bodyData: ""},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := HTTPRequest(tt.method, tt.url, tt.bodyData)

			if err != nil {
				t.Errorf("HTTPRequest returned an error when it should return <nil>")
			}
		})
	}
}

func TestHTTPRequest_Unsuccessful(t *testing.T) {
	var test = []struct {
		method string
		url string
		bodyData string
	}{
		{method: "", url: "https://www.wikipedia./", bodyData: ""},
		{method: "", url: "", bodyData: ""},
		{method: "", url: "google.com", bodyData: "blah"},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := HTTPRequest(tt.method, tt.url, tt.bodyData)

			if err == nil {
				t.Errorf("HTTPRequest returned <nil> when it should return an error")
			}
		})
	}
}

func TestJwtHTTPRequest_Successful(t *testing.T) {
	var test = []struct {
		method string
		url string
		jwt string
		bodyData string
	}{
		{method: "", url: "https://www.wikipedia.org/", jwt: "aaaaaaaaaa.bbbbbbbbbbb.cccccccccccc", bodyData: ""},
		{method: "", url: "https://www.google.com/", jwt: "aaaaaaaaaa.bbbbbbbbbbb.cccccccccccc", bodyData: ""},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := JwtHTTPRequest(tt.method, tt.url, tt.jwt, tt.bodyData)

			if err != nil {
				t.Errorf("JwtHTTPRequest returned an error when expected error is <nil>")
			}
		})
	}
}

func TestJwtHTTPRequest_Unsuccessful(t *testing.T) {
	var test = []struct {
		method string
		url string
		jwt string
		bodyData string
	}{
		{method: "", url: "https://www.wikipedia/", jwt: "aaaaaaaaaa.bbbbbbbbbbb.cccccccccccc", bodyData: ""},
		{method: "", url: "www.google.com/", jwt: "aaaaaaaaaa.bbbbbbbbbbb.cccccccccccc", bodyData: ""},
		{method: "", url: "google.com/", jwt: "", bodyData: ""},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := JwtHTTPRequest(tt.method, tt.url, tt.jwt, tt.bodyData)

			if err == nil {
				t.Errorf("JwtHTTPRequest returned <nil> when an error was expected")
			}
		})
	}
}

// TODO make sure these tests work
func TestExtractJwt_Successful(t *testing.T) {
	var test = []struct {
		method string
		url string
		body io.Reader
	}{
		{method: "", url: "https://www.wikipedia.com/", body: nil},
		{method: "POST", url: "https://www.wikipedia.com/", body: nil},
		{method: "DELETE", url: "https://www.wikipedia.com/", body: nil},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.url, tt.body)
			req.Header = map[string][]string{"Authorization": []string{"Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ=="}}

			_, err := ExtractJwt(req)

			if err != nil {
				t.Errorf("ExtractJwt returned an error when <nil> was expected")
			}
		})
	}
}

func TestExtractJwt_Unsuccessful(t *testing.T) {
	var test = []struct {
		method string
		url string
		body io.Reader
	}{
		{method: "", url: "https://www.wikipedia.com/", body: nil},
		{method: "POST", url: "https://www.wikipedia.com/", body: nil},
		{method: "DELETE", url: "https://www.wikipedia.com/", body: nil},
		{method: "", url: "https://wikipedia.com/", body: nil},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.url, tt.body)

			if i == 3 {
				req.Header = map[string][]string{"Authorization": []string{""}}
			}

			_, err := ExtractJwt(req)

			if err == nil {
				t.Errorf("ExtractJwt returned <nil> when an error was expected")
			}
		})
	}
}

func TestRm(t *testing.T) {
	files := []string{}

	for i := 0; i <= 3; i++ {
		file, err := ioutil.TempDir("", "prefix")
		if err != nil {
			t.Errorf("error with ioutil.TemDir directory generation")
		}

		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("error with ioutil.TemDir directory generation")
		}

		filepath, _ := filepath.Abs(file)

		files = append(files, filepath)
	}

	err := Rm(files...)
	if err != nil {
		t.Errorf("Rm returned an error when it should have returned <nil>")
	}

	for _, file := range files {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Errorf("Rm does not successfully remove given directories or files")
		}
	}
}

func TestLsr(t *testing.T) {
	newFile, err := ioutil.TempDir("", "TESTTESTTEST")
	if err != nil {
		t.Errorf("ioutil.TempDir returned an error")
	}

	newFilePath, _ := filepath.Abs(newFile)

	files := []string{}

	for i := 0; i <= 3; i++ {
		file, err := ioutil.TempFile(newFilePath, strconv.Itoa(i))
		if err != nil {
			t.Errorf("ioutil.TempFile returned an error")
		}

		files = append(files, file.Name())
	}

	check, err := Lsr(newFilePath)
	if err != nil {
		t.Errorf("Lsr returned an error when <nil> was expected")
	}

	for i, file := range files {
		if check[i] != file {
			t.Errorf("return value of Lsr did not match expected value")
		}
	}
}

// TODO: I don't think this function does what it purports to.. is it deprecated?
//  There is only one usage in the code base
func TestGetPath(t *testing.T) {
	var test = []struct {
		path string
	}{
		{path: "var/test/testing/123"},
		{path: "var/test"},
		{path: "var/"},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			fmt.Println(GetPath(tt.path))
		})
	}
}

func TestGetJSONInt64_Successful(t *testing.T) {
	var test = []struct {
		data map[string]interface{}
		field string
		out int64
		expected int64
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

// TODO: Finish this one
func TestGetJSONInt64_Unsuccessful(t *testing.T) {
	var test = []struct {
		data map[string]interface{}
		field string
		out int64
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
		data map[string]interface{}
		field string
		out string
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
		data map[string]interface{}
		field string
		out string
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
	var test = []struct{
		m1		map[string]interface{}
		m2		map[string]interface{}
		out		map[string]interface{}
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
			m1: map[string]interface{}{"test": 123},
			m2: map[string]interface{}{"test": 456},
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
		out map[string]string
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
