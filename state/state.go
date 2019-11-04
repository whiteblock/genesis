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

package state

import (
	"github.com/golang-collections/go-datastructures/queue"
	"github.com/whiteblock/genesis/command"
	"sync"
	"time"
)

// State represents the local state of Genesis.
type State struct {
	ExecutedCommands map[string]command.Command
	pending          *queue.Queue
	executor         command.Executor
	mu               sync.Mutex
	once             *sync.Once
}

func NewState(exec Executor) *State {
	return &State{
		ExecutedCommands: map[string]command.Command{},
		pending:          queue.New(20),
		executor:         exec,
		once:             &sync.Once{},
	}
}

var commandState *State

// Addcommand.Commands adds one more commands to the commands to be executed
func (s *State) AddCommands(commands ...command.Command) {
	for _, cmd := range commands {
		s.pending.Put(command)
	}
}

//Start starts the states inner consume loop
func (s *State) Start() {
	s.once.Do(func() {
		s.loop()
	})
}

func (s *State) loop() {
	for {
		cmds, err := s.pending.Get(1) //waits for new commands
		if err != nil {
			panic(err)
		}
		cmd := cmds[0].(command.Command)
		s.executor.RunAsync(cmd, func(cmd command.Command, stat Status) {
			if !stat.IsSuccess() {
				s.Addcommand.Commands(command)
			} else {
				s.mu.Lock()
				defer s.mu.Lock()
				s.Executedcommand.Commands[cmd.ID] = cmd
			}
		})
	}
}

func init() {
	commandState = NewState(Executor{
		func(order Order) bool {
			//TODO
			return true
		},
		func(cmd command.Command) {
			commandState.AddCommands(command)
		},
		func(id string) bool {
			if _, ok := commandState.Executedcommand.Commands[id]; ok {
				return true
			}
			return false
		},
		func() int64 { return time.Now().Unix() },
	})

	go commandState.Start()
}

// Getcommand.CommandState returns the singleton local state of Genesis.
func GetCommandState() *State {
	return commandState
}
