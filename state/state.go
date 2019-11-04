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

//Package state contains functionality for use by programs which wish to use Genesis in standalone mode.
package state

import (
	"context"
	"github.com/golang-collections/go-datastructures/queue"
	log "github.com/sirupsen/logrus"
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

//NewState creates a new state object and initializes the values
func NewState(exec command.Executor) *State {
	return &State{
		ExecutedCommands: map[string]command.Command{},
		pending:          queue.New(20),
		executor:         exec,
		once:             &sync.Once{},
	}
}

var commandState *State

// AddCommands adds one or more commands to be executed
func (s *State) AddCommands(commands ...command.Command) {
	for _, cmd := range commands {
		s.pending.Put(cmd)
	}
}

//Start starts the states inner consume loop
func (s *State) Start() {
	s.once.Do(func() {
		s.loop()
	})
}

//HasExecuted checks if there is a command with the given id in this objects map of executed commands
func (s *State) HasExecuted(id string) (ok bool) {
	s.mu.Lock()
	defer s.mu.Lock()
	_, ok = commandState.ExecutedCommands[id]
	return
}

func (s *State) loop() {
	for {
		cmds, err := s.pending.Get(1) //waits for new commands
		if err != nil {
			panic(err)
		}

		cmd := cmds[0].(command.Command)
		log.WithFields(log.Fields{"command": cmd}).Trace("attempting to run a command")
		s.executor.RunAsync(cmd, func(cmd command.Command, stat command.Result) {
			if !stat.IsSuccess() {
				s.AddCommands(cmd)
			} else {
				s.mu.Lock()
				defer s.mu.Lock()
				s.ExecutedCommands[cmd.ID] = cmd
			}
		})
	}
}

//Start causes the default command state to run its main loop, processing commands given to it.
func Start() {
	commandState = NewState(command.Executor{
		func(ctx context.Context, order command.Order) command.Result {
			//TODO
			return command.Result{Type: command.SuccessType, Error: nil}
		},
		func(cmd command.Command) {
			commandState.AddCommands(cmd)
		},
		func(id string) bool {
			return commandState.HasExecuted(id)
		},
		func() int64 { return time.Now().Unix() },
	})

	go commandState.Start()
}

// GetCommandState returns the singleton local state of Genesis.
func GetCommandState() *State {
	return commandState
}
