package db

import(
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"fmt"
	"errors"
	"log"
)

type TestNet struct {
	Id			string  `json:"id"`
	Blockchain	string	`json:"blockchain"`
	Nodes		int		`json:"nodes"`
	Image		string	`json:"image"`
	Ts			int64 	`json:"timestamp"`
}

/*
	Get all of the testnets
 */
func GetAllTestNets() ([]TestNet,error) {

	rows, err :=  db.Query(fmt.Sprintf("SELECT id, blockchain, nodes, image, ts FROM %s",TestTable ))
	if err != nil{
		log.Println(err)
		return nil,err
	}
	defer rows.Close()
	testnets := []TestNet{}

	for rows.Next() {
		var testnet TestNet
		err = rows.Scan(&testnet.Id,&testnet.Blockchain,&testnet.Nodes,&testnet.Image,&testnet.Ts)
		if err != nil{
			log.Println(err)
			return nil,err
		}
		testnets = append(testnets,testnet)
	}
	return testnets,nil
}

/*
	Get a testnet by id
 */
func GetTestNet(id string) (TestNet,error) {

	row :=  db.QueryRow(fmt.Sprintf("SELECT id,blockchain,nodes,image FROM %s WHERE id = \"%s\"",TestTable,id))

	var testnet TestNet

	if row.Scan(&testnet.Id,&testnet.Blockchain,&testnet.Nodes,&testnet.Image,&testnet.Ts) == sql.ErrNoRows {
		return testnet, errors.New("Not Found")
	}

	return testnet, nil
}

/*
	Insert a testnet into the database
 */
func InsertTestNet(testnet TestNet) (error) {

	tx,err := db.Begin()
	if err != nil{
		log.Println(err)
		return err
	}

	stmt,err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (id,blockchain,nodes,image,ts) VALUES (?,?,?,?,?)",TestTable))
	if err != nil{
		log.Println(err)
		return err
	}

	defer stmt.Close()

	_,err = stmt.Exec(testnet.Id,testnet.Blockchain,testnet.Nodes,testnet.Image,testnet.Ts)
	if err != nil{
		log.Println(err)
		return err
	}

	err = tx.Commit()
	if err != nil{
		log.Println(err)
		return err
	}

	return nil
}

/*
	Delete a testnet by id
*/
func DeleteTestNet(id int) error {

	_,err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = %d",TestTable,id))
	if err != nil{
		log.Println(err)
		return err
	}
	DeleteNodesByTestNet(id)
	return nil
}

/*
	Update a testnet by id
*/
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

/*
	Update the number of nodes in a testnet
*/
func UpdateTestNetNodes(id int,nodes int) error {
	_, err := db.Exec(fmt.Sprintf("UPDATE %s SET nodes = %d WHERE id = %d",TestTable,id,nodes))
	return err
}

