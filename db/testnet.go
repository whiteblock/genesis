package db

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type TestNet struct {
	Id         string `json:"id"`
	Blockchain string `json:"blockchain"`
	Nodes      int    `json:"nodes"`
	Image      string `json:"image"`
	Ts         int64  `json:"timestamp"`
}

/*
	Get all of the testnets
*/
func GetAllTestNets() ([]TestNet, error) {

	rows, err := db.Query(fmt.Sprintf("SELECT id, blockchain, nodes, image, ts FROM %s", TestTable))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()
	testnets := []TestNet{}

	for rows.Next() {
		var testnet TestNet
		err = rows.Scan(&testnet.Id, &testnet.Blockchain, &testnet.Nodes, &testnet.Image, &testnet.Ts)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		testnets = append(testnets, testnet)
	}
	return testnets, nil
}

/*
	Get a testnet by id
*/
func GetTestNet(id string) (TestNet, error) {

	row := db.QueryRow(fmt.Sprintf("SELECT id,blockchain,nodes,image,ts FROM %s WHERE id = '%s'", TestTable, id))

	var testnet TestNet

	err := row.Scan(&testnet.Id, &testnet.Blockchain, &testnet.Nodes, &testnet.Image, &testnet.Ts)

	if err == sql.ErrNoRows {
		return testnet, errors.New("Not Found")
	}

	return testnet, err
}

/*
	Insert a testnet into the database
*/
func InsertTestNet(testnet TestNet) error {

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return err
	}

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (id,blockchain,nodes,image,ts) VALUES (?,?,?,?,?)", TestTable))
	if err != nil {
		log.Println(err)
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(testnet.Id, testnet.Blockchain, testnet.Nodes, testnet.Image, testnet.Ts)
	if err != nil {
		log.Println(err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

/*
	Update a testnet by id
*/
func UpdateTestNet(id int, testnet TestNet) error {

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return err
	}

	stmt, err := tx.Prepare(fmt.Sprintf("UPDATE %s SET blockchain = ?, nodes = ?, image = ? WHERE id = ? ", TestTable))
	if err != nil {
		log.Println(err)
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(testnet.Blockchain, testnet.Nodes, testnet.Image, testnet.Id)
	if err != nil {
		log.Println(err)
		return err
	}

	return tx.Commit()

}

/*
	Update the number of nodes in a testnet
*/
func UpdateTestNetNodes(id int, nodes int) error {
	_, err := db.Exec(fmt.Sprintf("UPDATE %s SET nodes = %d WHERE id = %d", TestTable, id, nodes))
	return err
}
