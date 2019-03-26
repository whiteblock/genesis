package db

import(
    _ "github.com/mattn/go-sqlite3"   
    "log"
    "fmt"
    "errors"
    util "../util"
)

const Version = "2.1.0"

func Check() error {
    row := db.QueryRow(fmt.Sprintf("SELECT value FROM meta WHERE key = \"version\""))
    var version string
    err := row.Scan(&version)
    if err != nil{
        log.Println(err)
        return err
    }
    if version != Version {
        //Old version, previous database is now invalid
        return errors.New("Needs update")
    }
    return nil
}

func CheckAndUpdate() {
    if Check() != nil {
        log.Println("Updating the database...")
        util.Rm(dataLoc)
        db = getDB()
        log.Println("Database update finished")
    }
}

func SetVersion(version string) error {
    tx,err := db.Begin()

    if err != nil{
        log.Println(err)
        return err
    }

    stmt,err := tx.Prepare(fmt.Sprintf("INSERT INTO meta (key,value) VALUES (?,?)"))
    
    if err != nil{
        log.Println(err)
        return err
    }

    defer stmt.Close()


    _,err = stmt.Exec("version",version)
    
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