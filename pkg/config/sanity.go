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

package config

import (
	"os"
)

func assertNotEmpty(s string, errMsg string) {
	if len(s) == 0 {
		panic(errMsg)
	}
}

func SanityCheck(conf Config) {
	log := conf.GetLogger()

	dockerSanityCheck(conf.Docker)
	log.Info("docker configuration checks passed")
}

func dockerSanityCheck(conf Docker) {
	if !conf.LocalMode {
		dockerFilesConfCheck(conf)
	}
	if conf.SwarmPort == 0 {
		panic("invalid docker swarm port given")
	}
	assertNotEmpty(conf.DaemonPort, "invalid docker daemon port given")
	assertNotEmpty(conf.GlusterImage, "missing gluster image")
	assertNotEmpty(conf.GlusterDriver, "missing gluster driver")
}

func dockerFilesConfCheck(conf Docker) {
	_, err := os.Lstat(conf.CACertPath)
	if err != nil {
		panic(err)
	}

	_, err = os.Lstat(conf.CertPath)
	if err != nil {
		panic(err)
	}

	_, err = os.Lstat(conf.KeyPath)
	if err != nil {
		panic(err)
	}
}
