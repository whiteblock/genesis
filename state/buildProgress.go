package state


import(
    "sync"
    "errors"
    "log"
)
/**
 * Packages the global state nicely into an "object"
 */
var (
    mutex                           =   &sync.RWMutex{}
    errMutex                        =   &sync.RWMutex{}
    stopMux                         =   &sync.RWMutex{}

    building            bool        =   false
    progressIncrement   float64     =   0.00
    stopping            bool        =   false

    BuildingProgress    float64     =   0.00
    BuildError          CustomError =   CustomError{What:"",err:nil}
    BuildStage          string      =   ""     
)

/*
    AcquireBuilding acquires a build lock. Any function which modifies 
    the nodes in a testnet should only do so after calling this function 
    and ensuring that the returned value is nil
 */
func AcquireBuilding() error {
    mutex.Lock()
    defer mutex.Unlock()

    if building {
        return errors.New("Error: Build in progress")
    }

    building = true
    BuildingProgress = 0.00
    BuildError = CustomError{What:"",err:nil}
    return nil
}

/*
    DoneBuilding signals that the building process has finished and releases the
    build lock.
 */
func DoneBuilding(){
    mutex.Lock()
    defer mutex.Unlock()
    BuildingProgress = 100.00
    BuildStage = "Finished"
    building = false
    stopping = false
}

/*
    ReportError stores the given error to be passed onto any
    who query the build status. 
 */
func ReportError(err error){
    errMutex.Lock()
    defer errMutex.Unlock()
    BuildError = CustomError{What:err.Error(),err:err}
    log.Println("An error has been reported :"+err.Error())
}

/*
    Stop checks if the stop signal has been sent. If this returns true,
    a building process should return. The ssh client checks this for you. 
 */
func Stop() bool {
    stopMux.RLock()
    defer stopMux.RUnlock()
    return stopping
}

/*
    SignalStop flags that the current build should be stopped, if there is
    a current build. Returns an error if there is no build in progress
 */
func SignalStop() error {
    stopMux.Lock()
    defer stopMux.Unlock()
    mutex.RLock()
    defer mutex.RUnlock()
    
    if building{
        ReportError(errors.New("Build stopped by user"))
        stopping = true
        return nil
    }
    return errors.New("No build in progress")
}

/*
    ErrorFree checks that there has not been an error reported with
    ReportError
 */
func ErrorFree() bool {
    errMutex.RLock()
    defer errMutex.RUnlock()
    return BuildError.err == nil
}

/*
    GetError gets the currently stored error
*/
func GetError() error {
    errMutex.RLock()
    defer errMutex.RUnlock()
    return BuildError.err
}

/*
    SetDeploySteps sets the number of steps in the deployment process.
    Should be given a number equivalent to the number of times 
    IncrementDeployProgress will be called.
*/
func SetDeploySteps(steps int){
    progressIncrement = 25.00 / float64(steps)
}

/*
    IncrementDeployProgress increments the deploy process by one step.
*/
func IncrementDeployProgress(){
    BuildingProgress += progressIncrement
}

/*
    FinishDeploy signals that the deployment process has finished and the 
    blockchain specific process will begin.
 */
func FinishDeploy(){
    BuildingProgress = 25.00
}

/*
    SetBuildSteps sets the number of steps in the blockchain specific 
    build process. Must be equivalent to the number of times IncrementBuildProgress()
    will be called. 
 */
func SetBuildSteps(steps int){
    progressIncrement = 75.00 / float64(steps)
}

/*
    IncrementBuildProgress increments the build progress by one step.
 */
func IncrementBuildProgress(){
    BuildingProgress += progressIncrement
}

/*
    SetBuildStage updates the text which will be displayed along with the
    build progress percentage when the status of the build is queried.
 */
func SetBuildStage(stage string){
    BuildStage = stage
}

