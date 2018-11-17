package db

import(
	_ "github.com/mattn/go-sqlite3"
	"fmt"
	"regexp"
	"errors"
	util "../util"
)

type Server struct{
	Addr		string //IP to access the server
	Iaddr		Iface //Internal IP of the server for NIC attached to the vyos
	Nodes 		int
	Max 		int
	Id			int
	ServerID	int
	Iface		string
	Switches	[]Switch
	Ips			[]string
}

type Iface struct {
	Ip 			string
	Gateway 	string
	Subnet 		int
}

func (s Server) Validate() error {
	var re = regexp.MustCompile(`(?m)[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`)
	if !re.Match([]byte(s.Addr)) {
		return errors.New("Addr is invalid")
	}
	if !re.Match([]byte(s.Iaddr.Ip)) {
		return errors.New("Iaddr.Ip is invalid")
	}
	if s.Nodes < 0 {
		return errors.New("Nodes is invalid")
	}
	if s.Nodes > s.Max {
		return errors.New("Max is invalid")
	}
	if s.ServerID < 1 {
		return errors.New("ServerID is invalid")
	}
	if len(s.Switches) == 0 {
		return errors.New("Server currently must have a switch")
	}
	err := s.Switches[0].Validate()
	if err != nil {
		return err
	}
	return nil
}

func GetAllServers() map[string]Server {

	db := getDB()
	defer db.Close()

	rows, err :=  db.Query(fmt.Sprintf("SELECT id,server_id,addr, iaddr_ip,iaddr_gateway,iaddr_subnet,nodes,max,iface,switch,name FROM %s",ServerTable ))
	util.CheckFatal(err)
	defer rows.Close()
	allServers := make(map[string]Server)
	for rows.Next() {
		var name string
		var server Server
		var switchId int
		//var subnet string
		util.CheckFatal(rows.Scan(&server.Id,&server.ServerID,&server.Addr,
								  &server.Iaddr.Ip, &server.Iaddr.Gateway, &server.Iaddr.Subnet,
							 	  &server.Nodes,&server.Max,&server.Iface,&switchId,&name))
		//fmt.Println(subnet)
		
		swtch, err := GetSwitchById(switchId)
		util.CheckFatal(err)

		server.Switches = []Switch{ swtch }
		allServers[name] = server
	}
	return allServers
}

func GetServers(ids []int) ([]Server,error) {
	var servers []Server
	for _, id := range ids {
		server,_,err := GetServer(id)
		if err != nil {
			return servers, err
		}
		servers = append(servers,server)
	}
	return servers,nil
}

func GetServer(id int) (Server, string, error) {
	db := getDB()
	defer db.Close()
	var name string
	var server Server

	rows, err :=  db.Query(fmt.Sprintf("SELECT id,server_id,addr,iaddr_ip,iaddr_gateway,iaddr_subnet,nodes,max,iface,switch,name FROM %s WHERE id = %d",
		ServerTable,id ))
	if err != nil {
		return server,name,err
	}

	
	
	if !rows.Next() {
		return server, name, errors.New("Not found")
	}
	defer rows.Close()
	var switchId int
	//var subnet string
	util.CheckFatal(rows.Scan(&server.Id,&server.ServerID,&server.Addr,
							  &server.Iaddr.Ip, &server.Iaddr.Gateway, &server.Iaddr.Subnet,
						 	  &server.Nodes,&server.Max,&server.Iface,&switchId,&name))
	//fmt.Println(subnet)
	//fmt.Printf("Switch id is %d\n",switchId)
	swtch, err := GetSwitchById(switchId)
	if err != nil {
		return server,name,err
	}

	server.Switches = []Switch{ swtch }
	
	return server,name,nil
}

