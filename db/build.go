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
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3" //Bring db in
	"log"
)

/*
DeploymentDetails represents the data for the construction of a testnet.
*/
type DeploymentDetails struct {
	/*
	   Servers: The ids of the servers to build on
	*/
	Servers []int `json:"servers"`
	/*
	   Blockchain: The blockchain to build.
	*/
	Blockchain string `json:"blockchain"`
	/*
	   Nodes:  The number of nodes to build
	*/
	Nodes int `json:"nodes"`
	/*
	   Image: The docker image to build off of
	*/
	Images []string `json:"images"`
	/*
	   Params: The blockchain specific parameters
	*/
	Params map[string]interface{} `json:"params"`
	/*
	   Resources: The resources per node
	*/
	Resources []util.Resources `json:"resources"`
	/*
		Environments is the environment variables to be passed to each node.
		If it doesn't exist for a node, it defaults first to index 0.
	*/
	Environments []map[string]string `json:"environments"`
	/*
		Custom files for each node
	*/
	Files []map[string]string `json:"files"`
	/*
		Logs to keep track of for each node
	*/
	Logs []map[string]string `json:"logs"`

	/*
		Fairly Arbitrary extras for when additional customizations are added.
	*/
	Extras map[string]interface{} `json:"extras"`
	jwt    string
	kid    string
}

//SetJwt stores the callers jwt
func (dd *DeploymentDetails) SetJwt(jwt string) error {
	dd.jwt = jwt
	kid, err := util.GetKidFromJwt(dd.GetJwt())

	dd.kid = kid
	return err
}

//GetJwt gets the jwt of the creator of this build
func (dd DeploymentDetails) GetJwt() string {
	return dd.jwt
}

//GetKid gets the kid of the jwt of the creator of this build
func (dd DeploymentDetails) GetKid() string {
	return dd.kid
}

//QueryBuilds fetches DeploymentDetails based on the given SQL select query
func QueryBuilds(query string) ([]DeploymentDetails, error) {
	rows, err := db.Query(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()
	builds := []DeploymentDetails{}

	for rows.Next() {
		var build DeploymentDetails
		var servers []byte
		var params []byte
		var resources []byte
		var environment []byte
		var logs []byte
		var extras []byte
		var images []byte
		var files []byte

		err = rows.Scan(&servers, &build.Blockchain, &build.Nodes, &images, &params, &resources, &files, &environment, &logs, &extras, &build.kid)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		err = json.Unmarshal(servers, &build.Servers)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		err = json.Unmarshal(params, &build.Params)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		err = json.Unmarshal(resources, &build.Resources)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		err = json.Unmarshal(files, &build.Files)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		err = json.Unmarshal(environment, &build.Environments)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		err = json.Unmarshal(logs, &build.Logs)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		err = json.Unmarshal(extras, &build.Extras)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		err = json.Unmarshal(images, &build.Images)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		builds = append(builds, build)
	}
	return builds, nil
}

/*
GetAllBuilds gets all of the builds done by a user
*/
func GetAllBuilds() ([]DeploymentDetails, error) {
	return QueryBuilds(fmt.Sprintf("SELECT servers,blockchain,nodes,image,params,resources,files,environment,logs,extras,kid FROM %s", BuildsTable))
}

/*
GetBuildByTestnet gets the build paramters based off testnet id
*/
func GetBuildByTestnet(id string) (DeploymentDetails, error) {

	details, err := QueryBuilds(fmt.Sprintf("SELECT servers,blockchain,nodes,image,params,resources,files,environment,logs,extras,kid FROM %s WHERE testnet = \"%s\"", BuildsTable, id))
	if err != nil {
		log.Println(err)
		return DeploymentDetails{}, err
	}
	if len(details) == 0 {
		return DeploymentDetails{}, fmt.Errorf("no results found")
	}
	return details[0], nil
}

//GetLastBuildByKid gets the build paramters based off kid
func GetLastBuildByKid(kid string) (DeploymentDetails, error) {

	details, err := QueryBuilds(fmt.Sprintf(
		"SELECT servers,blockchain,nodes,image,params,resources,files,environment,logs,extras,kid FROM %s"+
			" WHERE kid = \"%s\" ORDER BY id DESC LIMIT 1", BuildsTable, kid))
	if err != nil {
		log.Println(err)
		return DeploymentDetails{}, err
	}
	if len(details) == 0 {
		return DeploymentDetails{}, fmt.Errorf("no results found")
	}
	return details[0], nil
}

//InsertBuild inserts a build
func InsertBuild(dd DeploymentDetails, testnetID string) error {

	tx, err := db.Begin()

	if err != nil {
		log.Println(err)
		return err
	}

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (testnet,servers,blockchain,nodes,image,params,resources,files,environment,logs,extras,kid)"+
		" VALUES (?,?,?,?,?,?,?,?,?,?,?,?)", BuildsTable))

	if err != nil {
		log.Println(err)
		return err
	}

	defer stmt.Close()

	servers, _ := json.Marshal(dd.Servers)
	params, _ := json.Marshal(dd.Params)
	resources, _ := json.Marshal(dd.Resources)
	logs, _ := json.Marshal(dd.Logs)
	extras, _ := json.Marshal(dd.Extras)
	images, _ := json.Marshal(dd.Images)
	files, _ := json.Marshal(dd.Files)
	environment, err := json.Marshal(dd.Environments)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = stmt.Exec(testnetID, string(servers), dd.Blockchain, dd.Nodes, string(images),
		string(params), string(resources), string(files), string(environment), string(logs), string(extras), dd.kid)

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
