/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package config

import (
	"fmt"
	"regexp"
)

func assertNotEmpty(s string, errMsg string) {
	if len(s) == 0 {
		panic(errMsg)
	}
}

// SanityCheck makes sure that the config is sane
func SanityCheck(conf Config) {
	log := conf.GetLogger()

	dockerSanityCheck(conf.Docker)
	log.Info("docker configuration checks passed")
}

var portRegexp = regexp.MustCompile(`[0-9]+`)

func dockerSanityCheck(conf Docker) {
	if conf.SwarmPort == 0 {
		panic("invalid docker swarm port given")
	}
	assertNotEmpty(conf.DaemonPort, "invalid docker daemon port given")
	assertNotEmpty(conf.GlusterImage, "missing gluster image")
	assertNotEmpty(conf.GlusterDriver, "missing gluster driver")

	if !portRegexp.MatchString(conf.DaemonPort) {
		panic(fmt.Sprintf(`daemon port is invalid: "%s"`, conf.DaemonPort))
	}
}
