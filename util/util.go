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
	"github.com/whiteblock/go.uuid"
)

//GetUUIDString generates a new UUID
func GetUUIDString() (string, error) {
	uid, err := uuid.NewV4()
	return uid.String(), err
}

// GetJSONString checks and extracts a string from data[field].
// Will return an error if data[field] does not exist or is of the wrong type.
func GetJSONString(data map[string]interface{}, field string, out *string) error {
	rawValue, exists := data[field]
	if exists && rawValue != nil {
		switch rawValue.(type) {
		case string:
			value := rawValue.(string)
			*out = value
			return nil
		default:
			return fmt.Errorf("incorrect type for %s", field)
		}
	} else {
		return fmt.Errorf("missing field \"%s\"", field)
	}
	return nil
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

// CopyMap performs a deep copy of the given map m.
func CopyMap(m map[string]interface{}) (map[string]interface{}, error) {
	var out map[string]interface{}
	tmp, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return out, json.Unmarshal(tmp, &out)
}
