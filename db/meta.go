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

package db

import (
	"encoding/json"
	"fmt"
	"github.com/whiteblock/genesis/util"
)

//SetMeta stores a key value pair in the sql-lite database as json
func SetMeta(key string, value interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return util.LogError(err)
	}

	stmt, err := tx.Prepare("INSERT INTO meta (key,value) VALUES (?,?)")

	if err != nil {
		return util.LogError(err)
	}

	defer stmt.Close()

	v, err := json.Marshal(value)
	if err != nil {
		return util.LogError(err)
	}

	_, err = stmt.Exec(key, string(v))
	if err != nil {
		return util.LogError(err)
	}
	return util.LogError(tx.Commit())
}

//GetMeta returns the value stored at key as interface
func GetMeta(key string) (interface{}, error) {
	row := db.QueryRow(fmt.Sprintf("SELECT value FROM meta WHERE key = \"%s\"", key))
	var data []byte
	err := row.Scan(&data)
	if err != nil {
		return nil, util.LogError(err)
	}
	var out interface{}
	return out, util.LogError(json.Unmarshal(data, &out))
}

//GetMetaP fetches the value of key and returns it to v, v should be a pointer
func GetMetaP(key string, v interface{}) error {
	row := db.QueryRow(fmt.Sprintf("SELECT value FROM meta WHERE key = \"%s\"", key))
	var data []byte
	err := row.Scan(&data)
	if err != nil {
		return util.LogError(err)
	}
	return util.LogError(json.Unmarshal(data, &v))
}

//DeleteMeta deletes the value stored at key
func DeleteMeta(key string) error {
	_, err := db.Exec(fmt.Sprintf("DELETE FROM meta WHERE key = \"%s\"", key))
	return err
}
