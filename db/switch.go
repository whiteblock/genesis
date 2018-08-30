package db;

import(
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"fmt"
	"errors"
)

type Switch struct {
	Addr 	string
	Iface	string
	Brand	int
	Id		int
}

func GetAllSwitches() Switch[] {
	db := getDB()
	defer db.Close()

	rows, err :=  db.Query(fmt.Sprintf("SELECT id,addr,iface,brand FROM %s",SWITCH_TABLE ))
	checkFatal(err)
	defer rows.Close()
	switches := Switch[]{}

	for rows.Next() {
		var swtch Switch
		checkFatal(rows.Scan(&swtch.Id,&swtch.Addr,&swtch.Iface,&swtch.Brand))
		switches = append(switches,swtch)
	}
	return allServers
}

func GetSwitchById(id int) (Switch, error) {
	db := getDB()
	defer db.Close()

	row, err :=  db.QueryRow(fmt.Sprintf("SELECT id,addr,iface,brand FROM %s WHERE id = %d",SWITCH_TABLE,id))
	checkFatal(err)
	var swtch Switch

	defer row.Close()

	if row.Scan(&swtch.Id,&swtch.Addr,&swtch.Iface,&swtch.Brand) == sql.ErrNoRows {
		return swtch, errors.New("Not Found")
	}

	return swtch, nil
}

func GetSwitchByIP(ip string) (Switch, error) {
	db := getDB()
	defer db.Close()

	rows, err :=  db.QueryRow(fmt.Sprintf("SELECT id,addr,iface,brand FROM %s WHERE addr = %s",SWITCH_TABLE,ip))
	checkFatal(err)
	var swtch Switch

	defer row.Close()

	if row.Scan(&swtch.Id,&swtch.Addr,&swtch.Iface,&swtch.Brand) == sql.ErrNoRows {
		return swtch, errors.New("Not Found")
	}

	return swtch, nil
}

func InsertSwitch(swtch Switch) int {
	db := getDB()
	defer db.Close()

	tx,err := db.Begin()
	checkFatal(err)

	stmt,err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (addr,iface,brand) VALUES (?,?,?,?)",SWITCH_TABLE))
	checkFatal(err)

	defer stmt.Close()

	res,err := stmt.Exec(swtch.Addr,swtch.Iface,swtch.Brand)
	checkFatal(err)

	tx.Commit()
	return int(res.LastInsertId())
}

func DeleteSwitch(id int){
	db := getDB()
	defer db.Close()

	db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = %d",SWITCH,id))
}

func UpdateSwitch(id int,swtch Switch){
	db := getDB()
	defer db.Close()

	tx,err := db.Begin()
	checkFatal(err)

	stmt,err := tx.Prepare(fmt.Sprintf("UPDATE %s SET addr = ?, iface = ?, brand = ? WHERE id = ? ",SWITCH_TABLE))
	checkFatal(err)
	defer stmt.Close()

	_,err := stmt.Exec(swtch.Addr,swtch.Iface,swtch.Brand,swtch.Id)
	checkFatal(err)
	tx.Commit()
}