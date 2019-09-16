/*
	Copyright 2019 whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Genesis is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

// Package state provides utilities to manage state
package state

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/utility/utils"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
)

//This code is full of potential race conditions but these race conditons are extremely rare

// CustomError is a custom wrapper for a go error, which
// has What containing error.Error()
type CustomError struct {
	What string `json:"what"`
	err  error
}

// BuildState packages the build state nicely into an object
type BuildState struct {
	errMutex *sync.RWMutex
	extraMux *sync.RWMutex
	freeze   *sync.RWMutex
	mutex    *sync.RWMutex

	building int32 //0 or 1. Made into atomic to reduce mutex hell
	frozen   int32 //0 or 1. Made into atomic to reduce mutex hell
	stopping int32 //0 or 1. Made into atomic to reduce mutex hell

	breakpoints       []float64              //must be in ascending order
	ExternExtras      map[string]interface{} //will be exported
	Extras            map[string]interface{}
	files             []string
	defers            []func() //Array of functions to run at the end of the build
	errorCleanupFuncs []func()
	//functions which will be called when the build state is destroyed.
	//No timing guarentees beyond the functions happening at some point after Destroy is called.
	onDestroyFuncs []func()
	asyncWaiter    *sync.WaitGroup

	Servers []int
	BuildID string

	BuildError CustomError
	BuildStage string

	DeployProgress uint64
	DeployTotal    uint64

	BuildProgress uint64
	BuildTotal    uint64

	SideCars        uint64 //The number of side cars
	SideCarProgress uint64
	SideCarTotal    uint64
}

//NewBuildState creates a new build state for the given servers with the given buildID
func NewBuildState(servers []int, buildID string) *BuildState {
	out := new(BuildState)

	out.errMutex = &sync.RWMutex{}
	out.extraMux = &sync.RWMutex{}
	out.freeze = &sync.RWMutex{}
	out.mutex = &sync.RWMutex{}
	out.asyncWaiter = &sync.WaitGroup{}

	out.building = 1
	out.frozen = 0
	out.stopping = 0

	out.breakpoints = []float64{}
	out.ExternExtras = map[string]interface{}{}
	out.Extras = map[string]interface{}{}
	out.files = []string{}
	out.defers = []func(){}
	out.errorCleanupFuncs = []func(){}

	out.Servers = servers
	out.BuildID = buildID
	out.BuildError = CustomError{What: "", err: nil}
	out.BuildStage = ""

	out.DeployProgress = 0
	out.DeployTotal = 0
	out.BuildProgress = 0
	out.BuildTotal = 1
	out.SideCarProgress = 0
	out.SideCarTotal = 1

	err := os.MkdirAll("/tmp/"+buildID, 0755)
	if err != nil {
		log.WithFields(log.Fields{"build": out.BuildID, "error": err}).Panic("couldn't create the tmp folder")
	}

	return out
}

// RestoreBuildState creates a BuildState from the previous BuildState
// with the same BuildID. If one does not exist, it returns an error.
func RestoreBuildState(buildID string) (*BuildState, error) {
	out := new(BuildState)
	err := db.GetMetaP(buildID, out)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "id": buildID}).Error("couldn't restore build state")
		return nil, err
	}
	out.errMutex = &sync.RWMutex{}
	out.extraMux = &sync.RWMutex{}
	out.freeze = &sync.RWMutex{}
	out.mutex = &sync.RWMutex{}
	out.asyncWaiter = &sync.WaitGroup{}

	out.Reset()
	return out, nil
}

// Async Set a function to be executed at some point during the build.
// All these functions must complete before the build is considered finished.
func (bs *BuildState) Async(fn func()) {
	bs.asyncWaiter.Add(1)
	go func() {
		defer bs.asyncWaiter.Done()
		fn()
	}()
}

// Freeze freezes the build
func (bs *BuildState) Freeze() error {
	log.WithFields(log.Fields{"build": bs.BuildID}).Info("freezing the build")

	if atomic.LoadInt32(&bs.frozen) != 0 {
		return fmt.Errorf("already frozen")
	}
	if atomic.LoadInt32(&bs.stopping) != 0 {
		return fmt.Errorf("build terminating")
	}

	atomic.StoreInt32(&bs.frozen, 1)

	bs.freeze.Lock()

	return nil
}

// Unfreeze unfreezes the build
func (bs *BuildState) Unfreeze() error {
	log.WithFields(log.Fields{"build": bs.BuildID}).Info("unfreezing the build")

	if atomic.LoadInt32(&bs.frozen) == 0 {
		return fmt.Errorf("not currently frozen")
	}
	bs.freeze.Unlock()
	atomic.StoreInt32(&bs.frozen, 0)
	return nil
}

// IsFrozen check if the build is currently frozen
func (bs *BuildState) IsFrozen() bool {
	return atomic.LoadInt32(&bs.frozen) != 0
}

// AddFreezePoint adds a point at which the build will freeze in the future
func (bs *BuildState) AddFreezePoint(freezePoint float64) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	i := 0
	for ; i < len(bs.breakpoints); i++ { //find insertion index
		if bs.breakpoints[i] > freezePoint {
			break
		} else if bs.breakpoints[i] == freezePoint {
			return //no duplicates
		}
	}
	bs.breakpoints = append(append(bs.breakpoints[:i], freezePoint), bs.breakpoints[i:]...)
}

// DoneBuilding signals that the building process has finished and releases the
// build lock.
func (bs *BuildState) DoneBuilding() {

	if bs.ErrorFree() {
		err := bs.Store()
		if err != nil {
			log.WithFields(log.Fields{"build": bs.BuildID}).Error("couldn't store the build")
		}
	} else {
		log.WithFields(log.Fields{"build": bs.BuildID}).Debug("running the on error function calls.")
		bs.extraMux.RLock()
		wg := sync.WaitGroup{} //Wait for completion, to prevent a potential race
		for i := range bs.errorCleanupFuncs {
			wg.Add(1)
			go func(fn *func()) {
				defer wg.Done()
				(*fn)()
			}(&bs.errorCleanupFuncs[i])
		}
		bs.extraMux.RUnlock()
		wg.Wait()
	}
	atomic.StoreUint64(&bs.DeployProgress, atomic.LoadUint64(&bs.DeployTotal))
	atomic.StoreUint64(&bs.BuildProgress, atomic.LoadUint64(&bs.BuildTotal))
	atomic.StoreUint64(&bs.SideCarProgress, atomic.LoadUint64(&bs.SideCarTotal))

	bs.asyncWaiter.Wait() //Wait for the async calls to complete

	bs.mutex.Lock()
	bs.BuildStage = "Finished"
	bs.mutex.Unlock()
	bs.errorCleanupFuncs = []func(){}
	atomic.StoreInt32(&bs.building, 0)
	atomic.StoreInt32(&bs.stopping, 0)

	log.WithFields(log.Fields{"build": bs.BuildID}).Debug("running the defered functions")
	for _, fn := range bs.defers {
		go fn() //No need to wait to confirm completion
	}
}

// Done checks if the build is done
func (bs *BuildState) Done() bool {
	return atomic.LoadInt32(&bs.building) == 0
}

// ReportError stores the given error to be passed onto any
// who query the build status.
func (bs *BuildState) ReportError(err error) error {
	if err == nil {
		log.Trace("ignoring nil error")
		return nil
	}
	bs.errMutex.Lock()
	defer bs.errMutex.Unlock()
	bs.BuildError = CustomError{What: err.Error(), err: err}

	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	}
	log.WithFields(log.Fields{"build": bs.BuildID, "file": file, "line": line, "error": err}).Error("an error was reported")
	return err
}

// Stop checks if the stop signal has been sent. If bs returns true,
// a building process should return. The ssh client checks bs for you.
func (bs *BuildState) Stop() bool {
	if bs == nil {
		return false //When the buildstate is nil, don't stop
	}
	bs.freeze.RLock() //Catch on freeze
	bs.freeze.RUnlock()

	if len(bs.breakpoints) > 0 { //Don't take the lock overhead if there aren't any breakpoints
		bs.mutex.Lock()
		if bs.breakpoints[0] >= bs.GetProgress() {
			if len(bs.breakpoints) > 1 {
				bs.breakpoints = bs.breakpoints[1:]
			} else {
				bs.breakpoints = []float64{}
			}
			bs.Freeze()
			bs.freeze.RLock()
		}
		bs.mutex.Unlock()
	}

	return atomic.LoadInt32(&bs.stopping) != 0
}

// SignalStop flags that the current build should be stopped, if there is
// a current build. Returns an error if there is no build in progress
func (bs *BuildState) SignalStop() error {

	bs.Unfreeze() //Unfeeze in order to actually stop the build via error propagation

	if atomic.LoadInt32(&bs.building) != 0 {
		bs.ReportError(fmt.Errorf("build stopped by user"))
		atomic.StoreInt32(&bs.stopping, 1)
		atomic.StoreInt32(&bs.building, 0)
		return nil
	}
	return fmt.Errorf("no build in progress")
}

// ErrorFree checks that there has not been an error reported with
// ReportError
func (bs *BuildState) ErrorFree() bool {
	bs.errMutex.RLock()
	defer bs.errMutex.RUnlock()
	return bs.BuildError.err == nil
}

// GetError gets the currently stored error
func (bs *BuildState) GetError() error {
	bs.errMutex.RLock()
	defer bs.errMutex.RUnlock()
	return bs.BuildError.err
}

// SetExt inserts a key value pair into the external state store, currently only supports string
// and []string on the other side
func (bs *BuildState) SetExt(key string, value interface{}) {
	bs.extraMux.Lock()
	defer bs.extraMux.Unlock()
	bs.ExternExtras[key] = value
}

// GetExt gets a value based on the given key from the external state store
func (bs *BuildState) GetExt(key string) (interface{}, bool) {
	bs.extraMux.RLock()
	defer bs.extraMux.RUnlock()
	res, ok := bs.ExternExtras[key]
	return res, ok
}

//GetExtP gets the value for key and puts it in the object that out is pointing to, from the exposed store
func (bs *BuildState) GetExtP(key string, out interface{}) bool {
	tmp, ok := bs.GetExt(key)
	if !ok {
		return false
	}
	tmpBytes, err := json.Marshal(tmp)
	if err != nil {
		log.WithFields(log.Fields{"build": bs.BuildID, "error": err}).Warn("couldn't marshal the value")
		return false
	}
	err = json.Unmarshal(tmpBytes, out)
	if err != nil {
		log.WithFields(log.Fields{"build": bs.BuildID, "error": err}).Warn("couldn't unmarshal the value")
		return false
	}
	return true
}

// GetExtExtras gets the entire external state store as JSON
func (bs *BuildState) GetExtExtras() ([]byte, error) {
	bs.extraMux.RLock()
	defer bs.extraMux.RUnlock()
	return json.Marshal(bs.ExternExtras)
}

// Set stores a key value pair
func (bs *BuildState) Set(key string, value interface{}) {
	bs.extraMux.Lock()
	defer bs.extraMux.Unlock()
	bs.Extras[key] = value
}

//Get fetches a value as interface from the given key
func (bs *BuildState) Get(key string) (interface{}, bool) {
	bs.extraMux.RLock()
	defer bs.extraMux.RUnlock()
	out, ok := bs.Extras[key]
	return out, ok
}

//GetP gets the value for key and puts it in the object that out is pointing to
func (bs *BuildState) GetP(key string, out interface{}) bool {
	tmp, ok := bs.Get(key)
	if !ok {
		return false
	}
	tmpBytes, err := json.Marshal(tmp)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.WithFields(log.Fields{"build": bs.BuildID,
			"error": err, "file": file, "line": line}).Warn("couldn't marshal the value")
		return false
	}
	err = json.Unmarshal(tmpBytes, out)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.WithFields(log.Fields{"build": bs.BuildID,
			"error": err, "file": file, "line": line}).Warn("couldn't unmarshal the value")
		return false
	}
	return true
}

//GetExtras returns the internal state store as a map[string]interface
func (bs *BuildState) GetExtras() map[string]interface{} {
	return bs.Extras
}

// Write writes data to a file, creating it if it doesn't exist,
// deleting and recreating it if it does, should be used instead of golangs internal
// io library as bs one provides automatic file cleanup and separation of files among
// different builds.
func (bs *BuildState) Write(file string, data string) error {
	bs.mutex.Lock()
	bs.files = append(bs.files, file)
	bs.mutex.Unlock()
	filepath := "/tmp/" + bs.BuildID + "/" + file
	return ioutil.WriteFile(filepath, []byte(data), 0664)
}

// Defer adds a function to be executed asynchronously after the build is completed.
func (bs *BuildState) Defer(fn func()) {
	bs.extraMux.Lock()
	defer bs.extraMux.Unlock()
	bs.defers = append(bs.defers, fn)
}

// OnError adds a function to be executed upon a build finishing in the error state
func (bs *BuildState) OnError(fn func()) {
	bs.extraMux.Lock()
	defer bs.extraMux.Unlock()
	bs.errorCleanupFuncs = append(bs.errorCleanupFuncs, fn)
}

//OnDestroy registers a function to be called when this build state is terminated,
//such as when the testnet is destroyed. No ordering guarentees beyond this
func (bs *BuildState) OnDestroy(fn func()) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	bs.onDestroyFuncs = append(bs.onDestroyFuncs, fn)
}

// SetDeploySteps sets the number of steps in the deployment process.
// Should be given a number equivalent to the number of times
// IncrementDeployProgress will be called.
func (bs *BuildState) SetDeploySteps(steps int) {
	atomic.StoreUint64(&bs.DeployTotal, uint64(steps))
}

// IncrementDeployProgress increments the deploy process by one step. This is thread safe.
func (bs *BuildState) IncrementDeployProgress() {
	atomic.AddUint64(&bs.DeployProgress, 1)
}

// FinishDeploy signals that the deployment process has finished and the
// blockchain specific process will begin.
func (bs *BuildState) FinishDeploy() {
	atomic.StoreUint64(&bs.DeployProgress, atomic.LoadUint64(&bs.DeployTotal))
}

// SetBuildSteps sets the number of steps in the blockchain specific
// build process. Must be equivalent to the number of times IncrementBuildProgress()
// will be called.
func (bs *BuildState) SetBuildSteps(steps int) {
	atomic.StoreUint64(&bs.BuildTotal, uint64(steps+1)) //stay one step ahead to prevent early termination
}

// IncrementBuildProgress increments the build progress by one step.
func (bs *BuildState) IncrementBuildProgress() {
	atomic.AddUint64(&bs.BuildProgress, 1)
}

// FinishMainBuild sets the main build as finished, and signals the start of the
// side car build
func (bs *BuildState) FinishMainBuild() {
	atomic.StoreUint64(&bs.BuildProgress, atomic.LoadUint64(&bs.BuildTotal)-1)
}

// SetSidecarSteps sets the number of steps in the sidecar specific
// build process. Must be equivalent to the number of times IncrementBuildProgress()
// will be called.
func (bs *BuildState) SetSidecarSteps(steps int) {
	atomic.StoreUint64(&bs.SideCarTotal, uint64(steps))
}

// SetSidecars sets the number of sidecars
func (bs *BuildState) SetSidecars(sidecars int) {
	atomic.StoreUint64(&bs.SideCars, uint64(sidecars))
}

// IncrementSideCarProgress increments the sidecar build progress by one step.
func (bs *BuildState) IncrementSideCarProgress() {
	atomic.AddUint64(&bs.SideCarProgress, 1)
}

// GetProgress gets the progress as a percentage, within the range
// 0.0% - 100.0%
func (bs *BuildState) GetProgress() float64 {
	dp := atomic.LoadUint64(&bs.DeployProgress)
	dt := atomic.LoadUint64(&bs.DeployTotal)
	bp := atomic.LoadUint64(&bs.BuildProgress)
	bt := atomic.LoadUint64(&bs.BuildTotal)
	sp := atomic.LoadUint64(&bs.SideCarProgress)
	st := atomic.LoadUint64(&bs.SideCarTotal)

	var out float64
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
	buildTotalPercentage := 75.0
	if bs.SideCars > 5 {
		buildTotalPercentage -= 25.0
	} else {
		buildTotalPercentage -= float64(bs.SideCars) * 5.0
	}

	out += (float64(bp) / float64(bt)) * buildTotalPercentage

	if st == 0 || bs.SideCars == 0 {
		if !bs.Done() && out >= 100.0 {
			out = 99.99 //Out cannot be 100% unless the build is completed
		}
		return out
	}

	out += (float64(sp) / float64(st)) * (75.0 - buildTotalPercentage)

	if !bs.Done() && out >= 100.0 {
		out = 99.99 //Out cannot be 100% unless the build is completed
	}

	return out
}

// SetBuildStage updates the text which will be displayed along with the
// build progress percentage when the status of the build is queried.
func (bs *BuildState) SetBuildStage(stage string) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	bs.BuildStage = stage

}

// Reset sets the build state back the beginning. Used for when
// additional nodes are being added, as the stores may want to be reused
func (bs *BuildState) Reset() {
	atomic.StoreInt32(&bs.building, 1)
	atomic.StoreInt32(&bs.frozen, 0)
	atomic.StoreInt32(&bs.stopping, 0)

	bs.breakpoints = []float64{}

	bs.files = []string{}
	bs.defers = []func(){}

	bs.BuildError = CustomError{What: "", err: nil}
	bs.BuildStage = ""

	atomic.StoreUint64(&bs.DeployProgress, 0)
	atomic.StoreUint64(&bs.DeployTotal, 1)
	atomic.StoreUint64(&bs.BuildProgress, 0)
	atomic.StoreUint64(&bs.BuildTotal, 1)

	err := os.MkdirAll("/tmp/"+bs.BuildID, 0755)
	if err != nil {
		log.WithFields(log.Fields{"build": bs.BuildID, "error": err}).Panic("couldn't create the tmp folder")
	}
	log.WithFields(log.Fields{"build": bs.BuildID}).Info("build has been reset!")
}

//Marshal turns the BuildState into json representing the current progress of the build
func (bs *BuildState) Marshal() string {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	if bs.ErrorFree() { //error should be null if there is not an error
		return fmt.Sprintf("{\"progress\":%f,\"error\":null,\"stage\":\"%s\",\"frozen\":%v}", bs.GetProgress(), bs.BuildStage, bs.IsFrozen())
	}
	//otherwise give the error as an object
	out, _ := json.Marshal(
		map[string]interface{}{"progress": bs.GetProgress(), "error": bs.BuildError, "stage": bs.BuildStage, "frozen": bs.IsFrozen()})
	return string(out)
}

//Store saves the BuildState for later retrieval
func (bs *BuildState) Store() error {
	return db.SetMeta(bs.BuildID, *bs)
}

//Destroy deletes all storage of the BuildState
func (bs *BuildState) Destroy() error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	os.RemoveAll("/tmp/" + bs.BuildID)
	err := db.DeleteMeta(bs.BuildID)
	if err != nil {
		return utils.LogError(err)
	}
	for _, fn := range bs.onDestroyFuncs {
		go fn() //run eventually
	}
	bs.onDestroyFuncs = []func(){} //only once execution
	return nil
}
