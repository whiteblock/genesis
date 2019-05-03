package db

import (
	"../util"
	"fmt"
	_ "github.com/mattn/go-sqlite3" //use sqlite
	"log"
)

// Version represents the database version, upon change of this constant, the database will
// be purged
const Version = "2.2.4"

func check() error {
	row := db.QueryRow("SELECT value FROM meta WHERE key = \"version\"")
	var version string
	err := row.Scan(&version)
	if err != nil {
		return util.LogError(err)
	}
	if version != Version {
		//Old version, previous database is now invalid
		return fmt.Errorf("needs update")
	}
	return nil
}

func checkAndUpdate() {
	if check() != nil {
		log.Println("Updating the database...")
		util.Rm(dataLoc)
		db = getDB()
		log.Println("Database update finished")
	}
}

func setVersion(version string) error {
	tx, err := db.Begin()

	if err != nil {
		return util.LogError(err)
	}

	stmt, err := tx.Prepare("INSERT INTO meta (key,value) VALUES (?,?)")

	if err != nil {
		return util.LogError(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec("version", version)

	if err != nil {
		return util.LogError(err)
	}

	err = tx.Commit()
	if err != nil {
		return util.LogError(err)
	}

	return nil
}
