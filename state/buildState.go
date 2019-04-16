/*
   Handles global state that can be changed
*/
package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

//This code is full of potential race conditons but these race conditons are extremely rare
/*
   CustomError is a custom wrapper for a go error, which
   has What containing error.Error()
*/
type CustomError struct {
	What string `json:"what"`
	err  error
}

/*
   Packages the build state nicely into an object
*/
type BuildState struct {
	errMutex  *sync.RWMutex
	extraMux  *sync.RWMutex
	freeze    *sync.RWMutex
	mutex     *sync.RWMutex
	stopMux   *sync.RWMutex
	freezeMux *sync.RWMutex

	building bool
	Frozen   bool
	stopping bool

	breakpoints       []float64 //must be in ascending order
	progressIncrement float64
	externExtras      map[string]interface{} //will be exported
	extras            map[string]interface{}
	files             []string
	defers            []func() //Array of functions to run at the end of the build
	asyncWaiter       sync.WaitGroup

	Servers          []int
	BuildId          string
	BuildingProgress float64
	BuildError       CustomError
	BuildStage       string
}

func NewBuildState(servers []int, buildId string) *BuildState {
	out := new(BuildState)

	out.errMutex = &sync.RWMutex{}
	out.extraMux = &sync.RWMutex{}
	out.freeze = &sync.RWMutex{}
	out.mutex = &sync.RWMutex{}
	out.stopMux = &sync.RWMutex{}
	out.freezeMux = &sync.RWMutex{}

	out.building = true
	out.Frozen = false
	out.stopping = false

	out.breakpoints = []float64{}
	out.progressIncrement = 0.0
	out.externExtras = map[string]interface{}{}
	out.extras = map[string]interface{}{}
	out.files = []string{}
	out.defers = []func(){}

	out.Servers = servers
	out.BuildId = buildId
	out.BuildingProgress = 0.00
	out.BuildError = CustomError{What: "", err: nil}
	out.BuildStage = ""

	err := os.MkdirAll("/tmp/"+buildId, 0755)
	if err != nil {
		panic(err) //Fatal error
	}

	return out
}

/*
Set a function to be executed at some point during the build.
All these functions must complete before the build is considered finished.
*/
func (this *BuildState) Async(fn func()) {
	this.asyncWaiter.Add(1)
	go func() {
		defer this.asyncWaiter.Done()
		fn()
	}()
}

func (this *BuildState) Freeze() error {
	log.Println("Freezing the build")
	this.mutex.Lock()
	if this.Frozen {
		this.mutex.Unlock()
		return fmt.Errorf("Already frozen")
	}
	if this.stopping {
		this.mutex.Unlock()
		return fmt.Errorf("The build is terminating")
	}

	this.Frozen = true
	this.mutex.Unlock()

	this.freeze.Lock()

	return nil
}

func (this *BuildState) Unfreeze() error {
	log.Println("Thawing the build")
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if !this.Frozen {
		return fmt.Errorf("Not currently frozen")
	}
	this.freeze.Unlock()
	this.Frozen = false
	return nil
}

func (this *BuildState) AddFreezePoint(freezePoint float64) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	i := 0
	for ; i < len(this.breakpoints); i++ { //find insertion index
		if this.breakpoints[i] > freezePoint {
			break
		} else if this.breakpoints[i] == freezePoint {
			return //no duplicates
		}
	}
	this.breakpoints = append(append(this.breakpoints[:i], freezePoint), this.breakpoints[i:]...)
}

/*
   DoneBuilding signals that the building process has finished and releases the
   build lock.
*/
func (this *BuildState) DoneBuilding() {
	//TODO add file cleanup
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.BuildingProgress = 100.00
	this.BuildStage = "Finished"
	this.building = false
	this.stopping = false
	this.asyncWaiter.Wait() //Wait for the async calls to complete
	os.RemoveAll("/tmp/" + this.BuildId)

	for _, fn := range this.defers {
		go fn() //No need to wait to confirm completion
	}
}

func (this *BuildState) Done() bool {
	return !this.building
}

/*
   ReportError stores the given error to be passed onto any
   who query the build status.
*/
func (this *BuildState) ReportError(err error) {
	this.errMutex.Lock()
	defer this.errMutex.Unlock()
	this.BuildError = CustomError{What: err.Error(), err: err}
	log.Println("An error has been reported :" + err.Error())
}

