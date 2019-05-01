package db

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3" //Include sqlite as the db
	"log"
)

//Node represents a node within the network
type Node struct {
	ID string `json:"id"`

	AbsoluteNum int `json:"absNum"`
	/*
	   TestNetId is the id of the testnet to which the node belongs to
	*/
	TestNetID string `json:"testnetId"`
	/*
	   Server is the id of the server on which the node resides
	*/
	Server int `json:"server"`
	/*
	   LocalId is the number of the node in the testnet
	*/
	LocalID int `json:"localId"`
	/*
	   IP is the ip address of the node
	*/
	IP string `json:"ip"`
	/*
	   Label is the string given to the node by the build process
	*/
	Label string `json:"label"`

	Image string `json:"image"`

	Blockchain string `json:"blockchain"`
}

//GetAllNodesByServer gets all nodes that have ever existed on a server
func GetAllNodesByServer(serverID int) ([]Node, error) {

	rows, err := db.Query(fmt.Sprintf("SELECT id,test_net,server,local_id,ip,label,abs_num FROM %s WHERE server = %d", NodesTable, serverID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := []Node{}
	for rows.Next() {
		var node Node
		err := rows.Scan(&node.ID, &node.TestNetID, &node.Server, &node.LocalID, &node.ID, &node.Label, &node.AbsoluteNum)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

//GetAllNodesByTestNet gets all the nodes which are in the given testnet
func GetAllNodesByTestNet(testID string) ([]Node, error) {
	nodes := []Node{}

	rows, err := db.Query(fmt.Sprintf("SELECT id,test_net,server,local_id,ip,label,abs_num FROM %s WHERE test_net = \"%s\"", NodesTable, testID))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var node Node
		err := rows.Scan(&node.ID, &node.TestNetID, &node.Server, &node.LocalID, &node.IP, &node.Label, &node.AbsoluteNum)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

//GetAllNodes gets every node that has ever existed.
func GetAllNodes() ([]Node, error) {

	rows, err := db.Query(fmt.Sprintf("SELECT id,test_net,server,local_id,ip,label,abs_num FROM %s", NodesTable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	nodes := []Node{}

	for rows.Next() {
		var node Node
		err := rows.Scan(&node.ID, &node.TestNetID, &node.Server, &node.LocalID, &node.IP, &node.Label, &node.AbsoluteNum)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

//GetNode fetches a node by id
func GetNode(id string) (Node, error) {

	row := db.QueryRow(fmt.Sprintf("SELECT id,test_net,server,local_id,ip,label,abs_num FROM %s WHERE id = %s", NodesTable, id))

	var node Node

	if row.Scan(&node.ID, &node.TestNetID, &node.Server, &node.LocalID, &node.IP, &node.Label, &node.AbsoluteNum) == sql.ErrNoRows {
		return node, errors.New("not found")

	}

	return node, nil
}

//InsertNode inserts a node into the database
func InsertNode(node Node) (int, error) {

	tx, err := db.Begin()
	if err != nil {
		return -1, err
	}

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (id,test_net,server,local_id,ip,label,abs_num) VALUES (?,?,?,?,?,?,?)", NodesTable))

	if err != nil {
		return -1, err
	}

	defer stmt.Close()

	res, err := stmt.Exec(node.ID, node.TestNetID, node.Server, node.LocalID, node.IP, node.Label, node.AbsoluteNum)
	if err != nil {
		return -1, nil
	}

	tx.Commit()
	id, err := res.LastInsertId()
	return int(id), err
}

// DeleteNode removes a node from the database
// (Deprecated)
func DeleteNode(id string) error {

	_, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = %s", NodesTable, id))
	return err
}

// DeleteNodesByTestNet removes all nodes in a testnet from the database.
// (Deprecated)
func DeleteNodesByTestNet(id string) error {

	_, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE test_net = %s", NodesTable, id))
	return err
}

//DeleteNodesByServer delete all nodes which have ever been on a given server.
func DeleteNodesByServer(id string) error {

	_, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE server = %s", NodesTable, id))
	return err
}

/**Helper functions which do not query the database**/

// GetNodeByLocalID looks up a node by its localID
func GetNodeByLocalID(nodes []Node, localID int) (Node, error) {
	for _, node := range nodes {
		if node.LocalID == localID {
			return node, nil
		}
	}

	return Node{}, errors.New("node not found")
}

//GetNodeByAbsNum finds a node based on its absolute node number
func GetNodeByAbsNum(nodes []Node, absNum int) (Node, error) {
	for _, node := range nodes {
		if node.AbsoluteNum == absNum {
			return node, nil
		}
	}

	return Node{}, errors.New("node not found")
}

//DivideNodesByAbsMatch spits the given nodes into nodes which have their absnum in the
//given nodeNums and those who don't
func DivideNodesByAbsMatch(nodes []Node, nodeNums []int) ([]Node, []Node, error) {
	matches := []Node{}
	notMatches := make([]Node, len(nodes))
	copy(notMatches, nodes)
	fmt.Printf("%#v\n", notMatches)
	for {
		num := nodeNums[0]
		index := -1
		for i, node := range notMatches {
			if node.AbsoluteNum == num {
				index = i
				break
			}
		}
		if index == -1 {
			return nil, nil, fmt.Errorf("node %d not found", num)
		}
		matches = append(matches, notMatches[index])
		if len(notMatches) == index-1 {
			notMatches = notMatches[:index]
		} else {
			notMatches = append(notMatches[:index], notMatches[index+1:]...)
		}

		if len(nodeNums) == 1 {
			break
		}
		nodeNums = nodeNums[1:]

	}
	return matches, notMatches, nil
}

//GetUniqueServerIDs extracts the unique server ids from a slice of Node
func GetUniqueServerIDs(nodes []Node) []int {
	out := []int{}
	for _, node := range nodes {
		shouldAdd := true
		for _, serverID := range out { //Check to make sure the serverID is not already in out
			if node.Server == serverID {
				shouldAdd = false
			}
		}
		if shouldAdd {
			out = append(out, node.Server)
		}
	}
	return out
}
