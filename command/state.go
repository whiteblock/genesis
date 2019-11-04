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
	"github.com/golang-collections/go-datastructures/queue"
	"sync"
	"time"
)

// State represents the local state of Genesis.
type State struct {
	ExecutedCommands map[string]Command
	pending          *queue.Queue
	executor         Executor
	mu               sync.Mutex
	once             *sync.Once
}

func NewState(exec Executor) *State {
	return &State{
		ExecutedCommands: map[string]Command{},
		pending:          queue.New(20),
		executor:         exec,
		once:             &sync.Once{},
	}
}

var commandState *State

// AddCommands adds one more commands to the commands to be executed
func (s *State) AddCommands(commands ...Command) {
	for _, command := range commands {
		s.pending.Put(command)
	}

}

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
		cmd := cmds[0].(Command)
		s.executor.RunAsync(cmd, func(command Command, success bool) {
			if !success {
				s.AddCommands(command)
			} else {
				s.mu.Lock()
				defer s.mu.Lock()
				s.ExecutedCommands[cmd.ID] = cmd
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
		func(command Command) {
			commandState.AddCommands(command)
		},
		func(id string) bool {
			if _, ok := commandState.ExecutedCommands[id]; ok {
				return true
			}
			return false
		},
		func() int64 { return time.Now().Unix() },
	})

	go commandState.Start()
}

// GetCommandState returns the singleton local state of Genesis.
func GetCommandState() *State {
	return commandState
}
