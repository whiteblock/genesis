package db;

import(
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"fmt"
)

type Node struct {
	Id			int
	TestNet		int
	Server		int
	LocalId		int
	ip			string
}


func GetAllNodesByServer(serverId int){
	db := getDB()
	defer db.Close()

	rows, err :=  db.Query(fmt.Sprintf("SELECT * FROM %s WHERE server = %d",NODES_TABLE ))
	checkFatal(err)
	defer rows.Close()
	
	nodes := Node[]{}
	for rows.Next() {
		var node Node
		checkFatal(rows.Scan(&node.Id,&node.TestNet,&node.Server,&node.LocalId,&node.Ip))
		nodes = append(nodes,node)
	}
	return nodes;
}

func GetAllNodesByTest(testId int){
	db := getDB()
	defer db.Close()

	rows, err :=  db.Query(fmt.Sprintf("SELECT * FROM %s WHERE test_net = %d",NODES_TABLE ))
	checkFatal(err)
	defer rows.Close()

	nodes := Node[]{}
	for rows.Next() {
		var node Node
		checkFatal(rows.Scan(&node.Id,&node.TestNet,&node.Server,&node.LocalId,&node.Ip))
		nodes = append(nodes,node)
	}
	return nodes
}