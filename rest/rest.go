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

// Package rest implements the REST interface which is used to communicate with this module
package rest

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/util"
	"net/http"
	"strings"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

// StartServer starts the rest server, blocking the calling thread from returning
func StartServer() {
	router := mux.NewRouter()

	router.HandleFunc("/command", addCommand).Methods("POST")
	router.HandleFunc("/health", addCommand).Methods("GET")

	log.WithFields(log.Fields{"socket": conf.Listen}).Info("listening for requests")
	log.Fatal(http.ListenAndServe(conf.Listen, removeTrailingSlash(router)))
}

func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		next.ServeHTTP(w, r)
	})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Fatal(err)
	}
}
