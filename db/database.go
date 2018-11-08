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

	serverSchema := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s, %s,%s,%s, %s,%s,%s, %s,%s);",
		ServerTable,
		
		"id INTEGER PRIMARY KEY AUTOINCREMENT",
		"server_id INTEGER",
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
	InsertServer("cloud1",
		Server{	
			Addr:"172.16.0.2",
			Iaddr:Iface{Ip:"172.16.0.2",Gateway:"172.16.0.1",Subnet:16},
			Nodes:0,
			Max:30,
			Id:1,
			ServerID:1,
			Iface:"ens4",
			Ips:[]string{},
			Switches:[]Switch{ Switch{Addr:"0.0.0.0",Iface:"dummy0",Brand:util.Vyos} }})

	InsertServer("cloud2",
		Server{	
			Addr:"172.16.0.3",
			Iaddr:Iface{Ip:"172.16.0.3",Gateway:"172.16.0.1",Subnet:16},
			Nodes:0,
			Max:30,
			Id:2,
			ServerID:2,
			Iface:"ens4",
			Ips:[]string{},
			Switches:[]Switch{ Switch{Addr:"0.0.0.0",Iface:"dummy0",Brand:util.Vyos} }})

	InsertServer("charlie",
		Server{
			Addr:"172.16.3.5",
			Iaddr:Iface{Ip:"10.254.3.100",Gateway:"10.254.3.1",Subnet:24},
			Nodes:0,
			Max:30,
			Id:3,
			ServerID:3,
			Iface:"eno3",
			Ips:[]string{},
			Switches:[]Switch{ Switch{Addr:"172.16.1.1",Iface:"eth3",Brand:util.Vyos} }})

	InsertServer("foxtrot",
		Server{
			Addr:"172.16.6.5",
			Iaddr:Iface{Ip:"10.254.6.100",Gateway:"10.254.6.1",Subnet:24},
			Nodes:0,
			Max:30,
			Id:6,
			ServerID:6,
			Iface:"eno3",
			Ips:[]string{},
			Switches:[]Switch{ Switch{Addr:"172.16.1.1",Iface:"eth6",Brand:util.Vyos} }})
	
}
