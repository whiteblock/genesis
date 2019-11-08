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
)

//CommandService is an interface to where the commands are kept for querying
type CommandService interface {
	//CheckDependenciesExecuted returns true if all of the commands dependencies have executed
	CheckDependenciesExecuted(cmd command.Command) bool
}

type commandService struct {
	//TODO
}

//NewCommandService creates a new CommandService
func NewCommandService() CommandService {
	return &commandService{}
}

//CheckDependenciesExecuted returns true if all of the commands dependencies have executed
func (cs commandService) CheckDependenciesExecuted(cmd command.Command) bool {
	//TODO
	return false
}