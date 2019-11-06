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
	//TooSoonType is the type of a result from a command which tried to execute too soon
	TooSoonType = "TooSoon"
	//FatalType is the type of a result which indicates a fatal error
	FatalType = "Fatal"

	numberOfRetries = 4
	waitBeforeRetry = 10
)

var (
	statusSuccess = Result{Error: nil}
	statusTooSoon = Result{Type: TooSoonType, Error: fmt.Errorf("command ran too soon")}
)

// Executor executes commands according to their schedule and their dependencies
type Executor struct {
	// Runs the order of a command
	Runner func(ctx context.Context, order Order) Result
	// Schedules a command to be executed
	Scheduler func(command Command)
	// Checks whether a command executed
	CommandExecuted func(id string) bool
	// Supplies the current time
	TimeSupplier func() int64
}

//RunAsync runs a command asynchronously and calls the given callback on completion
func (c Executor) RunAsync(command Command, callback func(command Command, stat Result)) {
	go func() {
		callback(command, c.Run(command))
	}()
}

// Run runs a command. If it returns true, the command is considered executed and should be consumed. If it returns false, the transaction should be rolled back.
func (c Executor) Run(command Command) (stat Result) {
	stat, ok := c.checkSanity(command)
	if !ok {
		return
	}
	log.WithField("command", command).Trace("Running command")
	stat = c.execute(command)
	log.WithField("command", command).WithField("Result", stat).Info("Ran command")
	if !stat.IsSuccess() {
		if command.Retry < numberOfRetries && !stat.IsFatal() {
			retryCommand := command.GetRetryCommand(c.TimeSupplier())
			log.WithField("retryCommand", retryCommand).Warn("command failed, rescheduling")
			c.Scheduler(retryCommand)
		} else if stat.IsFatal() {
			log.WithField("command", command).Error("command resulted in a non-recoverable error")
		} else {
			log.WithField("command", command).Error("command failed too many times")
		}
	}
	return
}

func (c Executor) checkSanity(command Command) (stat Result, ok bool) {
	ok = true
	if c.TimeSupplier() < command.Timestamp {
		ok = false
		stat = statusTooSoon
		return
	}
	for _, dep := range command.Dependencies {
		if !c.CommandExecuted(dep) {
			ok = false
			stat = statusTooSoon
			return
		}
	}
	return
}

func (c Executor) execute(command Command) Result {
	if command.Timeout == 0 {
		return c.Runner(context.Background(), command.Order)
	}
	ctx, cancelFn := context.WithTimeout(context.Background(), command.Timeout)
	defer cancelFn()
	return c.Runner(ctx, command.Order)
}
