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
	"github.com/whiteblock/definition/command"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/usecase"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Executor handles the  processing of mutliple commands
type Executor interface {
	ExecuteCommands(cmds []command.Command) entity.Result
}

type executor struct {
	usecase usecase.DockerUseCase
	log     logrus.Ext1FieldLogger
}

// NewExecutor creates a new DeliveryHandler which uses the given usecase for
// executing the extracted command
func NewExecutor(
	usecase usecase.DockerUseCase,
	log logrus.Ext1FieldLogger) Executor {
	return &executor{usecase: usecase, log: log}
}

func (exec executor) ExecuteCommands(cmds []command.Command) entity.Result {
	resultChan := make(chan entity.Result, len(cmds))
	for _, cmd := range cmds {
		go func(cmd command.Command) {
			resultChan <- exec.usecase.Run(cmd)
		}(cmd)
	}
	var err error
	for range cmds {
		result := <-resultChan
		exec.log.WithFields(logrus.Fields{"success": result.IsSuccess()}).Trace("finished processing a command")
		if !result.IsSuccess() {
			exec.log.WithField("result", result).Error("a command failed to execute")
			err = errors.Wrap(err, result.Error.Error())
			if result.IsFatal() {
				return result
			}
		}
	}
	if err != nil {
		return entity.NewErrorResult(err)
	}
	return entity.NewSuccessResult()
}
