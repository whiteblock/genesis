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

//func TestCommandService_CheckDependenciesExecuted(t *testing.T) {
//	repo := new(repository.CommandRepository)
//	repo.On("HasCommandExecuted", "test1").Return(true, nil)
//	repo.On("HasCommandExecuted", "test2").Return(false, nil)
//
//	serv := NewCommandService(repo)
//
//	ready, err := serv.CheckDependenciesExecuted(command.Command{
//		ID:      "TEST",
//		Timeout: 0,
//		Target:  command.Target{IP: "0.0.0.0"},
//		Order: command.Order{
//			Type:    "createContainer",
//			Payload: map[string]interface{}{},
//		},
//		Dependencies: []string{"test1", "test2"},
//	})
//	assert.NoError(t, err)
//	assert.False(t, ready)
//}
//
//func TestCommandService_ReportCommandResult(t *testing.T) {
//	//TODO
//}
