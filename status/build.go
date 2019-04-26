// Package status handles functions related to the current state of the network
package status

import (
	"../state"
	"log"
)

// CheckBuildStatus checks the current status of the build relating to the
// given build id
func CheckBuildStatus(buildID string) (string, error) {
	bs, err := state.GetBuildStateByID(buildID)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return bs.Marshal(), nil
}
