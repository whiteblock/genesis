package db

import(
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"fmt"
	"errors"
)

type TestNet struct {
	Id			int
	Blockchain	string
	Nodes		int
	Image		string
}

func GetAllTestNets() []TestNet {
	
	db := getDB()
	defer db.Close()

	rows, err :=  db.Query(fmt.Sprintf("SELECT id, blockchain, nodes, image FROM %s",TEST_TABLE ))
	checkFatal(err)
	defer rows.Close()
	testnets := []TestNet{}

	for rows.Next() {
		var testnet TestNet
		checkFatal(rows.Scan(&testnet.Id,&testnet.Blockchain,&testnet.Nodes,&testnet.Image))
		testnets = append(testnets,testnet)
	}
	return testnets
}

func GetTestNet(id int) (TestNet,error) {
	db := getDB()
	defer db.Close()

	row :=  db.QueryRow(fmt.Sprintf("SELECT id,blockchain,nodes,image FROM %s WHERE id = %d",TEST_TABLE,id))

	var testnet TestNet

	if row.Scan(&testnet.Id,&testnet.Blockchain,&testnet.Nodes,&testnet.Image) == sql.ErrNoRows {
		return testnet, errors.New("Not Found")
	}

	return testnet, nil
}

func InsertTestNet(testnet TestNet) int {
	db := getDB()
	defer db.Close()

	tx,err := db.Begin()
	checkFatal(err)

	stmt,err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (blockchain,nodes,image) VALUES (?,?,?)",TEST_TABLE))
	checkFatal(err)

	defer stmt.Close()

	res,err := stmt.Exec(testnet.Blockchain,testnet.Nodes,testnet.Image)
	checkFatal(err)
	tx.Commit()

	id, err := res.LastInsertId()
	checkFatal(err)

	return int(id)
}


func DeleteTestNet(id int){
	db := getDB()
	defer db.Close()

	db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = %d",TEST_TABLE,id))
	DeleteNodesByTestNet(id)
}

func UpdateTestNet(id int,testnet TestNet){
	db := getDB()
	defer db.Close()

	tx,err := db.Begin()
	checkFatal(err)

	stmt,err := tx.Prepare(fmt.Sprintf("UPDATE %s SET blockchain = ?, nodes = ?, image = ? WHERE id = ? ",TEST_TABLE))
	checkFatal(err)
	defer stmt.Close()

	_,err = stmt.Exec(testnet.Blockchain,testnet.Nodes,testnet.Image,testnet.Id)
	checkFatal(err)
	tx.Commit()
}

func UpdateTestNetNodes(id int,nodes int){
	db := getDB()
	defer db.Close()
	db.Exec(fmt.Sprintf("UPDATE %s SET nodes = %d WHERE id = %d",TEST_TABLE,id,nodes))
}