package db

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"regexp"
)

type Server struct {
	/*
	   Address of the server which is accessible from genesis
	*/
	Addr     	string   	`json:"addr"`
	Nodes    	int      	`json:"nodes"`
	Max      	int      	`json:"max"`
	Id       	int      	`json:"id"`
	SubnetID 	int      	`json:"subnetID"`
	Ips 		[]string    `json:"ips"`
}

/*
   Ensure that a server object contains valid data
*/
func (s Server) Validate() error {
	var re = regexp.MustCompile(`(?m)[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`)
	if !re.Match([]byte(s.Addr)) {
		return fmt.Errorf("Addr is invalid")
	}
	if s.Nodes < 0 {
		return fmt.Errorf("Nodes is invalid")
	}
	if s.Nodes > s.Max {
		return fmt.Errorf("Max is invalid")
	}
	if s.SubnetID < 1 {
		return fmt.Errorf("SubnetID is invalid")
	}
	return nil
}

/*
   Get all of the servers, indexed by name
*/
func GetAllServers() (map[string]Server, error) {

	rows, err := db.Query(fmt.Sprintf("SELECT id,server_id,addr,nodes,max,name FROM %s", ServerTable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	allServers := make(map[string]Server)
	for rows.Next() {
		var name string
		var server Server
		err := rows.Scan(&server.Id, &server.SubnetID, &server.Addr,
			&server.Nodes, &server.Max, &name)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		allServers[name] = server
	}
	return allServers, nil
}

/*
   Get servers from their ids
*/
func GetServers(ids []int) ([]Server, error) {
	var servers []Server
	for _, id := range ids {
		server, _, err := GetServer(id)
		if err != nil {
			log.Println(err)
			return servers, err
		}
		servers = append(servers, server)
	}
	return servers, nil
}

/*
   Get a server by id
*/
func GetServer(id int) (Server, string, error) {
	var name string
	var server Server

	rows, err := db.Query(fmt.Sprintf("SELECT id,server_id,addr,nodes,max,name FROM %s WHERE id = %d",
		ServerTable, id))
	if err != nil {
		log.Println(err)
		return server, name, err
	}

	if !rows.Next() {
		return server, name, fmt.Errorf("Not found")
	}
	defer rows.Close()
	err = rows.Scan(&server.Id, &server.SubnetID, &server.Addr,
		&server.Nodes, &server.Max, &name)
	if err != nil {
		log.Println(err)
		return server, name, err
	}

	return server, name, nil
}

/*
   Insert a new server into the database
*/
func InsertServer(name string, server Server) (int, error) {

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return -1, err
	}

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (addr,server_id,nodes,max,name) VALUES (?,?,?,?,?)", ServerTable))
	if err != nil {
		log.Println(err)
		return -1, err
	}

	defer stmt.Close()

	res, err := stmt.Exec(server.Addr, server.SubnetID,
		server.Nodes, server.Max, name)
	if err != nil {
		log.Println(err)
		return -1, err
	}
	tx.Commit()
	id, err := res.LastInsertId()
	return int(id), err
}

/*
   Delete a server by id
*/
func DeleteServer(id int) error {

	_, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = %d", ServerTable, id))
	return err
}

/*
   Update a server by id
*/
func UpdateServer(id int, server Server) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(fmt.Sprintf("UPDATE %s SET server_id = ?,addr = ?, nodes = ?, max = ? WHERE id = ? ", ServerTable))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(server.SubnetID,
		server.Addr,
		server.Nodes,
		server.Max,
		server.Id)
	if err != nil {
		return err
	}
	return tx.Commit()
}

/*
   Update the number of nodes a server has
*/
func UpdateServerNodes(id int, nodes int) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(fmt.Sprintf("UPDATE %s SET nodes = ? WHERE id = ?", ServerTable))

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(nodes, id)
	if err != nil {
		return err
	}
	return tx.Commit()

}

/*
   Get the ips of the hosts for a testnet
*/
func GetHostIPsByTestNet(id int) ([]string, error) {

	rows, err := db.Query(fmt.Sprintf("SELECT addr FROM %s INNER JOIN %s ON %s.id == %s.server WHERE %s.id == %d GROUP BY %s.id",
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

		ips = append(ips, ip)
	}
	return ips, nil
}
