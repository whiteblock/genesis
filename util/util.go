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
	"encoding/json"
	"fmt"
	"runtime"

	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/go.uuid"
)

//GetUUIDString generates a new UUID
func GetUUIDString() (string, error) {
	uid, err := uuid.NewV4()
	return uid.String(), err
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

// LogErrorf formats an error message, logs that error and returns that error.
// Used to help reduce code clutter and unify the error handling in the code.
// Has no effect if err == nil
func LogErrorf(format string, a ...interface{}) error {
	return LogError(fmt.Errorf(format, a...))
}
