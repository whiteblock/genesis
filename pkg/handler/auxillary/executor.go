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

package auxillary

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/whiteblock/genesis/pkg/config"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/usecase"

	"github.com/sirupsen/logrus"
	"github.com/whiteblock/definition/command"
	"golang.org/x/sync/semaphore"
)

// Executor handles the  processing of mutliple commands
type Executor interface {
	ExecuteCommands(cmds []command.Command) entity.Result
}

type executor struct {
	usecase usecase.DockerUseCase
	conf    config.Execution
	log     logrus.Ext1FieldLogger
}

// ErrDockerConnFailed is the error for when the docker daemon is unreachable
var ErrDockerConnFailed = entity.NewFatalResult("could not connect to docker")

// NewExecutor creates a new DeliveryHandler which uses the given usecase for
// executing the extracted command
func NewExecutor(
	conf config.Execution,
	usecase usecase.DockerUseCase,
	log logrus.Ext1FieldLogger) Executor {
	return &executor{usecase: usecase, conf: conf, log: log}
}

func (exec executor) ExecuteCommands(cmds []command.Command) entity.Result {
	resultChan := make(chan entity.Result, len(cmds))
	sem := semaphore.NewWeighted(exec.conf.LimitPerTest)
	for _, cmd := range cmds {
		go func(cmd command.Command) {

			for i := 0; i < exec.conf.ConnectionRetries; i++ {
				sem.Acquire(context.Background(), 1)
				res := exec.usecase.Run(cmd)
				sem.Release(1)
				if !res.IsSuccess() && strings.Contains(res.Error.Error(), "connect to the Docker daemon") {
					exec.log.WithFields(logrus.Fields{
						"result":  res,
						"time":    exec.conf.RetryDelay,
						"attempt": i,
					}).Info("connection to docker failed, retrying")
					time.Sleep(exec.conf.RetryDelay)
					continue
				}
				resultChan <- res.InjectMeta(map[string]interface{}{
					"command": cmd,
					"attempt": i,
				})
				return
			}
			resultChan <- ErrDockerConnFailed.InjectMeta(
				map[string]interface{}{
					"command": cmd,
				})
		}(cmd)
	}
	var err error
	isTrap := false
	failed := []string{}
	for range cmds {
		result := <-resultChan
		entry := exec.log.WithField("result", result)
		entry.Trace("finished processing a command")
		if !result.IsSuccess() {

			if result.IsFatal() {
				entry.Error("a command had a fatal error")
				return result
			}
			failed = append(failed, result.Meta["command"].(command.Command).ID)
			entry.Warn("a command failed to execute")
			err = fmt.Errorf("%v;%v", err, result.Error.Error())
		} else if result.IsTrap() {
			entry.Info("a command raised a trap")
			isTrap = true
		}
	}
	if err != nil {
		return entity.NewErrorResult(err).InjectMeta(map[string]interface{}{
			"failed": failed,
		})
	}
	if isTrap {
		return entity.NewTrapResult()
	}
	return entity.NewSuccessResult()
}
