package db;

import(
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"fmt"
	"os"
	util "../util"
)

const 	DATA_LOC		string		= 	"~/.dddata"
const 	SWITCH_TABLE	string		= 	"switches"
const 	SERVER_TABLE	string		= 	"servers"
const	TEST_TABLE		string		= 	"testnets"
const	NODES_TABLE		string		= 	"nodes"

func dbInit(){
	db := getDB()
	defer db.Close()

	switchTable := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s,%s);",
		SWITCH_TABLE,
		"id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT",
		"addr TEXT NOT NULL",
		"iface TEXT NOT NULL",
		"brand INTEGER NOT NULL")

	serverTable := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s, %s,%s,%s, %s,%s,%s, %s);",
		SERVER_TABLE,
		
		"id INTERGER NOT NULL PRIMARY KEY AUTOINCREMENT",
		"addr TEXT NOT NULL",
		"iaddr_ip TEXT NOT NULL",

		"iaddr_Gateway TEXT NOT NULL",
		"iaddr_Subnet INTERGER NOT NULL",
		"nodes INTEGER NOT NULL DEFAULT 0",

		"max INTEGER NOT NULL",
		"iface TEXT NOT NULL",
		"switch INTEGER NOT NULL",
		"name TEXT")

	testTable := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s,%s);",
		TEST_TABLE,
		"id INTERGER NOT NULL PRIMARY KEY AUTOINCREMENT",
		"blockchain TEXT NOT NULL",
		"nodes INTERGER NOT NULL",
		"image TEXT NOT NULL")

	nodesTable := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s, %s,%s);",
		NODES_TABLE,

		"id INTERGER NOT NULL PRIMARY KEY AUTOINCREMENT",
		"test_net INTERGER NOT NULL",
		"server INTERGER NOT NULL",

		"local_id INTERGER NOT NULL",
		"ip TEXT NOT NULL")

	

	db.Exec(switchTable)
	db.Exec(serverTable)
	db.Exec(testTable)
	db.Exec(nodesTable)

	InsertLocalServers(db);

}


func getDB() *sql.DB {
	if _, err := os.Stat(DATA_LOC); os.IsNotExist(err) {
	  	dbInit()
	}
	db, err := sql.Open("sqlite3", DATA_LOC)
	util.CheckFatal(err)
	return db
}

func InsertLocalServers(db *sql.DB) {
	InsertServer("alpha",
		Server{	
			Addr:"172.16.1.5",
			Iaddr:Iface{Ip:"10.254.1.100",Gateway:"10.254.1.1",Subnet:24},
			Nodes:0,
			Max:100,
			Id:1,
			Iface:"eno4",
			Ips:[]string{},
			Switches:[]Switch{ Switch{Addr:"172.16.1.1",Iface:"eth1",Brand:util.Vyos} }})

	InsertServer("bravo",
		Server{	
			Addr:"172.16.2.5",
			Iaddr:Iface{Ip:"10.254.2.100",Gateway:"10.254.2.1",Subnet:24},
			Nodes:0,
			Max:100,
			Id:2,
			Iface:"eno1",
			Ips:[]string{},
			Switches:[]Switch{ Switch{Addr:"172.16.1.1",Iface:"eth2",Brand:util.Vyos} }})

	InsertServer("charlie",
		Server{	
			Addr:"172.16.3.5",
			Iaddr:Iface{Ip:"10.254.3.100",Gateway:"10.254.3.1",Subnet:24},
			Nodes:0,
			Max:30,
			Id:3,
			Iface:"eno3",
			Ips:[]string{},
			Switches:[]Switch{ Switch{Addr:"172.16.1.1",Iface:"eth3",Brand:util.Vyos} }})

	InsertServer("delta",
		Server{	
			Addr:"172.16.4.5",
			Iaddr:Iface{Ip:"10.254.4.100",Gateway:"10.254.4.1",Subnet:24},
			Nodes:0,
			Max:32,
			Id:4,
			Iface:"eno3",
			Ips:[]string{},
			Switches:[]Switch{ Switch{Addr:"172.16.1.1",Iface:"eth4",Brand:util.Vyos} }})

	InsertServer("ns2", 
		Server{
			Addr:"172.16.8.8",
			Iaddr:Iface{Ip:"10.254.5.100",Gateway:"10.254.5.1",Subnet:24},
			Nodes:0,
			Max:10,
			Id:5,
			Iface:"eth0",
			Ips:[]string{},
			Switches:[]Switch{ Switch{Addr:"172.16.5.1",Iface:"eth0",Brand:util.Vyos} }})
}