/*
   Stop checks if the stop signal has been sent. If this returns true,
   a building process should return. The ssh client checks this for you.
*/
func (this *BuildState) Stop() bool {
	if this == nil { //golang allows for nil.Stop() to be a thing...
		return false
	}

	this.freeze.RLock()
	this.freeze.RUnlock()

	if len(this.breakpoints) > 0 { //Don't take the lock overhead if there aren't any breakpoints
		this.freezeMux.Lock()
		if this.breakpoints[0] >= this.BuildingProgress {
			if len(this.breakpoints) > 1 {
				this.breakpoints = this.breakpoints[1:]
			} else {
				this.breakpoints = []float64{}
			}
			this.Freeze()
			this.freeze.RLock()
		}
		this.freezeMux.Unlock()
	}

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
	this.Unfreeze()
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	if this.building {
		this.ReportError(fmt.Errorf("Build stopped by user"))
		this.stopping = true
		return nil
	}
	return fmt.Errorf("No build in progress")
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
	Insert a value into the state store, currently only supports string
	and []string on the other side
*/
func (this *BuildState) SetExt(key string, value interface{}) error {
	switch value.(type) {
	case string:
	case []string:
	default:
		return fmt.Errorf("Unsupported type for value")
	}
	this.extraMux.Lock()
	defer this.extraMux.Unlock()
	this.externExtras[key] = value
	return nil
}

func (this *BuildState) GetExt(key string) (interface{}, bool) {
	this.extraMux.RLock()
	defer this.extraMux.RUnlock()
	res, ok := this.externExtras[key]
	return res, ok
}

func (this *BuildState) GetExtExtras() ([]byte, error) {
	this.extraMux.RLock()
	defer this.extraMux.RUnlock()
	return json.Marshal(this.externExtras)
}

func (this *BuildState) Set(key string, value interface{}) {
	this.extraMux.Lock()
	defer this.extraMux.Unlock()
	this.extras[key] = value
}

func (this *BuildState) Get(key string) (interface{}, bool) {
	this.extraMux.RLock()
	defer this.extraMux.RUnlock()
	out, ok := this.extras[key]
	return out, ok
}

func (this *BuildState) GetExtras() map[string]interface{} {
	return this.extras
}

/*
	Write writes data to a file, creating it if it doesn't exist,
   	deleting and recreating it if it does.
*/
func (this *BuildState) Write(file string, data string) error {
	this.mutex.RLock()
	this.files = append(this.files, file)
	this.mutex.RUnlock()
	return ioutil.WriteFile("/tmp/"+this.BuildId+"/"+file, []byte(data), 0664)
}

/*
	Add a function to be executed asynchronously after the build is completed.
*/
func (this *BuildState) Defer(fn func()) {
	this.extraMux.Lock()
	this.defers = append(this.defers, fn)
	defer this.extraMux.Unlock()
}

/*
   SetDeploySteps sets the number of steps in the deployment process.
   Should be given a number equivalent to the number of times
   IncrementDeployProgress will be called.
*/
func (this *BuildState) SetDeploySteps(steps int) {
	this.progressIncrement = 25.00 / float64(steps)
}

/*
   IncrementDeployProgress increments the deploy process by one step.
*/
func (this *BuildState) IncrementDeployProgress() {
	this.BuildingProgress += this.progressIncrement
}

/*
   FinishDeploy signals that the deployment process has finished and the
   blockchain specific process will begin.
*/
func (this *BuildState) FinishDeploy() {
	this.BuildingProgress = 25.00
}

/*
   SetBuildSteps sets the number of steps in the blockchain specific
   build process. Must be equivalent to the number of times IncrementBuildProgress()
   will be called.
*/
func (this *BuildState) SetBuildSteps(steps int) {
	this.progressIncrement = 75.00 / float64(steps+1)
}

/*
   IncrementBuildProgress increments the build progress by one step.
*/
func (this *BuildState) IncrementBuildProgress() {
	this.BuildingProgress += this.progressIncrement
	if this.BuildingProgress >= 100.00 {
		this.BuildingProgress = 99.99
	}
}

/*
   SetBuildStage updates the text which will be displayed along with the
   build progress percentage when the status of the build is queried.
*/
func (this *BuildState) SetBuildStage(stage string) {
	this.BuildStage = stage
}
