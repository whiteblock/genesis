package state

import(
	"sync"
	"errors"
)
/**
 * Packages the global state nicely into an "object"
 */

var (
	mutex 						=	&sync.Mutex{}
	building			bool	=	false
	progressIncrement	float64	=	0.00
)

var (
	BuildingProgress	float64	=	0.00
	BuildError			error	=	nil
)

func AcquireBuilding() error {
	mutex.Lock()
	defer mutex.Unlock()

	if building {
		return errors.New("Error: Build in progress")
	}

	building = true
	BuildingProgress = 0.00
	BuildError = nil
	return nil
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