package state

import(
    //"sync"

)

var bs = NewBuildState()
var buildStates = []*BuildState{}

func GetBuildState(i int) *BuildState {
    return bs
}
/*
    AcquireBuilding acquires a build lock. Any function which modifies 
    the nodes in a testnet should only do so after calling this function 
    and ensuring that the returned value is nil
 */
func AcquireBuilding() error {
    return bs.AcquireBuilding()
}

/*
    DoneBuilding signals that the building process has finished and releases the
    build lock.
 */
func DoneBuilding(){
    bs.DoneBuilding()
}

/*
    ReportError stores the given error to be passed onto any
    who query the build status. 
 */
func ReportError(err error){
    bs.ReportError(err)
}

/*
    Stop checks if the stop signal has been sent. If this returns true,
    a building process should return. The ssh client checks this for you. 
 */
func Stop() bool {
    return bs.Stop()
}

/*
    SignalStop flags that the current build should be stopped, if there is
    a current build. Returns an error if there is no build in progress
 */
func SignalStop() error {
    return bs.SignalStop()
}

/*
    ErrorFree checks that there has not been an error reported with
    ReportError
 */
func ErrorFree() bool {
    return bs.ErrorFree()
}

/*
    GetError gets the currently stored error
*/
func GetError() error {
    return bs.GetError()
}

/*
    SetDeploySteps sets the number of steps in the deployment process.
    Should be given a number equivalent to the number of times 
    IncrementDeployProgress will be called.
*/
func SetDeploySteps(steps int){
    bs.SetDeploySteps(steps)
}

/*
    IncrementDeployProgress increments the deploy process by one step.
*/
func IncrementDeployProgress(){
    bs.IncrementDeployProgress()
}

/*
    FinishDeploy signals that the deployment process has finished and the 
    blockchain specific process will begin.
 */
func FinishDeploy(){
    bs.FinishDeploy()
}

/*
    SetBuildSteps sets the number of steps in the blockchain specific 
    build process. Must be equivalent to the number of times IncrementBuildProgress()
    will be called. 
 */
func SetBuildSteps(steps int){
    bs.SetBuildSteps(steps)
}

/*
    IncrementBuildProgress increments the build progress by one step.
 */
func IncrementBuildProgress(){
    bs.IncrementBuildProgress()
}

/*
    SetBuildStage updates the text which will be displayed along with the
    build progress percentage when the status of the build is queried.
 */
func SetBuildStage(stage string){
    bs.SetBuildStage(stage)
}

