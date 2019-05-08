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

//Package registrar handles the mappings between the blockchain libraries in a more scalable manor.
package registrar

import (
	"../../testnet"
	"../../util"
	"fmt"
	"sync"
)

var (
	//mux is actually not needed because all Register calls should be done via the init call, which
	//means that there should not be a race condition. However, golang does not provide any
	//method of enforcement for this to my knowledge.
	mux        = &sync.RWMutex{}
	buildFuncs = map[string]func(*testnet.TestNet) error{}
	addFuncs   = map[string]func(*testnet.TestNet) error{}

	serviceFuncs  = map[string]func() []util.Service{}
	paramsFuncs   = map[string]func() string{}
	defaultsFuncs = map[string]func() string{}
	logFiles      = map[string]map[string]string{}
)

// RegisterBuild associates a blockchain name with a build process
func RegisterBuild(blockchain string, fn func(*testnet.TestNet) error) {
	mux.Lock()
	defer mux.Unlock()
	buildFuncs[blockchain] = fn
}

// RegisterAddNodes associates a blockchain name with a add node process
func RegisterAddNodes(blockchain string, fn func(*testnet.TestNet) error) {
	mux.Lock()
	defer mux.Unlock()
	addFuncs[blockchain] = fn
}

// RegisterServices associates a blockchain name with a function that gets its required services
func RegisterServices(blockchain string, fn func() []util.Service) {
	mux.Lock()
	defer mux.Unlock()
	serviceFuncs[blockchain] = fn
}

// RegisterParams associates a blockchain name with a function that gets its parameters
func RegisterParams(blockchain string, fn func() string) {
	mux.Lock()
	defer mux.Unlock()
	paramsFuncs[blockchain] = fn
}

// RegisterDefaults associates a blockchain name with a function that gets its default parameter values
func RegisterDefaults(blockchain string, fn func() string) {
	mux.Lock()
	defer mux.Unlock()
	defaultsFuncs[blockchain] = fn
}

// RegisterAdditionalLogs associates a blockchain name with a map of additional logs
func RegisterAdditionalLogs(blockchain string, logs map[string]string) {
	mux.Lock()
	defer mux.Unlock()
	logFiles[blockchain] = logs
}

// GetBuildFunc gets the build function associated with the given blockchain name or error != nil if
// it is not found
func GetBuildFunc(blockchain string) (func(*testnet.TestNet) error, error) {
	mux.RLock()
	defer mux.RUnlock()
	out, ok := buildFuncs[blockchain]
	if !ok {
		return nil, fmt.Errorf("no entry found for blockchain \"%s\"", blockchain)
	}
	return out, nil
}

// GetAddNodeFunc gets the add node function associated with the given blockchain name or error != nil if
// it is not found
func GetAddNodeFunc(blockchain string) (func(*testnet.TestNet) error, error) {
	mux.RLock()
	defer mux.RUnlock()
	out, ok := addFuncs[blockchain]
	if !ok {
		return nil, fmt.Errorf("no entry found for blockchain \"%s\"", blockchain)
	}
	return out, nil
}

// GetServiceFunc gets the service function associated with the given blockchain name or error != nil if
// it is not found
func GetServiceFunc(blockchain string) (func() []util.Service, error) {
	mux.RLock()
	defer mux.RUnlock()
	out, ok := serviceFuncs[blockchain]
	if !ok {
		return nil, fmt.Errorf("no entry found for blockchain \"%s\"", blockchain)
	}
	return out, nil
}

// GetParamsFunc gets the Params function associated with the given blockchain name or error != nil if
// it is not found
func GetParamsFunc(blockchain string) (func() string, error) {
	mux.RLock()
	defer mux.RUnlock()
	out, ok := paramsFuncs[blockchain]
	if !ok {
		return nil, fmt.Errorf("no entry found for blockchain \"%s\"", blockchain)
	}
	return out, nil
}

// GetDefaultsFunc gets the Defaults function associated with the given blockchain name or error != nil if
// it is not found
func GetDefaultsFunc(blockchain string) (func() string, error) {
	mux.RLock()
	defer mux.RUnlock()
	out, ok := defaultsFuncs[blockchain]
	if !ok {
		return nil, fmt.Errorf("no entry found for blockchain \"%s\"", blockchain)
	}
	return out, nil
}

// GetAdditionalLogs gets additional logs of the blockchain if there are any
func GetAdditionalLogs(blockchain string) map[string]string {
	mux.RLock()
	defer mux.RUnlock()
	return logFiles[blockchain]
}

// GetSupportedBlockchains gets the blockchains which have a registered
// Build function
func GetSupportedBlockchains() []string {
	mux.RLock()
	defer mux.RUnlock()
	out := []string{}
	for blockchain := range buildFuncs {
		out = append(out, blockchain)
	}
	return out
}
