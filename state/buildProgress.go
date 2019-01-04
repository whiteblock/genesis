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
    mutex                       =   &sync.Mutex{}
    errMutex                    =   &sync.Mutex{}
    building            bool    =   false
    progressIncrement   float64 =   0.00

    BuildingProgress    float64 =   0.00
    BuildError          CustomError =   CustomError{What:"",err:nil}
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
    building = false
}

func ReportError(err error){
    errMutex.Lock()
    defer errMutex.Unlock()
    BuildError = CustomError{What:err.Error(),err:err}
    log.Println("An error has been reported :"+err.Error())

}

func ErrorFree() bool {
    errMutex.Lock()
    defer errMutex.Unlock()
    return BuildError.err == nil
}

func GetError() error {
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

