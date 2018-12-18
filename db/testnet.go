package db

import(
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"fmt"
	"errors"
	"log"
)

type TestNet struct {
	Id			int		`json:"id"`
	Blockchain	string	`json:"blockchain"`
	Nodes		int		`json:"nodes"`
	Image		string	`json:"image"`
}


func GetAllTestNets() ([]TestNet,error) {

	rows, err :=  db.Query(fmt.Sprintf("SELECT id, blockchain, nodes, image FROM %s",TestTable ))
	if err != nil{
		log.Println(err)
		return nil,err
	}
	defer rows.Close()
	testnets := []TestNet{}

	for rows.Next() {
		var testnet TestNet
		err = rows.Scan(&testnet.Id,&testnet.Blockchain,&testnet.Nodes,&testnet.Image)
		if err != nil{
			log.Println(err)
			return nil,err
		}
		testnets = append(testnets,testnet)
	}
	return testnets,nil
}

func GetTestNet(id int) (TestNet,error) {

	row :=  db.QueryRow(fmt.Sprintf("SELECT id,blockchain,nodes,image FROM %s WHERE id = %d",TestTable,id))

	var testnet TestNet

	if row.Scan(&testnet.Id,&testnet.Blockchain,&testnet.Nodes,&testnet.Image) == sql.ErrNoRows {
		return testnet, errors.New("Not Found")
	}

	return testnet, nil
}

func InsertTestNet(testnet TestNet) (int,error) {

	tx,err := db.Begin()
	if err != nil{
		log.Println(err)
		return -1,err
	}

	stmt,err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (blockchain,nodes,image) VALUES (?,?,?)",TestTable))
	if err != nil{
		log.Println(err)
		return -1,err
	}

	defer stmt.Close()

	res,err := stmt.Exec(testnet.Blockchain,testnet.Nodes,testnet.Image)
	if err != nil{
		log.Println(err)
		return -1,err
	}

	err = tx.Commit()
	if err != nil{
		log.Println(err)
		return -1,err
	}

	id, err := res.LastInsertId()
	if err != nil{
		log.Println(err)
		return -1,err
	}

	return int(id),nil
}


func DeleteTestNet(id int) error {

	_,err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = %d",TestTable,id))
	if err != nil{
		log.Println(err)
		return err
	}
	DeleteNodesByTestNet(id)
	return nil
}

func UpdateTestNet(id int,testnet TestNet) error {

	tx,err := db.Begin()
	if err != nil{
		log.Println(err)
		return err
	}

	stmt,err := tx.Prepare(fmt.Sprintf("UPDATE %s SET blockchain = ?, nodes = ?, image = ? WHERE id = ? ",TestTable))
	if err != nil{
		log.Println(err)
		return err
	}

	defer stmt.Close()

	_,err = stmt.Exec(testnet.Blockchain,testnet.Nodes,testnet.Image,testnet.Id)
	if err != nil{
		log.Println(err)
		return err
	}

	return tx.Commit()

}

func UpdateTestNetNodes(id int,nodes int) error {
	_, err := db.Exec(fmt.Sprintf("UPDATE %s SET nodes = %d WHERE id = %d",TestTable,id,nodes))
	return err
}

