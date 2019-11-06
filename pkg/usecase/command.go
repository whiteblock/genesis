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

package usecase

import(
	"github.com/docker/docker/client"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/service"
)

type CommandUseCase interface {
	CheckDependenciesExecuted(cmd command.Command) bool

}

type commandUseCase struct {
}

func NewCommandUseCase() ( CommandUseCase, error) {
	return commandUseCase{},nil
}


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
	along with this program. If not, see <https://www.gnu.org/licenses/>.
*/

package handler

import (
	"github.com/whiteblock/registrar/pkg/command"
	"github.com/whiteblock/registrar/pkg/entity"
	"github.com/whiteblock/registrar/pkg/usecase"
	"github.com/whiteblock/registrar/pkg/service"
)

const (
	numberOfRetries = 4
	waitBeforeRetry = 10
)

var (
	statusSuccess = entity.Result{Error: nil}
	statusTooSoon = entity.Result{Type: entity.TooSoonType, Error: fmt.Errorf("command ran too soon")}
)

type CommandUseCase interface {
	Run(cmd command.Command) entity.Result
	RunAsync(cmd command.Command, callback func(cmd command.Command, stat entity.Result))
}


type commandUseCase struct {
	service service.CommandService
	dockerUseCase DockerUseCase
	
}

//RunAsync runs a cmdasynchronously and calls the given callback on completion
func (c commandUseCase) RunAsync(cmd command.Command, callback func(cmd command.Command, stat entity.Result)) {
	go func() {
		callback(cmd, c.Run(cmd))
	}()
}

// Run runs a command. If it returns true, the cmdis considered executed and should be consumed. If it returns false, the transaction should be rolled back.
func (c commandUseCase) Run(cmd command.Command) (stat entity.Result) {
	stat, ok := c.checkSanity(cmd)
	if !ok {
		return
	}
	log.WithField("command", cmd).Trace("Running command")
	stat = c.execute(cmd)
	log.WithField("command", cmd).WithField("Result", stat).Info("Ran command")
	if !stat.IsSuccess() {
		if cmd.Retry < numberOfRetries && !stat.IsFatal() {
			retryCommand := cmd.GetRetryCommand(c.TimeSupplier() + waitBeforeRetry)
			log.WithField("retryCommand", retryCommand).Warn("cmdfailed, rescheduling")
		} else if stat.IsFatal() {
			log.WithField("command", cmd).Error("command resulted in a non-recoverable error")
		} else {
			log.WithField("command", cmd).Error("command failed too many times")
		}
	}
	return
}

func (c commandUseCase) checkSanity(cmd command.Command) (stat entity.Result, ok bool) {
	ok = true
	if c.TimeSupplier() < cmd.Timestamp {
		ok = false
		stat = statusTooSoon
		return
	}
	if !c.service.CheckDependenciesExecuted(cmd) {
		ok = false
		stat = statusTooSoon
		return 
	}
	return
}

func (c commandUseCase) execute(cmd command.Command) entity.Result {
	if cmd.Timeout == 0 {
		return c.Runner(context.Background(), cmd.Order)
	}
	ctx, cancelFn := context.WithTimeout(context.Background(), cmd.Timeout)
	defer cancelFn()
	return c.Runner(ctx, cmd.Order)
}

func (c commandUseCase) route(ctx context.Context,cmd command.Command) entity.Result {
	
}