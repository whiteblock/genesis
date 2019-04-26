/*
Handles functions related to the current state of the network
*/
package status

import (
	"../state"
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
	return bs.Marshal(), nil
}
