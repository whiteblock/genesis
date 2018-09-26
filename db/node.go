package db

import(
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"fmt"
	"errors"
	util "../util"
)

//TODO: Fix broken naming convention
type Node struct {	
	Id			int 	`json:"id"`
	TestNetId	int		`json:"testNetId"`
	Server		int		`json:"server"`
	LocalId		int		`json:"localid"`
	Ip			string	`json:"ip"`
}


func GetAllNodesByServer(serverId int) []Node {
	db := getDB()
	defer db.Close()

	rows, err :=  db.Query(fmt.Sprintf("SELECT id,test_net,server,local_id,ip FROM %s WHERE server = %d",NODES_TABLE ))
	util.CheckFatal(err)
	defer rows.Close()
	
	nodes := []Node{}
	for rows.Next() {
		var node Node
		util.CheckFatal(rows.Scan(&node.Id,&node.TestNetId,&node.Server,&node.LocalId,&node.Ip))
		nodes = append(nodes,node)
	}
	return nodes
}

func GetAllNodesByTestNet(testId int)[]Node{
	db := getDB()
	defer db.Close()

	rows, err :=  db.Query(fmt.Sprintf("SELECT id,test_net,server,local_id,ip FROM %s WHERE test_net = %d",NODES_TABLE ))
	util.CheckFatal(err)
	defer rows.Close()

	nodes := []Node{}
	for rows.Next() {
		var node Node
		util.CheckFatal(rows.Scan(&node.Id,&node.TestNetId,&node.Server,&node.LocalId,&node.Ip))
		nodes = append(nodes,node)
	}
	return nodes
}

func GetAllNodes() []Node {
	
	db := getDB()
	defer db.Close()

	rows, err :=  db.Query(fmt.Sprintf("SELECT id,test_net,server,local_id,ip FROM %s",NODES_TABLE ))
	util.CheckFatal(err)
	defer rows.Close()
	nodes := []Node{}

	for rows.Next() {
		var node Node
		util.CheckFatal(rows.Scan(&node.Id,&node.TestNetId,&node.Server,&node.LocalId,&node.Ip))
		nodes = append(nodes,node)
	}
	return nodes
}

func GetNode(id int) (Node,error) {
	db := getDB()
	defer db.Close()

	row :=  db.QueryRow(fmt.Sprintf("SELECT id,test_net,server,local_id,ip FROM %s WHERE id = %d",NODES_TABLE,id))

	var node Node

	if row.Scan(&node.Id,&node.TestNetId,&node.Server,&node.LocalId,&node.Ip) == sql.ErrNoRows {
		return node, errors.New("Not Found")
	}

	return node, nil
}

func InsertNode(node Node) int {
	db := getDB()
	defer db.Close()

	tx,err := db.Begin()
	util.CheckFatal(err)

	stmt,err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (test_net,server,local_id,ip) VALUES (?,?,?,?)",NODES_TABLE))
	util.CheckFatal(err)

	defer stmt.Close()

	res,err := stmt.Exec(node.TestNetId,node.Server,node.LocalId,node.Ip)
	util.CheckFatal(err)
	tx.Commit()
	id, err := res.LastInsertId()
	return int(id)
}


func DeleteNode(id int){
	db := getDB()
	defer db.Close()

	db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = %d",NODES_TABLE,id))
}

func DeleteNodesByTestNet(id int){
	db := getDB()
	defer db.Close()

	db.Exec(fmt.Sprintf("DELETE FROM %s WHERE test_net = %d",NODES_TABLE,id))
}

func DeleteNodesByServer(id int){
	db := getDB()
	defer db.Close()

	db.Exec(fmt.Sprintf("DELETE FROM %s WHERE server = %d",NODES_TABLE,id))
}
