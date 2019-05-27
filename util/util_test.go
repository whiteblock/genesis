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
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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
			t.Errorf("function Rm does not successfully remove given directories or files")
		}
	}
}
