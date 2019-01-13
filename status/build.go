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


func CheckBuildStatus() string {
    if state.ErrorFree() {
        return fmt.Sprintf("{\"progress\":%f,\"error\":null,\"stage\":\"%s\"}",state.BuildingProgress,state.BuildStage)
    }else{
        out,_ := json.Marshal(BuildStatus{ Progress:state.BuildingProgress, Error:state.BuildError,Stage:state.BuildStage })
        return string(out)
    }
}