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

package repository

import (
	"github.com/whiteblock/genesis/pkg/entity"
	"sync"
)

//CommandRepository presents access to whatever holds the commands
type CommandRepository interface {
	//HasCommandExecuted checks whether or not a command with the given id has executed
	HasCommandExecuted(id string) (bool, error)
	//ReportCommandFinished reports that the command with the given id has been executed
	ReportCommandFinished(id string, res entity.Result) error
}

//LOCAL IMPLEMENTATION

type localCommandRepository struct {
	mux      *sync.Mutex
	executed map[string]bool
}

//NewLocalCommandRepository creates a new local command repository. This should only be used when Genesis
//is being used a standalone application
func NewLocalCommandRepository() CommandRepository {
	return localCommandRepository{mux: &sync.Mutex{}, executed: map[string]bool{}}
}

func (lcr localCommandRepository) HasCommandExecuted(id string) (bool, error) {
	lcr.mux.Lock()
	defer lcr.mux.Unlock()
	_, ok := lcr.executed[id]
	return ok, nil
}

func (lcr localCommandRepository) ReportCommandFinished(id string, res entity.Result) error {
	lcr.mux.Lock()
	defer lcr.mux.Unlock()
	lcr.executed[id] = (res.Error == nil)
	return nil
}
