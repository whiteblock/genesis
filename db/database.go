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

var conf = util.GetConfig()

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
	InsertServer("cloud",
		Server{
			Addr:"127.0.0.1",
			Iaddr:Iface{Ip:"192.168.122.1",Gateway:"192.168.122.250",Subnet:24},
			Nodes:0,
			Max:conf.MaxNodes,
			ServerID:1,
			Id:-1,
			Iface:"wb_bridge",
			Ips:[]string{},
			Switches:[]Switch{Switch{Addr:"192.168.122.240",Iface:"eth1",Brand:util.Vyos}}})
}