func _insertServer(name string,server Server,switchId int) int {
	db := getDB()
	defer db.Close()

	tx,err := db.Begin()
	util.CheckFatal(err)

	stmt,err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (addr,server_id,iaddr_ip,iaddr_gateway,iaddr_subnet,nodes,max,iface,switch,name) VALUES (?,?,?,?,?,?,?,?,?,?)",ServerTable))
	util.CheckFatal(err)

	defer stmt.Close()

	res,err := stmt.Exec(server.Addr,server.ServerID,server.Iaddr.Ip, server.Iaddr.Gateway, server.Iaddr.Subnet,
					   server.Nodes,server.Max,server.Iface,switchId,name)
	util.CheckFatal(err)
	tx.Commit()
	id, err := res.LastInsertId()
	return int(id)
}

func InsertServer(name string,server Server) int {
	
	return _insertServer(name,server,InsertSwitch(server.Switches[0]))
}

func DeleteServer(id int){
	db := getDB()
	defer db.Close()

	db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = %d",ServerTable,id))
}

func UpdateServer(id int,server Server){
	//Handle Updating of Switch
	//fmt.Printf("UPDATED server is %+v\n",server)
	swtch,err := GetSwitchById(server.Switches[0].Id)

	var switchId int

	if err != nil {
		switchId = InsertSwitch(server.Switches[0])
	}else{
		switchId = swtch.Id
	}

	db := getDB()
	defer db.Close()

	tx,err := db.Begin()
	util.CheckFatal(err)

	stmt,err := tx.Prepare(fmt.Sprintf("UPDATE %s SET server_id = ?,addr = ?, iaddr_ip = ?, iaddr_gateway = ?, iaddr_subnet = ?, nodes = ?, max = ?, iface = ?, switch = ? WHERE id = ? ",ServerTable))
	util.CheckFatal(err)
	defer stmt.Close()

	_,err = stmt.Exec(server.ServerID,
					  server.Addr,
					  server.Iaddr.Ip,
					  server.Iaddr.Gateway,
					  server.Iaddr.Subnet,
					  server.Nodes,
					  server.Max,
					  server.Iface,
					  switchId,
					  server.Id)
	util.CheckFatal(err)
	tx.Commit()
}

func UpdateServerNodes(id int,nodes int){
	db := getDB()
	defer db.Close()
	db.Exec(fmt.Sprintf("UPDATE %s SET nodes = %d WHERE id = %d",ServerTable,id,nodes))
}

func GetHostIPsByTestNet(id int) ([]string,error) {
	db := getDB()
	defer db.Close()

	rows, err :=  db.Query(fmt.Sprintf("SELECT id,server_id,addr,iaddr_ip,iaddr_gateway,iaddr_subnet,nodes,max,iface,switch,name FROM %s INNER JOIN %s ON %s.id == %s.server WHERE %s.id == %d GROUP BY %s.id",
		ServerTable,
		NodesTable,
		ServerTable,
		NodesTable,
		ServerTable,
		id,
		ServerTable))

	ips := []string{}

	if err != nil {
		return ips, err
	}

	defer rows.Close()

	for rows.Next() {
		var ip string
		err = rows.Scan(&ip)
		if err != nil {
			return ips, err
		}

		ips = append(ips,ip)
	}
	return ips, nil
}

func GetServersByTestNet(id int) ([]Server,error) {
	db := getDB()
	defer db.Close()

	rows, err :=  db.Query(fmt.Sprintf("SELECT ip FROM %s INNER JOIN %s ON %s.id == %s.server WHERE %s.id == %d GROUP BY %s.id",
		ServerTable,
		NodesTable,
		ServerTable,
		NodesTable,
		ServerTable,
		id,
		ServerTable))
	
	servers := []Server{}

	if err != nil {
		return servers, err
	}

	defer rows.Close()

	for rows.Next() {
		var name string
		var server Server
		var switchId int

		err = rows.Scan(&server.Id,&server.Addr,&server.Iaddr.Ip,&server.Iaddr.Gateway,&server.Iaddr.Subnet,
							 &server.Nodes,&server.Max,&switchId,&name)
		if err != nil {
			return servers, err
		}
		swtch, err := GetSwitchById(switchId)
		
		if err != nil {
			return servers, err
		}

		server.Switches = []Switch{ swtch }

		servers = append(servers,server)
	}
	return servers, nil
}