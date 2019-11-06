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

package executor

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/pkg/command"
)

//Result is the last known status of the command, contains a type and possibly an error
type Result struct {
	Error error
	Type  string
}

//IsSuccess returns whether or not the result indicates success
func (res Result) IsSuccess() bool {
	return res.Error == nil
}

//IsFatal returns true if there is an errr and it is marked as a fatal error,
//meaning it should not be reattempted
func (res Result) IsFatal() bool {
	return res.Error != nil && res.Type == FatalType
}

const (
	//SuccessType is the type of a successful result
	SuccessType = "Success"
	//TooSoonType is the type of a result from a cmdwhich tried to execute too soon
	TooSoonType = "TooSoon"
	//FatalType is the type of a result which indicates a fatal error
	FatalType = "Fatal"

	numberOfRetries = 4
	waitBeforeRetry = 10
)

var (
	statusSuccess = Result{Error: nil}
	statusTooSoon = Result{Type: TooSoonType, Error: fmt.Errorf("cmdran too soon")}
)

// Executor executes commands according to their schedule and their dependencies
type Executor struct {
	// Runs the order of a command
	Runner func(ctx context.Context, order command.Order) Result
	// Schedules a cmdto be executed
	Scheduler func(cmd command.Command)
	// Checks whether a cmdexecuted
	CommandExecuted func(id string) bool
	// Supplies the current time
	TimeSupplier func() int64
}

//RunAsync runs a cmdasynchronously and calls the given callback on completion
func (c Executor) RunAsync(cmd command.Command, callback func(cmd command.Command, stat Result)) {
	go func() {
		callback(cmd, c.Run(cmd))
	}()
}

// Run runs a command. If it returns true, the cmdis considered executed and should be consumed. If it returns false, the transaction should be rolled back.
func (c Executor) Run(cmd command.Command) (stat Result) {
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
			c.Scheduler(retryCommand)
		} else if stat.IsFatal() {
			log.WithField("command", cmd).Error("command resulted in a non-recoverable error")
		} else {
			log.WithField("command", cmd).Error("command failed too many times")
		}
	}
	return
}

func (c Executor) checkSanity(cmd command.Command) (stat Result, ok bool) {
	ok = true
	if c.TimeSupplier() < cmd.Timestamp {
		ok = false
		stat = statusTooSoon
		return
	}
	for _, dep := range cmd.Dependencies {
		if !c.CommandExecuted(dep) {
			ok = false
			stat = statusTooSoon
			return
		}
	}
	return
}

func (c Executor) execute(cmd command.Command) Result {
	if cmd.Timeout == 0 {
		return c.Runner(context.Background(), cmd.Order)
	}
	ctx, cancelFn := context.WithTimeout(context.Background(), cmd.Timeout)
	defer cancelFn()
	return c.Runner(ctx, cmd.Order)
}
