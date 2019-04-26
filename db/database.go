/*
Package db manages persistent state and keeps track of previous and current builds.
*/
package db

import (
	"../util"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3" //needed for db
	"os"
)

var dataLoc = os.Getenv("HOME") + "/.config/whiteblock/.gdata"

//ServerTable contains name of the server table
const ServerTable string = "servers"

//NodesTable contains name of the nodes table
const NodesTable string = "nodes"

//BuildsTable contains name of the builds table
const BuildsTable string = "builds"

var conf = util.GetConfig()

var db *sql.DB

func init() {
	db = getDB()
	db.SetMaxOpenConns(50)
	checkAndUpdate()
}
func getDB() *sql.DB {
	if _, err := os.Stat(dataLoc); os.IsNotExist(err) {
		dbInit()
	}
	d, err := sql.Open("sqlite3", dataLoc)
	if err != nil {
		panic(err)
	}
	return d
}

func dbInit() {
	_, err := os.Create(dataLoc)
	if err != nil {
		panic(err)
	}
	db = getDB()

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

	buildSchema := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s, %s,%s,%s, %s,%s,%s, %s,%s,%s);",
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

		"logs TEXT",
		"extras TEXT",
		"kid TEXT")

	versionSchema := fmt.Sprintf("CREATE TABLE meta (%s,%s);",
		"key TEXT",
		"value TEXT",
	)

	_, err = db.Exec(serverSchema)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(nodesSchema)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(buildSchema)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(versionSchema)
	if err != nil {
		panic(err)
	}
	insertLocalServers(db)
	setVersion(Version)
}

//insertLocalServers adds the default server(s) to the servers database, allowing immediate use of the application
//without having to register a server
func insertLocalServers(db *sql.DB) {
	InsertServer("cloud",
		Server{
			Addr:     "127.0.0.1",
			Nodes:    0,
			Max:      conf.MaxNodes,
			SubnetID: 1,
			Id:       -1,
			Ips:      []string{}})
}
