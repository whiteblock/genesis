package main;

import(
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"fmt"
)

func dbInit(){
	db := getDB()
	defer db.Close()

	switchTable := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s,%s);",
		SWITCH_TABLE,
		"id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT",
		"addr TINYTEXT NOT NULL",
		"iface TINYTEXT NOT NULL",
		"brand INTEGER NOT NULL")

	serverTable := fmt.Sprintf("CREATE TABLE %s (%s,%s,%s, %s,%s,%s, %s,%s,%s, %s);",
		SERVER_TABLE,
		
		"id INTERGER NOT NULL PRIMARY KEY AUTOINCREMENT",
		"addr TINYTEXT NOT NULL",
		"iaddr_ip TINYTEXT NOT NULL",

		"iaddr_gateway TINYTEXT NOT NULL",
		"iaddr_subnet INTERGER NOT NULL",
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

}


func getDB() *DB {
	if _, err := os.Stat(DATA_LOC); os.IsNotExist(err) {
	  	dbInit()
	}
	db, err := sql.Open("sqlite3", DATA_LOC)
	checkFatal(err)
	return db;
}


	addr 		string //IP to access the server
	iaddr		Iface //Internal IP of the server for NIC attached to the vyos
	nodes 		int
	max 		int
	id 			int
	iface		string
	ips 		[]string
	switches 	[]Switch

func GetAllServers() map[string]Server {
	
	db := getDB()
	defer db.Close()

	rows, err :=  db.Query(fmt.Sprintf("SELECT * FROM %s",SERVER_TABLE ))
	checkFatal(err)
	defer rows.Close()
	allServers := make(map[string]Server)
	for rows.Next() {
		var name string;
		var server Server;
		var switchId int;
		checkFatal(rows.Scan(&server.id,&server.addr,&server.iaddr.ip,&server.iaddr.gateway,&server.iaddr.subnet,
							 &server.nodes,&server.max,&switch_id,&name));
		server.switches = []Switch{ GetSwitchById(switchId) }
		allServers[name] = server
	}
	return allServers;
}

func GetServer(id int) (Server,string) {
	db := getDB()
	defer db.Close()

	rows, err :=  db.Query(fmt.Sprintf("SELECT * FROM %s WHERE id = %d",SERVER_TABLE,id))
	checkFatal(err)
	defer rows.Close()
	
	var server Server
	var name string

	for rows.Next() { //This needs to be improved
		var switchId int;
		checkFatal(rows.Scan(&server.id,&server.addr,&server.iaddr.ip,&server.iaddr.gateway,&server.iaddr.subnet,
							 &server.nodes,&server.max,&switch_id,&name));
		server.switches = []Switch{ GetSwitchById(switchId) }
	}

	return server, name
}

func InsertServer(name string,server Server){
	db := getDB()
	defer db.Close()

	tx,err := db.Begin();
	checkFatal(err);

	stmt,err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (addr,iaddr_ip,iaddr_gateway,iaddr_subnet,nodes,max,iface,switch,name)",SERVER_TABLE))
	checkFatal(err);

	defer stmt.Close()

	_,err := stmt.Exec(server.id,server.addr,server.iaddr.ip,server.iaddr.gateway,server.iaddr.subnet,
					   server.nodes,server.max,switch_id,name)
	checkFatal(err)
	tx.Commit()
}

func GetAllNodesByServer(){

}

func GetAllNodesByTest(){

}

func GetAllSwitches() Switch[] {

}

func GetSwitchById(id int){
	db := getDB()
	defer db.Close()

	rows, err :=  db.Query(fmt.Sprintf("SELECT addr,iface,brand FROM %s WHERE id = %d",SWITCH_TABLE,id))
	checkFatal(err)
	defer rows.Close()
	
	var swtch Switch

	for rows.Next() { //This needs to be improved
		checkFatal(rows.Scan(&swtch.addr,&swtch.iface,&swtch.brand);
	}

	return swtch
}