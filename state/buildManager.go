/*
	Copyright 2019 Whiteblock Inc.
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

package state

import (
	"fmt"
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
			for _, serverID1 := range buildStates[i].Servers { //Check if the build actually needs to be removed.
				for _, serverID2 := range servers {
					if serverID1 == serverID2 {
						needsToDie = true
					}
				}
			}
			if !needsToDie {
				continue
			}
			//Remove the build state
			for _, serverID1 := range buildStates[i].Servers {
				for j := 0; j < len(serversInUse); j++ {
					if serverID1 == serversInUse[j] {
						serversInUse = append(serversInUse[:i], serversInUse[i+1:]...)
						j--
					}
				}
			}
			buildStates[i].Destroy()
			buildStates = append(buildStates[:i], buildStates[i+1:]...)
			i--
		}
	}
}

// GetBuildStateByServerID gets the current build state on a server. DEPRECATED, use
// GetBuildStateById instead.
func GetBuildStateByServerID(serverID int) *BuildState {
	mux.RLock()
	defer mux.RUnlock()

	for _, bs := range buildStates {
		for _, sid := range bs.Servers {
			if serverID == sid {
				return bs
			}
		}
	}
	return nil
}

// GetBuildStateByID gets the current build state based off the build id.
// Will given an error if the build is not found
func GetBuildStateByID(buildID string) (*BuildState, error) {
	mux.RLock()

	for _, bs := range buildStates {
		if bs.BuildID == buildID {
			mux.RUnlock()
			return bs, nil
		}
	}
	mux.RUnlock()
	mux.Lock()
	defer mux.Unlock()
	bs, err := RestoreBuildState(buildID)
	if err != nil || bs == nil {
		log.Println(err)
		return nil, fmt.Errorf("couldn't find the request build")
	}
	buildStates = append(buildStates, bs)
	serversInUse = append(serversInUse, bs.Servers...)
	return bs, nil
}

// AcquireBuilding acquires a build lock. Any function which modifies
// the nodes in a testnet should only do so after calling this function
// and ensuring that the returned value is nil
func AcquireBuilding(servers []int, buildID string) error {
	mux.Lock()
	defer mux.Unlock()

	cleanBuildStates(servers)
	for _, id := range serversInUse {
		for _, id2 := range servers {
			if id == id2 {
				return fmt.Errorf("error: Build in progress on server %d", id)
			}
		}
	}
	buildStates = append(buildStates, NewBuildState(servers, buildID))
	serversInUse = append(serversInUse, servers...)
	return nil
}

// Stop checks if the stop signal has been sent. If this returns true,
// a building process should return. The ssh client checks this for you.
// This is fairly naive and will need to be changed for multi-tenancy
func Stop(serverID int) bool {

	bs := GetBuildStateByServerID(serverID)
	if bs == nil {
		log.Println("No build found for check")
		return false
	}
	return bs.Stop()
}

// SignalStop flags that the current build should be stopped, if there is
// a current build. Returns an error if there is no build in progress. Signal
// the build to stop by the build id.
func SignalStop(buildID string) error {
	bs, err := GetBuildStateByID(buildID)
	if err != nil {
		log.Println(err)
		return err
	}
	if bs == nil {
		return fmt.Errorf("build \"%s\" does not exist", buildID)
	}
	log.Printf("Sending stop signal to build:%s\n", buildID)
	return bs.SignalStop()
}
