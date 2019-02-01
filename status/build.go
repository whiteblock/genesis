/*
Handles functions related to the current state of the network
 */
package status

import(
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
func CheckBuildStatus() string {
    if state.ErrorFree() {
        return fmt.Sprintf("{\"progress\":%f,\"error\":null,\"stage\":\"%s\"}",state.BuildingProgress,state.BuildStage)
    }else{
        out,_ := json.Marshal(BuildStatus{ Progress:state.BuildingProgress, Error:state.BuildError,Stage:state.BuildStage })
        return string(out)
    }
}