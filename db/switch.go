package db

import(
    _ "github.com/mattn/go-sqlite3"
    "database/sql"
    "fmt"
    "errors"
    "regexp"
    "log"
)

type Switch struct {
    Addr    string  `json:"addr"`
    Iface   string  `json:"iface"`
    Brand   int     `json:"brand"`
    Id      int     `json:"id"`
}

func (s Switch) Validate() error {
    var re = regexp.MustCompile(`(?m)[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`)
    if !re.Match([]byte(s.Addr)) {
        return errors.New("Switch Addr is invalid")
    }
    if s.Brand < 1 && s.Brand > 2{
        return errors.New("Invalid Brand: Currently have 1-Vyos, 2-HP")
    }
    return nil
}

func GetAllSwitches() ([]Switch,error) {

    rows, err :=  db.Query(fmt.Sprintf("SELECT id,addr,iface,brand FROM %s",SwitchTable ))
    if err != nil{
        log.Println(err)
        return nil,err
    }
    defer rows.Close()
    switches := []Switch{}

    for rows.Next() {
        var swtch Switch
        err = rows.Scan(&swtch.Id,&swtch.Addr,&swtch.Iface,&swtch.Brand)
        if err != nil{
            log.Println(err)
            return nil,err
        }
        switches = append(switches,swtch)
    }
    return switches,nil
}

func GetSwitchById(id int) (Switch, error) {

    row := db.QueryRow(fmt.Sprintf("SELECT id,addr,iface,brand FROM %s WHERE id = %d",SwitchTable,id))

    var swtch Switch

    if row.Scan(&swtch.Id,&swtch.Addr,&swtch.Iface,&swtch.Brand) == sql.ErrNoRows {
        return swtch, errors.New("Not Found")
    }

    return swtch, nil
}

func GetSwitchByIP(ip string) (Switch, error) {

    row := db.QueryRow(fmt.Sprintf("SELECT id,addr,iface,brand FROM %s WHERE addr = %s",SwitchTable,ip))
    
    var swtch Switch

    if row.Scan(&swtch.Id,&swtch.Addr,&swtch.Iface,&swtch.Brand) == sql.ErrNoRows {
        return swtch, errors.New("Not Found")
    }

    return swtch, nil
}

func InsertSwitch(swtch Switch) (int,error) {

    tx,err := db.Begin()

    if err != nil{
        log.Println(err)
        return -1,err
    }

    stmt,err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (addr,iface,brand) VALUES (?,?,?)",SwitchTable))
    
    if err != nil{
        log.Println(err)
        return -1,err
    }

    defer stmt.Close()

    res,err := stmt.Exec(swtch.Addr,swtch.Iface,swtch.Brand)
    
    if err != nil{
        log.Println(err)
        return -1,err
    }

    err = tx.Commit()
    if err != nil{
        log.Println(err)
        return -1,err
    }
    id, err := res.LastInsertId();
    if err != nil{
        log.Println(err)
        return -1,err
    }

    return int(id),nil
}

func DeleteSwitch(id int) error {

    _,err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = %d",SwitchTable,id))
    return err
}

func UpdateSwitch(id int,swtch Switch) error {

    tx,err := db.Begin()
    if err != nil{
        log.Println(err)
        return err
    }

    stmt,err := tx.Prepare(fmt.Sprintf("UPDATE %s SET addr = ?, iface = ?, brand = ? WHERE id = ? ",SwitchTable))
    if err != nil{
        log.Println(err)
        return err
    }
    defer stmt.Close()

    _,err = stmt.Exec(swtch.Addr,swtch.Iface,swtch.Brand,swtch.Id)
    if err != nil{
        log.Println(err)
        return err
    }
    return tx.Commit()
}