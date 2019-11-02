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

package command

import (
	log "github.com/sirupsen/logrus"
)

type commandstatus string

const (
	success = commandstatus("success")
	failure = commandstatus("failure")
	later   = commandstatus("later")

	numberOfRetries = 4
	waitBeforeRetry = 10
)

// Executor executes commands according to their schedule and their dependencies
type Executor struct {
	// Runs the order of a command
	Runner func(order Order) bool
	// Schedules a command to be executed
	Scheduler func(command Command)
	// Checks whether a command executed
	CommandExecuted func(id string) bool
	// Supplies the current time
	TimeSupplier func() int64
}

// Execute runs a command. If it returns true, the command is considered executed and should be consumed. If it returns false, the transaction should be rolled back.
func (c *Executor) Execute(command Command) bool {
	log.WithField("command", command).Trace("Running command")
	status := c.executeCommand(command)
	log.WithField("command", command).WithField("status", status).Info("Ran command")
	if status == later {
		return false
	}
	if status == failure {
		if command.Retry < numberOfRetries {
			retryCommand := Command{command.ID, c.TimeSupplier() + waitBeforeRetry, command.Retry + 1, command.Target, command.Dependencies, command.Order}
			log.WithField("retryCommand", retryCommand).Warn("Command failed, rescheduling")
			c.Scheduler(retryCommand)
		} else {
			log.WithField("command", command).Error("Command failed too many times")
		}
	}
	return true
}

func (c *Executor) executeCommand(command Command) commandstatus {

	if c.TimeSupplier() < command.Timestamp {
		return later
	}
	for _, dep := range command.Dependencies {
		if !c.CommandExecuted(dep) {
			return later
		}
	}

	if c.Runner(command.Order) {
		return success
	}
	return failure
}
