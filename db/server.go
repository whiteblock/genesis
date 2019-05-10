/*
	Copyright 2019 Whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Genesis is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package db

import (
	"github.com/Whiteblock/genesis/util"
	"fmt"
	_ "github.com/mattn/go-sqlite3" //sqlite
	"regexp"
)

// Server represents a server on which genesis can build
type Server struct {
	// Addr is the address of the server which is accessible by genesis
	Addr string `json:"addr"`
	// Nodes is the number of nodes currently on this server
	Nodes int `json:"nodes"`
	// Max is the maximum number of nodes that server supports
	Max int `json:"max"`
	// ID is the ID of the server
	ID int `json:"id"`
	// SubnetID is the number used in the IP scheme for nodes on this server
	SubnetID int `json:"subnetID"`
	Ips      []string //To be removed
}

// Validate ensures that the  server object contains valid data
func (s Server) Validate() error {
	var re = regexp.MustCompile(`(?m)[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`)
	if !re.Match([]byte(s.Addr)) {
		return fmt.Errorf("invalid addr")
	}
	if s.Nodes < 0 {
		return fmt.Errorf("invalid nodes")
	}
	if s.Nodes > s.Max {
		return fmt.Errorf("invalid max")
	}
	if s.SubnetID < 1 {
		return fmt.Errorf("invalid SubnetID")
	}
	return nil
}

// GetAllServers gets all of the servers, indexed by name
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
		err := rows.Scan(&server.ID, &server.SubnetID, &server.Addr,
			&server.Nodes, &server.Max, &name)
		if err != nil {
			return nil, util.LogError(err)
		}

		allServers[name] = server
	}
	return allServers, nil
}

//GetServers gets servers from their ids
func GetServers(ids []int) ([]Server, error) {
	var servers []Server
	for _, id := range ids {
		server, _, err := GetServer(id)
		if err != nil {
			return servers, util.LogError(err)
		}
		servers = append(servers, server)
	}
	return servers, nil
}

//GetServer gets a server by its id
func GetServer(id int) (Server, string, error) {
	var name string
	var server Server

	rows, err := db.Query(fmt.Sprintf("SELECT id,server_id,addr,nodes,max,name FROM %s WHERE id = %d",
		ServerTable, id))
	if err != nil {
		return server, name, util.LogError(err)
	}

	if !rows.Next() {
		return server, name, fmt.Errorf("not found")
	}
	defer rows.Close()
	err = rows.Scan(&server.ID, &server.SubnetID, &server.Addr,
		&server.Nodes, &server.Max, &name)
	if err != nil {
		return server, name, util.LogError(err)
	}

	return server, name, nil
}

//InsertServer inserts a new server into the database
func InsertServer(name string, server Server) (int, error) {

	tx, err := db.Begin()
	if err != nil {
		return -1, util.LogError(err)
	}

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (addr,server_id,nodes,max,name) VALUES (?,?,?,?,?)", ServerTable))
	if err != nil {
		return -1, util.LogError(err)
	}

	defer stmt.Close()

	res, err := stmt.Exec(server.Addr, server.SubnetID,
		server.Nodes, server.Max, name)
	if err != nil {
		return -1, util.LogError(err)
	}
	tx.Commit()
	id, err := res.LastInsertId()
	return int(id), err
}

// DeleteServer deletes a server by id
func DeleteServer(id int) error {

	_, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = %d", ServerTable, id))
	return err
}

//UpdateServer updates a server by id
func UpdateServer(id int, server Server) error {

	tx, err := db.Begin()
	if err != nil {
		return util.LogError(err)
	}

	stmt, err := tx.Prepare(fmt.Sprintf("UPDATE %s SET server_id = ?,addr = ?, nodes = ?, max = ? WHERE id = ? ", ServerTable))
	if err != nil {
		return util.LogError(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(server.SubnetID,
		server.Addr,
		server.Nodes,
		server.Max,
		server.ID)
	if err != nil {
		return util.LogError(err)
	}
	return tx.Commit()
}

//UpdateServerNodes update the number of nodes a server has
func UpdateServerNodes(id int, nodes int) error {

	tx, err := db.Begin()
	if err != nil {
		return util.LogError(err)
	}

	stmt, err := tx.Prepare(fmt.Sprintf("UPDATE %s SET nodes = ? WHERE id = ?", ServerTable))

	if err != nil {
		return util.LogError(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(nodes, id)
	if err != nil {
		return util.LogError(err)
	}
	return tx.Commit()

}

//GetHostIPsByTestNet gets the ips of the hosts for a testnet
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
		return ips, util.LogError(err)
	}

	defer rows.Close()

	for rows.Next() {
		var ip string
		err = rows.Scan(&ip)
		if err != nil {
			return ips, util.LogError(err)
		}

		ips = append(ips, ip)
	}
	return ips, nil
}
