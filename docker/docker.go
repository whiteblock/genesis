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

package docker

import (
	"context"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os/exec"
)

type Executor struct {
	//Docker daemon endpoint
	Endpoint string
	//path to cert for auth
	CertPath string
	//path to key for auth
	KeyPath string
	//context for aborting the execution
	Ctx context.Context
}

type Result struct {
	Error  error
	Stdout []byte
	Stderr []byte
}

func (e Executor) Run(cmd string) (result Result) {
	command := fmt.Sprintf("docker -H %s --tlscert %s --tlskey %s %s", e.Endpoint, e.CertPath, e.KeyPath, cmd)
	log.WithFields(log.Fields{"remote": e.Endpoint, "command": command}).Debug("engaging with the docker daemon")
	proc := exec.Command(command)
	err := proc.Start()
	if err != nil {
		result.Error = err
		return
	}

	stdout, err := proc.StdoutPipe()
	if err != nil {
		result.Error = err
		return
	}

	stderr, err := proc.StderrPipe()
	if err != nil {
		result.Error = err
		return
	}

	result.Stdout, err = ioutil.ReadAll(stdout)
	if err != nil {
		result.Error = err
		return
	}

	result.Stderr, err = ioutil.ReadAll(stderr)
	if err != nil {
		result.Error = err
		return
	}

	result.Error = proc.Wait()
	return
}

func (e Executor) RunAsync(cmd string, callback func(result Result)) {
	go callback(e.Run(cmd))
}
