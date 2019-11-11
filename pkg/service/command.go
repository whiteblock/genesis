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

package service

import (
	"github.com/whiteblock/genesis/pkg/command"
	"github.com/whiteblock/genesis/pkg/entity"
	"github.com/whiteblock/genesis/pkg/repository"
)

//CommandService is an interface to where the commands are kept for querying
type CommandService interface {
	//CheckDependenciesExecuted returns true if all of the commands dependencies have executed
	CheckDependenciesExecuted(cmd command.Command) (bool, error)
	//ReportCommandResult reports that the command has finished execution, with the result of execution
	ReportCommandResult(cmd command.Command, res entity.Result) error
}

type commandService struct {
	repo repository.CommandRepository
}

//NewCommandService creates a new CommandService
func NewCommandService(repo repository.CommandRepository) CommandService {
	return &commandService{repo: repo}
}

type depCheckResult struct {
	executed bool
	err      error
}

//CheckDependenciesExecuted returns true if all of the commands dependencies have executed
func (cs commandService) CheckDependenciesExecuted(cmd command.Command) (bool, error) {
	resChan := make(chan depCheckResult, len(cmd.Dependencies))
	for _, dep := range cmd.Dependencies {
		go func(depID string) {
			executed, err := cs.repo.HasCommandExecuted(depID)
			resChan <- depCheckResult{executed: executed, err: err}
		}(dep)
	}
	for range cmd.Dependencies {
		result := <-resChan
		if result.err != nil {
			return false, result.err
		}
		if result.executed == false {
			return false, nil
		}
	}
	return true, nil
}

// ReportCommandResult reports that the command has finished execution, with the result of execution
func (cs commandService) ReportCommandResult(cmd command.Command, res entity.Result) error {
	return cs.repo.ReportCommandFinished(cmd.ID, res)
}
