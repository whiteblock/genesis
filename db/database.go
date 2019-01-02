package db;

import(
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"fmt"
	"os"
	util "../util"
)

var 	dataLoc			string		= 	os.Getenv("HOME")+"/.config/whiteblock/.gdata"
const 	SwitchTable		string		= 	"switches"
const 	ServerTable		string		= 	"servers"
const	TestTable		string		= 	"testnets"
const	NodesTable		string		= 	"nodes"

var db *sql.DB

func init(){
	db = getDB()
	db.SetMaxOpenConns(50)
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

	nodesSchema := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s, %s,%s,%s);",
		NodesTable,
		"id INTEGER PRIMARY KEY AUTOINCREMENT",
		"test_net INTERGER",
		"server INTERGER",
		"local_id INTERGER",
		"ip TEXT NOT NULL",
		"label TEXT")

	

	_,err = db.Exec(switchSchema)
	if err != nil {
		panic(err)
	}
	_,err = db.Exec(serverSchema)
	if err != nil {
		panic(err)
	}
	_,err = db.Exec(testSchema)
	if err != nil {
		panic(err)
	}
	_,err = db.Exec(nodesSchema)
	if err != nil {
		panic(err)
	}

	InsertLocalServers(db);

}


func InsertLocalServers(db *sql.DB) {
	InsertServer("bravo",
		Server{
			Addr:"172.16.2.5",
			Iaddr:Iface{Ip:"10.254.2.100",Gateway:"10.254.2.1",Subnet:24},
			Nodes:0,
			Max:30,
			Id:2,
			ServerID:2,
			Iface:"eno4",
			Ips:[]string{},
			Switches:[]Switch{ Switch{Addr:"172.16.1.1",Iface:"eth2",Brand:util.Vyos} }})

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
	InsertServer("cloud",
		Server{
			Addr:"127.0.0.1",
			Iaddr:Iface{Ip:"10.0.0.2",Gateway:"10.0.0.1",Subnet:8},
			Nodes:0,
			Max:200,
			ServerID:1,
			Id:-1,
			Iface:"wb_bridge",
			Ips:[]string{},
			Switches:[]Switch{Switch{Addr:"192.168.122.240",Iface:"eth1",Brand:util.Vyos}}})
}
