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
	"sync"
	"time"
)

// State represents the local state of Genesis.
type State struct {
	ExecutedCommands map[string]Command
	PendingCommands  []Command
	executor         Executor
	mu               sync.Mutex
}

var commandState State

// AddCommand adds a command to the commands to be executed
func (s *State) AddCommand(command Command) {
	s.mu.Lock()
	s.PendingCommands = append(s.PendingCommands, command)
	s.mu.Unlock()
}

// AddCommands adds one more commands to the commands to be executed
func (s *State) AddCommands(commands []Command) {
	s.mu.Lock()
	for _, command := range commands {
		s.PendingCommands = append(s.PendingCommands, command)
	}
	s.mu.Unlock()
}

func init() {
	commandState = State{map[string]Command{}, []Command{}, Executor{
		func(order Order) bool {
			//TODO
			return true
		},
		func(command Command) {
			commandState.AddCommand(command)
		},
		func(id string) bool {
			if _, ok := commandState.ExecutedCommands[id]; ok {
				return true
			}
			return false
		},
		func() int64 { return time.Now().Unix() },
	},
		sync.Mutex{},
	}

	go executeLoop()
}

func executeLoop() {
	for {
		executed := false
		commandState.mu.Lock()
		for i, cmd := range commandState.PendingCommands {
			if commandState.executor.Execute(cmd) {

				commandState.ExecutedCommands[cmd.ID] = cmd
				commandState.PendingCommands = append(commandState.PendingCommands[:i], commandState.PendingCommands[i+1:]...)

				executed = true
			}
		}
		commandState.mu.Unlock()

		if !executed {
			time.Sleep(1 * time.Second)
		}
	}
}

// GetCommandState returns the singleton local state of Genesis.
func GetCommandState() *State {
	return &commandState
}
