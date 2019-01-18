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

func DoneBuilding(){
    mutex.Lock()
    defer mutex.Unlock()
    BuildingProgress = 100.00
    BuildStage = "Finished"
    building = false
    stopping = false
}

func ReportError(err error){
    errMutex.Lock()
    defer errMutex.Unlock()
    BuildError = CustomError{What:err.Error(),err:err}
    log.Println("An error has been reported :"+err.Error())
}

func Stop() bool {
    stopMux.RLock()
    defer stopMux.RUnlock()
    return stopping
}

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

func ErrorFree() bool {
    errMutex.RLock()
    defer errMutex.RUnlock()
    return BuildError.err == nil
}

func GetError() error {
    errMutex.RLock()
    defer errMutex.RUnlock()
    return BuildError.err
}


func SetDeploySteps(steps int){
    progressIncrement = 25.00 / float64(steps)
}

func IncrementDeployProgress(){
    BuildingProgress += progressIncrement
}

func FinishDeploy(){
    BuildingProgress = 25.00
}

func SetBuildSteps(steps int){
    progressIncrement = 75.00 / float64(steps)
}

func IncrementBuildProgress(){
    BuildingProgress += progressIncrement
}

func SetBuildStage(stage string){
    BuildStage = stage
}

