/*
	Copyright 2019 Whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Genesis is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package db

import (
	"github.com/Whiteblock/genesis/util"
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
		log.Println("Updating the databasegithub.com/Whiteblock/genesis.")
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
