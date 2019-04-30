//Package registrar handles the mappings between the blockchain libraries in a more scalable manor.
package registrar

import (
	"../../testnet"
	"../../util"
	"fmt"
	"sync"
)

var (
	mux           = &sync.RWMutex{}
	buildFuncs    = map[string]func(*testnet.TestNet) ([]string, error){}
	addFuncs      = map[string]func(*testnet.TestNet) ([]string, error){}
	serviceFuncs  = map[string]func() []util.Service{}
	paramsFuncs   = map[string]func() string{}
	defaultsFuncs = map[string]func() string{}
	logFiles      = map[string]map[string]string{}
)

// RegisterBuild associates a blockchain name with a build process
func RegisterBuild(blockchain string, fn func(*testnet.TestNet) ([]string, error)) {
	mux.Lock()
	defer mux.Unlock()
	buildFuncs[blockchain] = fn
}

// RegisterAddNodes associates a blockchain name with a add node process
func RegisterAddNodes(blockchain string, fn func(*testnet.TestNet) ([]string, error)) {
	mux.Lock()
	defer mux.Unlock()
	addFuncs[blockchain] = fn
}

// RegisterAddNodes associates a blockchain name with a function that gets its required services
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

// GetBuildFunc get the build function associated with the given blockchain name or error != nil if
// it is not found
func GetBuildFunc(blockchain string) (func(*testnet.TestNet) ([]string, error), error) {
	mux.RLock()
	defer mux.RUnlock()
	out, ok := buildFuncs[blockchain]
	if !ok {
		return nil, fmt.Errorf("no entry found for blockchain \"%s\"", blockchain)
	}
	return out, nil
}

// GetAddNodeFunc get the add node function associated with the given blockchain name or error != nil if
// it is not found
func GetAddNodeFunc(blockchain string) (func(*testnet.TestNet) ([]string, error), error) {
	mux.RLock()
	defer mux.RUnlock()
	out, ok := addFuncs[blockchain]
	if !ok {
		return nil, fmt.Errorf("no entry found for blockchain \"%s\"", blockchain)
	}
	return out, nil
}

// GetServiceFunc get the service function associated with the given blockchain name or error != nil if
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

// GetAdditionalLogs get additional logs of the blockchain if there are any
func GetAdditionalLogs(blockchain string) map[string]string {
	mux.RLock()
	defer mux.RUnlock()
	return logFiles[blockchain]
}
