package state

import(
	"sync"
	"errors"
	"log"
)
/**
 * Packages the global state nicely into an "object"
 */
type CustomError struct{
	What	string		`json:"what"`
	err		error
}

var (
	mutex 						=	&sync.Mutex{}
	errMutex					=	&sync.Mutex{}
	building			bool	=	false
	progressIncrement	float64	=	0.00
)

var (
	BuildingProgress	float64	=	0.00
	BuildError			CustomError	=	CustomError{What:"",err:nil}
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

func DoneBuilding(){
	mutex.Lock()
	defer mutex.Unlock()
	BuildingProgress = 100.00
	building = false
}

func SetBuildSteps(steps int){
	progressIncrement = 100.00 / float64(steps)
}

func IncrementBuildProgress(){
	BuildingProgress += progressIncrement
}