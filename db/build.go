package db

import (
	"../util"
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
	Resources    []util.Resources       `json:"resources"`
	Environments []map[string]string    `json:"environments"`
	Files        []map[string]string    `json:"files"`
	Logs         []map[string]string    `json:"logs"`
	Extras       map[string]interface{} `json:"extras"`
	jwt          string
	kid          string
}

func (this *DeploymentDetails) SetJwt(jwt string) error {
	this.jwt = jwt
	kid, err := util.GetKidFromJwt(this.GetJwt())

	this.kid = kid
	return err
}

func (this DeploymentDetails) GetJwt() string {
	return this.jwt
}

func (this DeploymentDetails) GetKid() string {
	return this.kid
}

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

		err = rows.Scan(&servers, &build.Blockchain, &build.Nodes, &images, &params, &resources, &environment, &logs, &extras, &build.kid)
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
	return QueryBuilds(fmt.Sprintf("SELECT servers,blockchain,nodes,image,params,resources,environment,logs,extras,kid FROM %s", BuildsTable))
}

/*
GetBuildByTestnet gets the build paramters based off testnet id
*/
func GetBuildByTestnet(id string) (DeploymentDetails, error) {

	details, err := QueryBuilds(fmt.Sprintf("SELECT servers,blockchain,nodes,image,params,resources,environment,logs,extras,kid FROM %s WHERE testnet = \"%s\"", BuildsTable, id))
	if err != nil {
		log.Println(err)
		return DeploymentDetails{}, err
	}
	if len(details) == 0 {
		return DeploymentDetails{}, fmt.Errorf("no results found")
	}
	return details[0], nil
}

/*
GetBuildByTestnet gets the build paramters based off testnet id
*/
func GetLastBuildByKid(kid string) (DeploymentDetails, error) {

	details, err := QueryBuilds(fmt.Sprintf(
		"SELECT servers,blockchain,nodes,image,params,resources,environment,logs,extras,kid FROM %s"+
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

/*
InsertBuild inserts a build
*/
func InsertBuild(dd DeploymentDetails, testnetID string) error {

	tx, err := db.Begin()

	if err != nil {
		log.Println(err)
		return err
	}

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (testnet,servers,blockchain,nodes,image,params,resources,environment,logs,extras,kid) VALUES (?,?,?,?,?,?,?,?,?,?,?)", BuildsTable))

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
	environment, err := json.Marshal(dd.Environments)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = stmt.Exec(testnetID, string(servers), dd.Blockchain, dd.Nodes, string(images),
		string(params), string(resources), string(environment), string(logs), string(extras), dd.kid)

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
