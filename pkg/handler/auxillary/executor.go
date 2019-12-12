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
			i := 0
			for ; i < exec.conf.ConnectionRetries; i++ {
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
				resultChan <- res
				break
			}
			if i == exec.conf.ConnectionRetries {
				resultChan <- entity.NewFatalResult("could not connect to docker")
			}
		}(cmd)
	}
	var err error
	isTrap := false
	for range cmds {
		result := <-resultChan
		exec.log.WithFields(logrus.Fields{"success": result.IsSuccess()}).Trace("finished processing a command")
		if !result.IsSuccess() {

			if result.IsFatal() {
				exec.log.WithField("result", result).Error("a command had a fatal error")
				if exec.conf.DebugMode {
					exec.log.Info("trapping fatal error due to debug mode")
					return entity.NewTrapResult()
				}
				return result
			}
			exec.log.WithField("result", result).Warn("a command failed to execute")
			err = fmt.Errorf("%v;%v", err, result.Error.Error())
		} else if result.IsTrap() {
			exec.log.WithField("result", result).Info("a command raised a trap")
			isTrap = true
		}
	}
	if err != nil {
		return entity.NewErrorResult(err)
	}
	if isTrap {
		return entity.NewTrapResult()
	}
	return entity.NewSuccessResult()
}
