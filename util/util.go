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

// Package util provides a multitude of support functions to
// help make development easier. Use of these functions should be preferred,
// as it allows for easier maintenance.
package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/go.uuid"
)

// HTTPRequest Sends a HTTP request and returns the body. Gives an error if the http request failed
// or returned a non success code.
func HTTPRequest(method string, url string, bodyData string) ([]byte, error) {
	log.WithFields(log.Fields{"method": method, "url": url, "body": bodyData}).Trace("sending an http request")
	body := strings.NewReader(bodyData)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, LogError(err)
	}

	req.Close = true
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, LogError(err)
	}

	defer resp.Body.Close()
	buf := new(bytes.Buffer)

	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, LogError(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(buf.String())
	}
	return []byte(buf.String()), nil
}

// JwtHTTPRequest is similar to HttpRequest, but it have the content-type set as application/json and it will
// put the given jwt in the auth header
func JwtHTTPRequest(method string, url string, jwt string, bodyData string) (string, error) {
	log.WithFields(log.Fields{"method": method, "url": url, "body": bodyData}).Trace("sending an http request with a jwt")
	body := strings.NewReader(bodyData)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return "", LogError(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	req.Close = true
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", LogError(err)
	}

	defer resp.Body.Close()
	buf := new(bytes.Buffer)

	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", LogError(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf(buf.String())
	}
	return buf.String(), nil
}

// ExtractJwt will attempt to extract and return the jwt from the auth header
func ExtractJwt(r *http.Request) (string, error) {
	tokenString := r.Header.Get("Authorization")

	if len(tokenString) == 0 {
		return "", fmt.Errorf("missing JWT in authorization header")
	}
	splt := strings.Split(tokenString, " ")
	if len(splt) < 2 {
		return "", fmt.Errorf("invalid auth header")
	}
	return splt[1], nil
}

//GetKidFromJwt will attempt to extract the kid from the given jwt
func GetKidFromJwt(jwt string) (string, error) {
	if len(jwt) == 0 {
		return "", fmt.Errorf("given empty string for JWT")
	}
	headerb64 := strings.Split(jwt, ".")[0]
	headerJSON, err := base64.StdEncoding.DecodeString(headerb64)
	if err != nil {
		return "", LogError(err)
	}
	var header map[string]interface{}
	err = json.Unmarshal(headerJSON, &header)
	if err != nil {
		return "", LogError(err)
	}
	kidAsI, ok := header["kid"]
	if !ok {
		return "", fmt.Errorf("JWT does not have kid in header")
	}
	kid, ok := kidAsI.(string)
	if !ok {
		return "", fmt.Errorf("kid is not string as expected")
	}
	return kid, nil
}

//GetUUIDString generates a new UUID
func GetUUIDString() (string, error) {
	uid, err := uuid.NewV4()
	return uid.String(), err
}

/****Basic Linux Functions****/

// Rm removes all of the given directories or files. Convenience function for os.RemoveAll
func Rm(directories ...string) error {
	for _, directory := range directories {
		log.WithFields(log.Fields{"dir": directory}).Info("removing directory")

		err := os.RemoveAll(directory)
		if err != nil {
			return LogError(err)
		}
	}
	return nil
}

// Lsr lists the contents of a directory recursively
func Lsr(_dir string) ([]string, error) {
	dir := _dir
	if dir[len(dir)-1:] != "/" {
		dir += "/"
	}
	out := []string{}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, LogError(err)
	}
	for _, f := range files {
		if f.IsDir() {
			contents, err := Lsr(fmt.Sprintf("%s%s/", dir, f.Name()))
			if err != nil {
				return nil, LogError(err)
			}
			out = append(out, contents...)
		} else {
			out = append(out, fmt.Sprintf("%s%s", dir, f.Name()))
		}
	}
	return out, nil
}

// CombineConfig combines an Array with \n as the delimiter.
// Useful for generating configuration files. DEPRECATED
func CombineConfig(entries []string) string { // TODO: delete this function if deprecated.. it has 4 usages throughout the code base
	out := ""
	for _, entry := range entries {
		out += fmt.Sprintf("%s\n", entry)
	}
	return out
}

/*
   BashExec executes _cmd in bash then return the result
*/
/*func BashExec(_cmd string) (string, error) {
	cmd := exec.Command("bash", "-c", _cmd)

	var resultsRaw bytes.Buffer

	cmd.Stdout = &resultsRaw
	err := cmd.Start()
	if err != nil {
		return "", LogError(err)
	}
	err = cmd.Wait()
	if err != nil {
		return "", LogError(err)
	}

	return resultsRaw.String(), nil
}*/

// GetPath extracts the base path of the given path
func GetPath(path string) string {
	index := strings.LastIndex(path, "/")
	if index != -1 {
		return path
	}
	return path[:index]
}

/******* JSON helper functions *******/

// GetJSONInt64 checks and extracts a int64 from data[field].
// Will return an error if data[field] does not exist or is of the wrong type.
func GetJSONInt64(data map[string]interface{}, field string, out *int64) error {
	rawValue, exists := data[field]
	if exists && rawValue != nil {
		switch rawValue.(type) {
		case json.Number:
			value, err := rawValue.(json.Number).Int64()
			if err != nil {
				return err
			}
			*out = value
			return nil
		default:
			return fmt.Errorf("incorrect type for %s", field)
		}
	}
	return nil
}

// GetJSONString checks and extracts a string from data[field].
// Will return an error if data[field] does not exist or is of the wrong type.
func GetJSONString(data map[string]interface{}, field string, out *string) error {
	rawValue, exists := data[field]
	if exists && rawValue != nil {
		switch rawValue.(type) {
		case string:
			value, valid := rawValue.(string)
			if !valid {
				return fmt.Errorf("invalid string")
			}
			*out = value
			return nil
		default:
			return fmt.Errorf("incorrect type for %s", field)
		}
	}
	return nil
}

// MergeStringMaps merges two maps of string to interface together and returns it
// If there are conflicting keys, the value in m2 will be chosen.
func MergeStringMaps(m1 map[string]interface{}, m2 map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k1, v1 := range m1 {
		out[k1] = v1
	}

	for k2, v2 := range m2 {
		out[k2] = v2
	}
	return out
}

// ConvertToStringMap converts a map of string to interface to a map of string to json
func ConvertToStringMap(data map[string]interface{}) map[string]string {
	out := make(map[string]string)

	for key, value := range data {
		strval, _ := json.Marshal(value)
		out[key] = string(strval)
	}
	return out
}

// FormatError produced a standard error for execution.
func FormatError(res string, err error) error {
	return fmt.Errorf("%s\n%s", res, err.Error())
}

// CopyMap performs a deep copy of the given map m.
func CopyMap(m map[string]interface{}) (map[string]interface{}, error) {
	var out map[string]interface{}
	tmp, err := json.Marshal(m)
	if err != nil {
		return nil, LogError(err)
	}
	return out, LogError(json.Unmarshal(tmp, &out))
}

// LogError takes in an error, logs that error and returns that error.
// Used to help reduce code clutter and unify the error handling in the code.
// Has no effect if err == nil
func LogError(err error) error {
	if err == nil {
		return err // do nothing if the given err is nil
	}
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		log.Error(err.Error())
	} else {
		log.WithFields(log.Fields{"file": file, "line": line}).Error(err.Error())
	}

	return err
}

func LogErrorf(format string, a ...interface{}) error {
	return LogError(fmt.Errorf(format, a...))
}
