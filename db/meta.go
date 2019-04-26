package db

import (
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3" //Include sqlite as the db
	"log"
)

//SetMeta stores a key value pair in the sql-lite database as json
func SetMeta(key string, value interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return err
	}

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO meta (key,value) VALUES (?,?)"))

	if err != nil {
		log.Println(err)
		return err
	}

	defer stmt.Close()

	v, err := json.Marshal(value)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = stmt.Exec(key, string(v))
	if err != nil {
		log.Println(err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

//GetMeta returns the value stored at key as interface
func GetMeta(key string) (interface{}, error) {
	row := db.QueryRow(fmt.Sprintf("SELECT value FROM meta WHERE key = \"%s\"", key))
	var data []byte
	err := row.Scan(&data)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var out interface{}
	err = json.Unmarshal(data, &out)
	return out, err
}

//GetMetaP fetches the value of key and returns it to v, v should be a pointer
func GetMetaP(key string, v interface{}) error {
	row := db.QueryRow(fmt.Sprintf("SELECT value FROM meta WHERE key = \"%s\"", key))
	var data []byte
	err := row.Scan(&data)
	if err != nil {
		log.Println(err)
		return err
	}
	return json.Unmarshal(data, &v)
}

//DeleteMeta deletes the value stored at key
func DeleteMeta(key string) error {
	_, err := db.Exec(fmt.Sprintf("DELETE FROM meta WHERE key = \"%s\"", key))
	return err
}
