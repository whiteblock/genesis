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
	LocalId		int		`json:"localId"`
	Ip			string	`json:"ip"`
}


func GetAllNodesByServer(serverId int) ([]Node,error) {
	db := getDB()
	defer db.Close()

	rows, err :=  db.Query(fmt.Sprintf("SELECT id,test_net,server,local_id,ip FROM %s WHERE server = %d",NodesTable ))
	if err != nil {
		return nil,err
	}
	defer rows.Close()
	
	nodes := []Node{}
	for rows.Next() {
		var node Node
		err := rows.Scan(&node.Id,&node.TestNetId,&node.Server,&node.LocalId,&node.Ip)
		if err != nil {
			return nil,err
		}
		nodes = append(nodes,node)
	}
	return nodes,nil
}

func GetAllNodesByTestNet(testId int) ([]Node,error) {
	db := getDB()
	defer db.Close()
	nodes := []Node{}

	rows, err :=  db.Query(fmt.Sprintf("SELECT id,test_net,server,local_id,ip FROM %s WHERE test_net = %d",NodesTable,testId ))
	if err != nil {
		return nodes,err
	}
	defer rows.Close()

	
	for rows.Next() {
		var node Node
		err := rows.Scan(&node.Id,&node.TestNetId,&node.Server,&node.LocalId,&node.Ip)
		if err != nil {
			return nodes, err
		}
		nodes = append(nodes,node)
	}
	return nodes, nil
}

func GetAllNodes() ([]Node,error) {
	
	db := getDB()
	defer db.Close()

	rows, err :=  db.Query(fmt.Sprintf("SELECT id,test_net,server,local_id,ip FROM %s",NodesTable ))
	if err != nil {
		return nil,err
	}
	defer rows.Close()
	nodes := []Node{}

	for rows.Next() {
		var node Node
		err := rows.Scan(&node.Id,&node.TestNetId,&node.Server,&node.LocalId,&node.Ip)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes,node)
	}
	return nodes,nil
}

func GetNode(id int) (Node,error) {
	db := getDB()
	defer db.Close()

	row :=  db.QueryRow(fmt.Sprintf("SELECT id,test_net,server,local_id,ip FROM %s WHERE id = %d",NodesTable,id))

	var node Node

	if row.Scan(&node.Id,&node.TestNetId,&node.Server,&node.LocalId,&node.Ip) == sql.ErrNoRows {
		return node, errors.New("Not Found")
	}

	return node, nil
}

func InsertNode(node Node) (int,error) {
	db := getDB()
	defer db.Close()

	tx,err := db.Begin()
	if err != nil {
		return -1, err
	}

	stmt,err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (test_net,server,local_id,ip) VALUES (?,?,?,?)",NodesTable))
	
	if err != nil {
		return -1, err
	}

	defer stmt.Close()

	res,err := stmt.Exec(node.TestNetId,node.Server,node.LocalId,node.Ip)
	if err != nil {
		return -1, nil
	}
	
	tx.Commit()
	id, err := res.LastInsertId()
	return int(id), err
}


func DeleteNode(id int) error {
	db := getDB()
	defer db.Close()

	_,err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = %d",NodesTable,id))
	return err
}

func DeleteNodesByTestNet(id int) error {
	db := getDB()
	defer db.Close()

	_,err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE test_net = %d",NodesTable,id))
	return err
}	

func DeleteNodesByServer(id int) error {
	db := getDB()
	defer db.Close()

	_, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE server = %d",NodesTable,id))
	return err
}


/*******COMMON QUERY FUNCTIONS********/

func GetAvailibleNodes(serverId int, nodesRequested int) ([]int,error){

	nodes,err := GetAllNodesByServer(serverId)
	if err != nil {
		return nil,err
	}
	server,_,err := GetServer(serverId)
	if err != nil {
		return nil,err
	}
	out := util.IntArrFill(server.Max,func(index int) int{
		return index
	})

	for _,node := range nodes {
		out = util.IntArrRemove(out,node.Id)
	}
	return out,nil
}