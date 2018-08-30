package db;

import(
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"fmt"
	"errors"
)

func GetAllServers() map[string]Server {
	
	db := getDB()
	defer db.Close()

	rows, err :=  db.Query(fmt.Sprintf("SELECT * FROM %s",SERVER_TABLE ))
	checkFatal(err)
	defer rows.Close()
	allServers := make(map[string]Server)
	for rows.Next() {
		var name string
		var server Server
		var switchId int
		checkFatal(rows.Scan(&server.Id,&server.Addr,&server.Iaddr.Ip,&server.Iaddr.Gateway,&server.Iaddr.Subnet,
							 &server.Nodes,&server.Max,&switch_id,&name))
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
		checkFatal(rows.Scan(&server.Id,&server.Addr,&server.Iaddr.Ip,&server.Iaddr.Gateway,&server.Iaddr.Subnet,
							 &server.Nodes,&server.Max,&switch_id,&name));
		server.switches = []Switch{ GetSwitchById(switchId) }
	}

	return server, name
}

func _InsertServer(name string,server Server,switch_id int){
	db := getDB()
	defer db.Close()

	tx,err := db.Begin()
	checkFatal(err)

	stmt,err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (addr,iaddr_ip,iaddr_gateway,iaddr_subnet,nodes,max,iface,switch,name) VALUES (?,?,?,?,?,?,?,?,?)",SERVER_TABLE))
	checkFatal(err)

	defer stmt.Close()

	_,err := stmt.Exec(server.Id,server.Addr,server.Iaddr.Ip,server.Iaddr.Gateway,server.Iaddr.Subnet,
					   server.Nodes,server.Max,switch_id,name)
	checkFatal(err)
	tx.Commit()
}