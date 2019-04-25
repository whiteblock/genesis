/*
   Handles global state that can be changed
*/
package state

import (
	db "../db"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"sync/atomic"
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

	breakpoints  []float64              //must be in ascending order
	ExternExtras map[string]interface{} //will be exported
	Extras       map[string]interface{}
	files        []string
	defers       []func() //Array of functions to run at the end of the build
	asyncWaiter  sync.WaitGroup

	Servers []int
	BuildId string

	BuildError CustomError
	BuildStage string

	DeployProgress uint64
	DeployTotal    uint64

	BuildProgress uint64
	BuildTotal    uint64
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
	out.ExternExtras = map[string]interface{}{}
	out.Extras = map[string]interface{}{}
	out.files = []string{}
	out.defers = []func(){}

	out.Servers = servers
	out.BuildId = buildId
	out.BuildError = CustomError{What: "", err: nil}
	out.BuildStage = ""

	out.DeployProgress = 0
	out.DeployTotal = 0
	out.BuildProgress = 0
	out.BuildTotal = 1

	err := os.MkdirAll("/tmp/"+buildId, 0755)
	if err != nil {
		panic(err) //Fatal error
	}

	return out
}

func RestoreBuildState(buildID string) (*BuildState, error) {
	out := new(BuildState)
	err := db.GetMetaP(buildID, out)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	out.errMutex = &sync.RWMutex{}
	out.extraMux = &sync.RWMutex{}
	out.freeze = &sync.RWMutex{}
	out.mutex = &sync.RWMutex{}
	out.stopMux = &sync.RWMutex{}
	out.freezeMux = &sync.RWMutex{}

	out.Reset()
	return out, nil
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
	atomic.StoreUint64(&this.BuildProgress, this.BuildTotal)
	this.BuildStage = "Finished"
	this.building = false
	this.stopping = false
	this.asyncWaiter.Wait() //Wait for the async calls to complete
	os.RemoveAll("/tmp/" + this.BuildId)
	if this.ErrorFree() {
		err := this.Store()
		if err != nil {
			log.Println(err)
		}
	}
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
		if this.breakpoints[0] >= this.GetProgress() {
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
	this.ExternExtras[key] = value
	return nil
}

func (this *BuildState) GetExt(key string) (interface{}, bool) {
	this.extraMux.RLock()
	defer this.extraMux.RUnlock()
	res, ok := this.ExternExtras[key]
	return res, ok
}

func (this *BuildState) GetExtExtras() ([]byte, error) {
	this.extraMux.RLock()
	defer this.extraMux.RUnlock()
	return json.Marshal(this.ExternExtras)
}

func (this *BuildState) Set(key string, value interface{}) {
	this.extraMux.Lock()
	defer this.extraMux.Unlock()
	this.Extras[key] = value
}

func (this *BuildState) Get(key string) (interface{}, bool) {
	this.extraMux.RLock()
	defer this.extraMux.RUnlock()
	out, ok := this.Extras[key]
	return out, ok
}

func (this *BuildState) GetP(key string, out interface{}) bool {
	tmp, ok := this.Get(key)
	if !ok {
		return false
	}
	tmpBytes, err := json.Marshal(tmp)
	if err != nil {
		log.Println(err)
		return false
	}
	err = json.Unmarshal(tmpBytes, out)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (this *BuildState) GetExtras() map[string]interface{} {
	return this.Extras
}

/*
	Write writes data to a file, creating it if it doesn't exist,
   	deleting and recreating it if it does.
*/
func (this *BuildState) Write(file string, data string) error {
	this.mutex.Lock()
	this.files = append(this.files, file)
	this.mutex.Unlock()
	err := ioutil.WriteFile("/tmp/"+this.BuildId+"/"+file, []byte(data), 0664)

	return err
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
	atomic.StoreUint64(&this.DeployTotal, uint64(steps))
}

/*
   IncrementDeployProgress increments the deploy process by one step.
*/
func (this *BuildState) IncrementDeployProgress() {
	atomic.AddUint64(&this.DeployProgress, 1)
}

/*
   FinishDeploy signals that the deployment process has finished and the
   blockchain specific process will begin.
*/
func (this *BuildState) FinishDeploy() {
	atomic.StoreUint64(&this.DeployProgress, atomic.LoadUint64(&this.DeployTotal))
}

/*
   SetBuildSteps sets the number of steps in the blockchain specific
   build process. Must be equivalent to the number of times IncrementBuildProgress()
   will be called.
*/
func (this *BuildState) SetBuildSteps(steps int) {
	atomic.StoreUint64(&this.BuildTotal, uint64(steps+1)) //stay one step ahead to prevent early termination
}

/*
   IncrementBuildProgress increments the build progress by one step.
*/
func (this *BuildState) IncrementBuildProgress() {
	if atomic.LoadUint64(&this.BuildProgress) < atomic.LoadUint64(&this.BuildTotal) {
		atomic.AddUint64(&this.BuildProgress, 1)
	}
}

func (this *BuildState) GetProgress() float64 {
	dp := atomic.LoadUint64(&this.DeployProgress)
	dt := atomic.LoadUint64(&this.DeployTotal)
	bp := atomic.LoadUint64(&this.BuildProgress)
	bt := atomic.LoadUint64(&this.BuildTotal)

	var out float64 = 0
	if dp == 0 || dt == 0 {
		return out
	}
	if dp == dt {
		out = 25.0
	} else {
		return float64(dp) / float64(dt) * 25.0
	}
	if bt == 0 {
		return out
	}
	return out + (float64(bp)/float64(bt))*75.0
}

/*
   SetBuildStage updates the text which will be displayed along with the
   build progress percentage when the status of the build is queried.
*/
func (this *BuildState) SetBuildStage(stage string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.BuildStage = stage

}

func (this *BuildState) Reset() {

	this.building = true
	this.Frozen = false
	this.stopping = false

	this.breakpoints = []float64{}

	this.files = []string{}
	this.defers = []func(){}

	this.BuildError = CustomError{What: "", err: nil}
	this.BuildStage = ""

	this.DeployProgress = 0
	this.DeployTotal = 0
	this.BuildProgress = 0
	this.BuildTotal = 0

	err := os.MkdirAll("/tmp/"+this.BuildId, 0755)
	if err != nil {
		panic(err) //Fatal error
	}
	fmt.Println("BUILD has been reset!")
}

func (this *BuildState) Marshal() string {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	if this.ErrorFree() { //error should be null if there is not an error
		return fmt.Sprintf("{\"progress\":%f,\"error\":null,\"stage\":\"%s\",\"frozen\":%v}", this.GetProgress(), this.BuildStage, this.Frozen)
	}
	//otherwise give the error as an object
	out, _ := json.Marshal(
		map[string]interface{}{"progress": this.GetProgress(), "error": this.BuildError, "stage": this.BuildStage, "frozen": this.Frozen})
	return string(out)
}

func (this *BuildState) Store() error {
	return db.SetMeta(this.BuildId, *this)
}

func (this *BuildState) Destroy() error {
	return db.DeleteMeta(this.BuildId)
}
