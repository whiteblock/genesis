package status

import(
    "fmt"
    "encoding/json"
    state "../state"
)


type BuildStatus struct {
    Error       state.CustomError   `json:"error"`
    Progress    float64             `json:"progress"`
}


func CheckBuildStatus() string {
    if state.ErrorFree() {
        return fmt.Sprintf("{\"progress\":%f,\"error\":null}",state.BuildingProgress)
    }else{
        out,_ := json.Marshal(BuildStatus{ Progress:state.BuildingProgress, Error:state.BuildError })
        return string(out)
    }
}