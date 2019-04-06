/*
Handles functions related to the current state of the network
*/
package status

import (
	state "../state"
	"encoding/json"
	"fmt"
	"log"
)

type BuildStatus struct {
	Error    state.CustomError `json:"error"`
	Progress float64           `json:"progress"`
	Stage    string            `json:"stage"`
	Frozen   bool              `json:"frozen"`
}

/*
   Check the current status of the build
*/
func CheckBuildStatus(buildId string) (string, error) {
	bs, err := state.GetBuildStateById(buildId)
	if err != nil {
		log.Println(err)
		return "", err
	}
	if bs.ErrorFree() { //error should be null if there is not an error
		return fmt.Sprintf("{\"progress\":%f,\"error\":null,\"stage\":\"%s\",\"frozen\":%v}", bs.BuildingProgress, bs.BuildStage, bs.Frozen), nil
	}
	//otherwise give the error as an object
	out, _ := json.Marshal(BuildStatus{Progress: bs.BuildingProgress, Error: bs.BuildError, Stage: bs.BuildStage, Frozen: bs.Frozen})
	return string(out), nil
}
