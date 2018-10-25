package db;

import(
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"fmt"
	"os"
	util "../util"
)

const 	dataLoc			string		= 	".gdata"
const 	SwitchTable		string		= 	"switches"
const 	ServerTable		string		= 	"servers"
const	TestTable		string		= 	"testnets"
const	NodesTable		string		= 	"nodes"

func dbInit(){
	_, err := os.Create(dataLoc)
	if err != nil {
		panic(err)
	}
	db := getDB()
	defer db.Close()

	switchSchema := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s,%s);",
		SwitchTable,
		"id INTEGER PRIMARY KEY AUTOINCREMENT",
		"addr TEXT NOT NULL",
		"iface TEXT NOT NULL",
		"brand INTEGER")

	serverSchema := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s, %s,%s,%s, %s,%s,%s, %s);",
		ServerTable,
		
		"id INTEGER PRIMARY KEY AUTOINCREMENT",
		"addr TEXT NOT NULL",
		"iaddr_ip TEXT NOT NULL",

		"iaddr_gateway TEXT NOT NULL",
		"iaddr_subnet INTEGER",
		"nodes INTEGER DEFAULT 0",

		"max INTEGER",
		"iface TEXT NOT NULL",
		"switch INTEGER",
		"name TEXT")

	testSchema := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s,%s);",
		TestTable,
		"id INTEGER PRIMARY KEY AUTOINCREMENT",
		"blockchain TEXT NOT NULL",
		"nodes INTERGER",
		"image TEXT NOT NULL")

	nodesSchema := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s, %s,%s);",
		NodesTable,

		"id INTEGER PRIMARY KEY AUTOINCREMENT",
		"test_net INTERGER",
		"server INTERGER",

		"local_id INTERGER",
		"ip TEXT NOT NULL")

	

	_,err = db.Exec(switchSchema)
	util.CheckFatal(err)
	_,err = db.Exec(serverSchema)
	util.CheckFatal(err)
	_,err = db.Exec(testSchema)
	util.CheckFatal(err)
	_,err = db.Exec(nodesSchema)
	util.CheckFatal(err)

	InsertLocalServers(db);

}


func getDB() *sql.DB {
	if _, err := os.Stat(dataLoc); os.IsNotExist(err) {
	  	dbInit()
	}
	db, err := sql.Open("sqlite3", dataLoc)
	util.CheckFatal(err)
	return db
}

func InsertLocalServers(db *sql.DB) {
	InsertServer("alpha",
		Server{	
			Addr:"172.16.0.2",
			Iaddr:Iface{Ip:"172.16.0.2",Gateway:"172.16.0.1",Subnet:16},
			Nodes:0,
			Max:30,
			Id:1,
			Iface:"ens4",
			Ips:[]string{},
			Switches:[]Switch{ Switch{Addr:"0.0.0.0",Iface:"dummy0",Brand:util.Vyos} }})

	InsertServer("bravo",
		Server{	
			Addr:"172.16.0.3",
			Iaddr:Iface{Ip:"172.16.0.3",Gateway:"172.16.0.1",Subnet:16},
			Nodes:0,
			Max:30,
			Id:2,
			Iface:"ens4",
			Ips:[]string{},
			Switches:[]Switch{ Switch{Addr:"0.0.0.0",Iface:"dummy0",Brand:util.Vyos} }})

	
}
