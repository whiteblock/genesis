/*
	Copyright 2019 whiteblock Inc.
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
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3" //Include sqlite as the db
	"github.com/whiteblock/genesis/util"
)

// Node represents a node within the network
type Node struct {
	// ID is the UUID of the node
	ID string `json:"id"`

	//AbsoluteNum is the number of the node in the testnet
	AbsoluteNum int `json:"absNum"`

	// TestNetId is the id of the testnet to which the node belongs to
	TestNetID string `json:"testnetId"`

	// Server is the id of the server on which the node resides
	Server int `json:"server"`

	// LocalID is the number of the node on the server it resides
	LocalID int `json:"localId"`

	// IP is the ip address of the node
	IP string `json:"ip"`

	// Label is the string given to the node by the build process
	Label string `json:"label"`

	// Image is the docker image used to build this node
	Image string `json:"image"`

	// Protocol is the protocol type of this node
	Protocol string `json:"protocol"`

	// PortMappings keeps tracks of the ports exposed externally on the for this
	// node
	PortMappings map[string]string `json:"portMappings,omitonempty"`
}

// GetID gets the id of this side car
func (n Node) GetID() string {
	return n.ID
}

// GetAbsoluteNumber gets the absolute number of the node in the testnet
func (n Node) GetAbsoluteNumber() int {
	return n.AbsoluteNum
}

// GetIP gets the ip address of this node
func (n Node) GetIP() string {
	return n.IP
}

// GetRelativeNumber gets the local id of the node
func (n Node) GetRelativeNumber() int {
	return n.LocalID
}

// GetServerID gets the id of the server on which this node resides
func (n Node) GetServerID() int {
	return n.Server
}

// GetTestNetID gets the id of the testnet this node is a part of
func (n Node) GetTestNetID() string {
	return n.TestNetID
}

// GetNodeName gets the whiteblock name of this node
func (n Node) GetNodeName() string {
	return fmt.Sprintf("%s%d", conf.NodePrefix, n.AbsoluteNum)
}

const selectFieldStr = "id,test_net,server,local_id,ip,label,abs_num,image,protocol,port_mappings"

func getNodesByQuery(query string) ([]Node, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, util.LogError(err)
	}
	defer rows.Close()

	nodes := []Node{}
	for rows.Next() {
		var node Node
		var rawNodeMapping []byte
		err := rows.Scan(&node.ID, &node.TestNetID, &node.Server, &node.LocalID, &node.IP,
			&node.Label, &node.AbsoluteNum, &node.Image, &node.Protocol, &rawNodeMapping)
		if err != nil {
			return nil, util.LogError(err)
		}
		err = json.Unmarshal(rawNodeMapping, &node.PortMappings)
		if err != nil {
			return nil, util.LogError(err)
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// GetAllNodesByServer gets all nodes that have ever existed on a server
func GetAllNodesByServer(serverID int) ([]Node, error) {
	return getNodesByQuery(fmt.Sprintf("SELECT %s FROM %s WHERE server = %d",
		selectFieldStr, NodesTable, serverID))
}

// GetAllNodesByTestNet gets all the nodes which are in the given testnet
func GetAllNodesByTestNet(testID string) ([]Node, error) {
	return getNodesByQuery(fmt.Sprintf("SELECT %s FROM %s WHERE test_net = \"%s\"",
		selectFieldStr, NodesTable, testID))
}

// GetAllNodes gets every node that has ever existed.
func GetAllNodes() ([]Node, error) {
	return getNodesByQuery(fmt.Sprintf("SELECT %s FROM %s", selectFieldStr, NodesTable))
}

// GetNode fetches a node by id
func GetNode(id string) (Node, error) {
	nodes, err := getNodesByQuery(
		fmt.Sprintf("SELECT %s FROM %s WHERE id = %s", selectFieldStr, NodesTable, id))

	if len(nodes) == 0 || err == sql.ErrNoRows {
		return Node{}, fmt.Errorf("node %s not found", id)
	}
	return nodes[0], nil
}

// InsertNode inserts a node into the database
func InsertNode(node Node) (int, error) {

	tx, err := db.Begin()
	if err != nil {
		return -1, util.LogError(err)
	}

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (%s) "+
		" VALUES (?,?,?,?,?,?,?,?,?,?)", NodesTable, selectFieldStr))

	if err != nil {
		return -1, util.LogError(err)
	}

	defer stmt.Close()

	rawNodeMapping, err := json.Marshal(node.PortMappings)
	if err != nil {
		return -1, util.LogError(err)
	}

	res, err := stmt.Exec(node.ID, node.TestNetID, node.Server, node.LocalID, node.IP, node.Label,
		node.AbsoluteNum, node.Image, node.Protocol, rawNodeMapping)
	if err != nil {
		return -1, util.LogError(err)
	}

	tx.Commit()
	id, err := res.LastInsertId()
	return int(id), util.LogError(err)
}

/**Helper functions which do not query the database**/

// GetNodeByLocalID looks up a node by its localID
func GetNodeByLocalID(nodes []Node, localID int) (Node, error) {
	for _, node := range nodes {
		if node.LocalID == localID {
			return node, nil
		}
	}

	return Node{}, fmt.Errorf("node %d not found", localID)
}

// GetNodeByAbsNum finds a node based on its absolute node number
func GetNodeByAbsNum(nodes []Node, absNum int) (Node, error) {
	for _, node := range nodes {
		if node.AbsoluteNum == absNum {
			return node, nil
		}
	}
	return Node{}, fmt.Errorf("node %d not found", absNum)
}

// DivideNodesByAbsMatch spits the given nodes into nodes which have their absnum in the
// given nodeNums and those who don't
func DivideNodesByAbsMatch(nodes []Node, nodeNums []int) ([]Node, []Node, error) {
	matches := []Node{}
	notMatches := make([]Node, len(nodes))
	copy(notMatches, nodes)
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

// GetUniqueServerIDs extracts the unique server ids from a slice of Node
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
