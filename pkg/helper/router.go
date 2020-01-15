/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package helper

import (
	"github.com/gorilla/mux"
	"net/http"
)

//Router represents the functionality needed by the REST router
type Router interface {
	//HandleFunc see github.com/gorilla/mux.Router.HandleFunc
	HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *mux.Route
	//ServeHTTP see github.com/gorilla/mux.Router.ServeHTTP
	ServeHTTP(http.ResponseWriter, *http.Request)
}
