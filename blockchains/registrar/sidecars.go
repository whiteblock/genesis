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

package registrar

import (
	"fmt"
	"github.com/Whiteblock/genesis/testnet"
)

// SideCar represents the side car registration details needed for building or other purposes
type SideCar struct {
	// Image is the docker image to build the side car from
	Image string
}

var (
	sideCars           = map[string]SideCar{}
	blockchainSideCars = map[string][]string{}
	sideCarBuildFuncs  = map[string]func(*testnet.Adjunct) error{}
	sideCarAddFuncs    = map[string]func(*testnet.Adjunct) error{}
)

// RegisterBlockchainSideCars associates a blockchain name with a
func RegisterBlockchainSideCars(blockchain string, scs []string) {
	mux.Lock()
	defer mux.Unlock()
	blockchainSideCars[blockchain] = scs
}

// RegisterSideCar associates a blockchain name with a
func RegisterSideCar(name string, sc SideCar) {
	mux.Lock()
	defer mux.Unlock()
	sideCars[name] = sc
}

// RegisterAddSideCar associates a blockchain name with a add node process
func RegisterAddSideCar(sideCarName string, fn func(*testnet.Adjunct) error) {
	mux.Lock()
	defer mux.Unlock()
	sideCarAddFuncs[sideCarName] = fn
}

// RegisterBuildSideCar associates a blockchain name with a add node process
func RegisterBuildSideCar(sideCarName string, fn func(*testnet.Adjunct) error) {
	mux.Lock()
	defer mux.Unlock()
	sideCarBuildFuncs[sideCarName] = fn
}

// GetBlockchainSideCars associates a blockchain name with a
func GetBlockchainSideCars(blockchain string) ([]string, error) {
	mux.RLock()
	defer mux.RUnlock()
	out, ok := blockchainSideCars[blockchain]
	if !ok {
		return nil, fmt.Errorf("no entry found for blockchain \"%s\"", blockchain)
	}
	return out, nil
}

// GetAddSideCar gets the function to add a sidecar
func GetAddSideCar(sideCarName string) (func(*testnet.Adjunct) error, error) {
	mux.RLock()
	defer mux.RUnlock()
	out, ok := sideCarAddFuncs[sideCarName]
	if !ok {
		return nil, fmt.Errorf("no entry found for side car \"%s\"", sideCarName)
	}
	return out, nil
}

// GetBuildSideCar gets the function to build a sidecar
func GetBuildSideCar(sideCarName string) (func(*testnet.Adjunct) error, error) {
	mux.RLock()
	defer mux.RUnlock()
	out, ok := sideCarBuildFuncs[sideCarName]
	if !ok {
		return nil, fmt.Errorf("no entry found for side car \"%s\"", sideCarName)
	}
	return out, nil
}

// GetSideCar gets the details about a sidecar
func GetSideCar(sideCarName string) (*SideCar, error) {
	mux.Lock()
	defer mux.Unlock()
	out, ok := sideCars[sideCarName]
	if !ok {
		return nil, fmt.Errorf("no entry found for side car \"%s\"", sideCarName)
	}
	return &out, nil
}
