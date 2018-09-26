package db

import(
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"fmt"
	"errors"
	util "../util"
)

type Switch struct {
	Addr 	string
	Iface	string
	Brand	int
	Id		int
}

func GetAllSwitches() []Switch {
	db := getDB()
	defer db.Close()

	rows, err :=  db.Query(fmt.Sprintf("SELECT id,addr,iface,brand FROM %s",SwitchTable ))
	util.CheckFatal(err)
	defer rows.Close()
	switches := []Switch{}

	for rows.Next() {
		var swtch Switch
		util.CheckFatal(rows.Scan(&swtch.Id,&swtch.Addr,&swtch.Iface,&swtch.Brand))
		switches = append(switches,swtch)
	}
	return switches
}

func GetSwitchById(id int) (Switch, error) {
	db := getDB()
	defer db.Close()

	row := db.QueryRow(fmt.Sprintf("SELECT id,addr,iface,brand FROM %s WHERE id = %d",SwitchTable,id))

	var swtch Switch


	if row.Scan(&swtch.Id,&swtch.Addr,&swtch.Iface,&swtch.Brand) == sql.ErrNoRows {
		return swtch, errors.New("Not Found")
	}

	return swtch, nil
}

func GetSwitchByIP(ip string) (Switch, error) {
	db := getDB()
	defer db.Close()

	row := db.QueryRow(fmt.Sprintf("SELECT id,addr,iface,brand FROM %s WHERE addr = %s",SwitchTable,ip))
	
	var swtch Switch

	if row.Scan(&swtch.Id,&swtch.Addr,&swtch.Iface,&swtch.Brand) == sql.ErrNoRows {
		return swtch, errors.New("Not Found")
	}

	return swtch, nil
}

func InsertSwitch(swtch Switch) int {
	db := getDB()
	defer db.Close()

	tx,err := db.Begin()
	util.CheckFatal(err)

	stmt,err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (addr,iface,brand) VALUES (?,?,?)",SwitchTable))
	util.CheckFatal(err)

	defer stmt.Close()

	res,err := stmt.Exec(swtch.Addr,swtch.Iface,swtch.Brand)
	util.CheckFatal(err)

	tx.Commit()
	id, err := res.LastInsertId()
	util.CheckFatal(err)

	return int(id)
}

func DeleteSwitch(id int){
	db := getDB()
	defer db.Close()

	db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = %d",SwitchTable,id))
}

func UpdateSwitch(id int,swtch Switch){
	db := getDB()
	defer db.Close()

	tx,err := db.Begin()
	util.CheckFatal(err)

	stmt,err := tx.Prepare(fmt.Sprintf("UPDATE %s SET addr = ?, iface = ?, brand = ? WHERE id = ? ",SwitchTable))
	util.CheckFatal(err)
	defer stmt.Close()

	_,err = stmt.Exec(swtch.Addr,swtch.Iface,swtch.Brand,swtch.Id)
	util.CheckFatal(err)
	tx.Commit()
}