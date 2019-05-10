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

// Package db manages persistent state and keeps track of previous and current builds.
package db

import (
	"database/sql"
	"fmt"
	"github.com/Whiteblock/genesis/util"
	_ "github.com/mattn/go-sqlite3" //needed for db
	"os"
)

//ServerTable contains name of the server table
const ServerTable string = "servers"

//NodesTable contains name of the nodes table
const NodesTable string = "nodes"

//BuildsTable contains name of the builds table
const BuildsTable string = "builds"

var dataFolder = os.Getenv("HOME") + "/.config/whiteblock/"
var dataLoc = dataFolder + ".gdata"

var conf = util.GetConfig()

var db *sql.DB

func init() {
	var err error
	db, err = getDB()
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(50)
	checkAndUpdate()
}
func getDB() (*sql.DB, error) {

	err := os.MkdirAll(dataFolder, 0777)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(dataLoc); os.IsNotExist(err) {
		err = dbInit(dataLoc)
		if err != nil {
			return nil, err
		}
	}
	d, err := sql.Open("sqlite3", dataLoc)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func dbInit(dataLoc string) error {
	_, err := os.Create(dataLoc)
	if err != nil {
		return err
	}
	db, err = getDB()
	if err != nil {
		return err
	}

	serverSchema := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s, %s,%s,%s);",
		ServerTable,

		"id INTEGER PRIMARY KEY AUTOINCREMENT",
		"server_id INTEGER",
		"addr TEXT NOT NULL",
		"nodes INTEGER DEFAULT 0",
		"max INTEGER",
		"name TEXT")

	nodesSchema := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s, %s,%s,%s, %s);",
		NodesTable,
		"id TEXT",
		"abs_num INTEGER",
		"test_net TEXT",
		"server INTEGER",
		"local_id INTEGER",
		"ip TEXT NOT NULL",
		"label TEXT")

	buildSchema := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s, %s,%s,%s, %s,%s,%s, %s,%s,%s, %s);",
		BuildsTable,
		"id INTEGER PRIMARY KEY AUTOINCREMENT",
		"testnet TEXT",
		"servers TEXT",
		"blockchain TEXT",
		"nodes INTEGER",
		"image TEXT",
		"params TEXT",
		"resources TEXT",
		"environment TEXT",
		"files TEXT",
		"logs TEXT",
		"extras TEXT",
		"kid TEXT")

	versionSchema := fmt.Sprintf("CREATE TABLE meta (%s,%s);",
		"key TEXT",
		"value TEXT",
	)

	_, err = db.Exec(serverSchema)
	if err != nil {
		return err
	}

	_, err = db.Exec(nodesSchema)
	if err != nil {
		return err
	}
	_, err = db.Exec(buildSchema)
	if err != nil {
		return err
	}
	_, err = db.Exec(versionSchema)
	if err != nil {
		return err
	}
	err = insertLocalServers()
	if err != nil {
		return err
	}
	err = setVersion(Version)
	return err
}

//insertLocalServers adds the default server(s) to the servers database, allowing immediate use of the application
//without having to register a server
func insertLocalServers() error {
	_, err := InsertServer("cloud",
		Server{
			Addr:     "127.0.0.1",
			Nodes:    0,
			Max:      conf.MaxNodes,
			SubnetID: 1,
			ID:       -1,
			Ips:      []string{}})
	return err
}
