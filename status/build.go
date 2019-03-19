/*
Handles functions related to the current state of the network
 */
package status

import(
    "log"
    "fmt"
    "encoding/json"
    state "../state"
)


type BuildStatus struct {
    Error       state.CustomError   `json:"error"`
    Progress    float64             `json:"progress"`
    Stage       string              `json:"stage"`
}


/*
    Check the current status of the build
 */
func CheckBuildStatus(buildId string) (string,error) {
    bs,err := state.GetBuildStateById(buildId)
    if err != nil {
        log.Println(err)
        return "",err
    }
    if bs.ErrorFree() {
        return fmt.Sprintf("{\"progress\":%f,\"error\":null,\"stage\":\"%s\"}",bs.BuildingProgress,bs.BuildStage),nil
    }

    out,_ := json.Marshal(BuildStatus{ Progress:bs.BuildingProgress, Error:bs.BuildError,Stage:bs.BuildStage })
    return string(out),nil
}