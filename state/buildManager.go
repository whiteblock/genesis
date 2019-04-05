package state

import (
	"errors"
	"log"
	"sync"
)

var buildStates = []*BuildState{}

var serversInUse = []int{}

var mux = sync.RWMutex{}

/*
   Remove all of the finished build states

*/
func cleanBuildStates(servers []int) {
	for i := 0; i < len(buildStates); i++ {
		if buildStates[i].Done() {
			needsToDie := false
			for _, serverId1 := range buildStates[i].Servers { //Check if the build actually needs to be removed.
				for _, serverId2 := range servers {
					if serverId1 == serverId2 {
						needsToDie = true
					}
				}
			}
			if !needsToDie {
				continue
			}
			//Remove the build state
			for _, serverId1 := range buildStates[i].Servers {
				for j := 0; j < len(serversInUse); j++ {
					if serverId1 == serversInUse[j] {
						serversInUse = append(serversInUse[:i], serversInUse[i+1:]...)
						j--
					}
				}
			}
			buildStates = append(buildStates[:i], buildStates[i+1:]...)
			i--
		}
	}
}

/*
   Get the current build state for a server.
*/
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
   Get the current build state based off the build id.
   Will given an error if the build is not found
*/
func GetBuildStateById(buildId string) (*BuildState, error) {
	mux.RLock()
	defer mux.RUnlock()

	for _, bs := range buildStates {
		if bs.BuildId == buildId {
			return bs, nil
		}
	}

	return nil, errors.New("Couldn't find the request build")
}

/*
   AcquireBuilding acquires a build lock. Any function which modifies
   the nodes in a testnet should only do so after calling this function
   and ensuring that the returned value is nil
*/
func AcquireBuilding(servers []int, buildId string) error {
	mux.Lock()
	defer mux.Unlock()

	cleanBuildStates(servers)
	for _, id := range serversInUse {
		for _, id2 := range servers {
			if id == id2 {
				return errors.New("Error: Build in progress on one of the given servers")
			}
		}
	}
	buildStates = append(buildStates, NewBuildState(servers, buildId))
	serversInUse = append(serversInUse, servers...)
	return nil
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
   a current build. Returns an error if there is no build in progress. Signal
   the build to stop by the build id.
*/
func SignalStop(buildId string) error {
	bs, err := GetBuildStateById(buildId)
	if err != nil {
		log.Println(err)
		return err
	}
	if bs == nil {
		return errors.New("Build does not exist")
	}
	log.Printf("Sending stop signal to build:%s\n", buildId)
	return bs.SignalStop()
}
