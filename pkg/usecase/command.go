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

package usecase

import (
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/service"
	"time"
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
	TimeSupplier() int64
}

type commandUseCase struct {
	service       service.CommandService
	dockerUseCase DockerUseCase
}

func (c commandUseCase) TimeSupplier() int64 {
	return time.Now().Unix()
}

// Run runs a command. If it returns true, the cmdis considered executed and should be consumed. If it returns false, the transaction should be rolled back.
func (c commandUseCase) Run(cmd command.Command) entity.Result {
	stat, ok := c.checkSanity(cmd)
	if !ok {
		return stat
	}
	log.WithField("command", cmd).Trace("running command")
	return c.execute(cmd)
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
		return c.dockerUseCase.Execute(context.Background(), cmd.Order)
	}
	ctx, cancelFn := context.WithTimeout(context.Background(), cmd.Timeout)
	defer cancelFn()
	return c.dockerUseCase.Runner(ctx, cmd.Order)
}
