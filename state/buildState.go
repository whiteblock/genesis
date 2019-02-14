package state

import(
    "sync"
    "errors"
    "log"
)

/**
 * Packages the build state nicely into an object
 */
type BuildState struct{
    mutex               sync.RWMutex
    errMutex            sync.RWMutex
    stopMux             sync.RWMutex
    building            bool
    progressIncrement   float64
    stopping            bool

    Servers             []int
    BuildingProgress    float64
    BuildError          CustomError
    BuildStage          string
}


func NewBuildState(servers []int) *BuildState {
    out := new(BuildState)
    out.building = false
    out.progressIncrement = 0.00
    out.stopping = false

    out.Servers = servers
    out.BuildingProgress = 0.00
    out.BuildError = CustomError{What:"",err:nil}
    out.BuildStage = ""    
    return out 
}


/*
    DoneBuilding signals that the building process has finished and releases the
    build lock.
 */
func (this *BuildState) DoneBuilding(){
    this.mutex.Lock()
    defer this.mutex.Unlock()
    this.BuildingProgress = 100.00
    this.BuildStage = "Finished"
    this.building = false
    this.stopping = false
}

func (this *BuildState) Done() bool {
    return !this.building
}

/*
    ReportError stores the given error to be passed onto any
    who query the build status. 
 */
func (this *BuildState) ReportError(err error){
    this.errMutex.Lock()
    defer this.errMutex.Unlock()
    this.BuildError = CustomError{What:err.Error(),err:err}
    log.Println("An error has been reported :"+err.Error())
}

/*
    Stop checks if the stop signal has been sent. If this returns true,
    a building process should return. The ssh client checks this for you. 
 */
func (this *BuildState) Stop() bool {
    this.stopMux.RLock()
    defer this.stopMux.RUnlock()
    return this.stopping
}

/*
    SignalStop flags that the current build should be stopped, if there is
    a current build. Returns an error if there is no build in progress
 */
func (this *BuildState) SignalStop() error {
    this.stopMux.Lock()
    defer this.stopMux.Unlock()
    this.mutex.RLock()
    defer this.mutex.RUnlock()
    
    if this.building {
        this.ReportError(errors.New("Build stopped by user"))
        this.stopping = true
        return nil
    }
    return errors.New("No build in progress")
}

/*
    ErrorFree checks that there has not been an error reported with
    ReportError
 */
func (this *BuildState) ErrorFree() bool {
    this.errMutex.RLock()
    defer this.errMutex.RUnlock()
    return this.BuildError.err == nil
}

/*
    GetError gets the currently stored error
*/
func (this *BuildState) GetError() error {
    this.errMutex.RLock()
    defer this.errMutex.RUnlock()
    return this.BuildError.err
}

/*
    SetDeploySteps sets the number of steps in the deployment process.
    Should be given a number equivalent to the number of times 
    IncrementDeployProgress will be called.
*/
func (this *BuildState) SetDeploySteps(steps int){
    this.progressIncrement = 25.00 / float64(steps)
}

/*
    IncrementDeployProgress increments the deploy process by one step.
*/
func (this *BuildState) IncrementDeployProgress(){
    this.BuildingProgress += this.progressIncrement
}

/*
    FinishDeploy signals that the deployment process has finished and the 
    blockchain specific process will begin.
 */
func (this *BuildState) FinishDeploy(){
    this.BuildingProgress = 25.00
}

/*
    SetBuildSteps sets the number of steps in the blockchain specific 
    build process. Must be equivalent to the number of times IncrementBuildProgress()
    will be called. 
 */
func (this *BuildState) SetBuildSteps(steps int){
    this.progressIncrement = 75.00 / float64(steps)
}

/*
    IncrementBuildProgress increments the build progress by one step.
 */
func (this *BuildState) IncrementBuildProgress(){
    this.BuildingProgress += this.progressIncrement
}

/*
    SetBuildStage updates the text which will be displayed along with the
    build progress percentage when the status of the build is queried.
 */
func (this *BuildState) SetBuildStage(stage string){
    this.BuildStage = stage
}

