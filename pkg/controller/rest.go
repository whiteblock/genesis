/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package controller

import (
	"net/http"
	"strings"

	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/handler"
	"github.com/whiteblock/genesis/pkg/helper"

	"github.com/sirupsen/logrus"
)

//RestController handles the REST API server
type RestController interface {
	//Start attempts to start the server
	Start()
}

type restController struct {
	conf entity.RestConfig
	hand handler.RestHandler
	mux  helper.Router
	log  logrus.Ext1FieldLogger
}

//NewRestController creates a new rest controller
func NewRestController(
	conf entity.RestConfig,
	hand handler.RestHandler,
	mux helper.Router,
	log logrus.Ext1FieldLogger) RestController {

	log.Trace("creating a new rest controller")
	return &restController{conf: conf, hand: hand, mux: mux, log: log}
}

// Start starts the rest server, blocking the calling thread from returning
func (rc restController) Start() {

	rc.mux.HandleFunc("/command", rc.hand.AddCommands).Methods("POST")
	rc.mux.HandleFunc("/health", rc.hand.HealthCheck).Methods("GET")

	rc.log.WithFields(logrus.Fields{"socket": rc.conf.Listen}).Info("listening for requests")
	rc.log.Fatal(http.ListenAndServe(rc.conf.Listen, removeTrailingSlash(rc.mux)))
}

func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		next.ServeHTTP(w, r)
	})
}
