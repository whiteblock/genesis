package db

import(
    _ "github.com/mattn/go-sqlite3"
    "database/sql"
    "fmt"
    "errors"
    util "../util"
)

/*
    Node represents a node within the network
 */
type Node struct {  
    Id          int     `json:"id"`
    /*
        TestNetId is the id of the testnet to which the node belongs to
     */
    TestNetId   int     `json:"testNetId"`
    /*
        Server is the id of the server on which the node resides
     */
    Server      int     `json:"server"`
    /*
        LocalId is the number of the node in the testnet
     */
    LocalId     int     `json:"localId"`
    /*
        Ip is the ip address of the node
     */
    Ip          string  `json:"ip"`
    /*
        Label is the string given to the node by the build process
     */
    Label       string  `json:"label"`
}

/*
    GetAllNodesByServer gets all nodes that have ever existed on a server
 */
func GetAllNodesByServer(serverId int) ([]Node,error) {

    rows, err :=  db.Query(fmt.Sprintf("SELECT id,test_net,server,local_id,ip,label FROM %s WHERE server = %d",NodesTable ))
    if err != nil {
        return nil,err
    }
    defer rows.Close()
    
    nodes := []Node{}
    for rows.Next() {
        var node Node
        err := rows.Scan(&node.Id,&node.TestNetId,&node.Server,&node.LocalId,&node.Ip,&node.Label)
        if err != nil {
            return nil,err
        }
        nodes = append(nodes,node)
    }
    return nodes,nil
}

/*
    GetAllNodesByTestNet gets all the nodes which are in the given testnet
 */
func GetAllNodesByTestNet(testId int) ([]Node,error) {
    nodes := []Node{}

    rows, err :=  db.Query(fmt.Sprintf("SELECT id,test_net,server,local_id,ip,label FROM %s WHERE test_net = %d",NodesTable,testId ))
    if err != nil {
        return nil,err
    }
    defer rows.Close()

    
    for rows.Next() {
        var node Node
        err := rows.Scan(&node.Id,&node.TestNetId,&node.Server,&node.LocalId,&node.Ip,&node.Label)
        if err != nil {
            return nil, err
        }
        nodes = append(nodes,node)
    }
    return nodes, nil
}

/*
    GetAllNodes gets every node that has ever existed.
 */
func GetAllNodes() ([]Node,error) {

    rows, err :=  db.Query(fmt.Sprintf("SELECT id,test_net,server,local_id,ip,label FROM %s",NodesTable ))
    if err != nil {
        return nil,err
    }
    defer rows.Close()
    nodes := []Node{}

    for rows.Next() {
        var node Node
        err := rows.Scan(&node.Id,&node.TestNetId,&node.Server,&node.LocalId,&node.Ip,&node.Label)
        if err != nil {
            return nil, err
        }
        nodes = append(nodes,node)
    }
    return nodes,nil
}

/*
    GetNode fetches a node by id
 */
func GetNode(id int) (Node,error) {

    row :=  db.QueryRow(fmt.Sprintf("SELECT id,test_net,server,local_id,ip,label FROM %s WHERE id = %d",NodesTable,id))

    var node Node

    if row.Scan(&node.Id,&node.TestNetId,&node.Server,&node.LocalId,&node.Ip,&node.Label) == sql.ErrNoRows {
        return node, errors.New("Not Found")
    }

    return node, nil
}

/*
    InsertNode inserts a node into the database
 */
func InsertNode(node Node) (int,error) {

    tx,err := db.Begin()
    if err != nil {
        return -1, err
    }

    stmt,err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (test_net,server,local_id,ip,label) VALUES (?,?,?,?,?)",NodesTable))
    
    if err != nil {
        return -1, err
    }

    defer stmt.Close()

    res,err := stmt.Exec(node.TestNetId,node.Server,node.LocalId,node.Ip,node.Label)
    if err != nil {
        return -1, nil
    }
    
    tx.Commit()
    id, err := res.LastInsertId()
    return int(id), err
}

/*
    DeleteNode removes a node from the database
    (Deprecated)
 */
func DeleteNode(id int) error {

    _,err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = %d",NodesTable,id))
    return err
}

/*
    DeleteNodesByTestNet removes all nodes in a testnet from the database. 
    (Deprecated)
 */
func DeleteNodesByTestNet(id int) error {

    _,err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE test_net = %d",NodesTable,id))
    return err
}   

/*
    DeleteNodesByServer delete all nodes which have ever been on a given server.
 */
func DeleteNodesByServer(id int) error {

    _, err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE server = %d",NodesTable,id))
    return err
}


/*******COMMON QUERY FUNCTIONS********/

/*
    GetAvailibleNodes gets the node numbers which are availible on a server,
    up to nodesRequested. 
    (Deprecated)
 */
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