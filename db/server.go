package db

import(
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"fmt"
	"errors"
)

type Server struct{
	Addr		string //IP to access the server
	Iaddr		Iface //Internal IP of the server for NIC attached to the vyos
	Nodes 		int
	Max 		int
	Id			int
	Iface		string
	Ips			[]string
	Switches	[]Switch
}

type Iface struct {
	Ip 			string
	Gateway 	string
	Subnet 		int
}

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
							 &server.Nodes,&server.Max,&switchId,&name))
		swtch, err := GetSwitchById(switchId)
		checkFatal(err)

		server.Switches = []Switch{ swtch }
		allServers[name] = server
	}
	return allServers
}

func GetServer(id int) (Server, string, error) {
	db := getDB()
	defer db.Close()

	row := db.QueryRow(fmt.Sprintf("SELECT * FROM %s WHERE id = %d",SERVER_TABLE,id))
	
	var server Server
	var name string
	var switchId int

	if row.Scan(&server.Id,&server.Addr,&server.Iaddr.Ip,&server.Iaddr.Gateway,&server.Iaddr.Subnet,
				&server.Nodes,&server.Max,&switchId,&name) == sql.ErrNoRows {
		return server, name, errors.New("Not Found")
	}

	swtch, err := GetSwitchById(switchId)
	checkFatal(err)

	server.Switches = []Switch{ swtch }

	return server, name, nil
}

func _insertServer(name string,server Server,switchId int) int {
	db := getDB()
	defer db.Close()

	tx,err := db.Begin()
	checkFatal(err)

	stmt,err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (addr,iaddr_ip,iaddr_gateway,iaddr_subnet,nodes,max,iface,switch,name) VALUES (?,?,?,?,?,?,?,?,?)",SERVER_TABLE))
	checkFatal(err)

	defer stmt.Close()

	res,err := stmt.Exec(server.Id,server.Addr,server.Iaddr.Ip,server.Iaddr.Gateway,server.Iaddr.Subnet,
					   server.Nodes,server.Max,switchId,name)
	checkFatal(err)
	tx.Commit()
	id, err := res.LastInsertId()
	return int(id)
}

func InsertServer(name string,server Server) int {
	db := getDB()
	defer db.Close()

	swtch,err := GetSwitchByIP(server.Switches[0].Addr)

	if err == nil {
		return _insertServer(name,server,InsertSwitch(server.Switches[0]))
	}
	return _insertServer(name,server,swtch.Id)
	

}

func DeleteServer(id int){
	db := getDB()
	defer db.Close()

	db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = %d",SERVER_TABLE,id))
}

func UpdateServer(id int,server Server){
	//Handle Updating of Switch
	swtch,err := GetSwitchByIP(server.Switches[0].Addr)

	var switchId int

	if err != nil {
		switchId = InsertSwitch(server.Switches[0])
	}else{
		switchId = swtch.Id
	}

	db := getDB()
	defer db.Close()

	tx,err := db.Begin()
	checkFatal(err)

	stmt,err := tx.Prepare(fmt.Sprintf("UPDATE %s SET addr = ?, iaddr_ip = ?, iaddr_gateway = ?, iaddr_subnet = ?, nodes = ?, max = ?, iface = ?, switch = ? WHERE id = ? ",SERVER_TABLE))
	checkFatal(err)
	defer stmt.Close()

	_,err = stmt.Exec(server.Addr,server.Iaddr.Ip,server.Iaddr.Gateway,server.Iaddr.Subnet,
					   server.Nodes,server.Max,switchId,server.Id)
	checkFatal(err)
	tx.Commit()
}

func UpdateServerNodes(id int,nodes int){
	db := getDB()
	defer db.Close()
	db.Exec(fmt.Sprintf("UPDATE %s SET nodes = %d WHERE id = %d",SERVER_TABLE,id,nodes))
}