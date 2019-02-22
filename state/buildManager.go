package state

import(
    "log"
    "sync"
    "errors"
)

var buildStates = []*BuildState{}

var serversInUse = []int{}

var mux = sync.RWMutex{}

func cleanBuildStates() {
    for i := 0; i < len(buildStates);i++ {
        if buildStates[i].Done() {
            for _,serverId1 := range buildStates[i].Servers {
                for j := 0; j < len(serversInUse); j++{
                    if serverId1 == serversInUse[j] {
                        serversInUse = append(serversInUse[:i],serversInUse[i+1:]...)
                        j--
                    }
                }
            }
            buildStates = append(buildStates[:i],buildStates[i+1:]...)
            i--
        }
    }
}

func GetBuildState(i int) *BuildState {
    mux.RLock()
    defer mux.RUnlock()

    if i >= len(buildStates) {
        log.Println("Requested an invalid build state")
        return nil
    }
    return buildStates[i]
}

func GetBuildStateByServerId(serverId int) *BuildState {
    mux.RLock()
    defer mux.RUnlock()

    for _, bs := range buildStates {
        for _, sid := range bs.Servers {
            if serverId == sid {
                return bs
            }
        }
    }
    return nil
}

/*
    AcquireBuilding acquires a build lock. Any function which modifies 
    the nodes in a testnet should only do so after calling this function 
    and ensuring that the returned value is nil
 */
func AcquireBuilding(servers []int) error {
    mux.Lock()
    defer mux.Unlock()

    cleanBuildStates()
    for _,id := range serversInUse {
        for _,id2 := range servers {
            if id == id2 {
                return errors.New("Error: Build in progress on one of the given servers")
            }
        }
    }
    buildStates = append(buildStates,NewBuildState(servers))
    serversInUse = append(serversInUse,servers...)
    return nil
}

/*
    DoneBuilding signals that the building process has finished and releases the
    build lock.
 */
func DoneBuilding(i int){
    bs := GetBuildState(i)
    if bs == nil {
        return
    }
    bs.DoneBuilding()
}


/*
    Stop checks if the stop signal has been sent. If this returns true,
    a building process should return. The ssh client checks this for you. 
 */
func Stop(serverId int) bool {

    bs := GetBuildStateByServerId(serverId)
    if bs == nil {
        log.Println("No build found for check")
        return false
    }
    return bs.Stop()
}

/*
    SignalStop flags that the current build should be stopped, if there is
    a current build. Returns an error if there is no build in progress
 */
func SignalStop(i int) error {
    bs := GetBuildState(i)
    if bs == nil {
        return errors.New("Build does not exist")
    }
    log.Printf("Sending stop signal to %d\n",i)
    return bs.SignalStop()
}
