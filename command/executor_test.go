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
	"testing"
	"time"
)

func TestOneCommand(t *testing.T) {

	executor := Executor{func(order Order) bool { return true },
		func(command Command) {},
		func(id string) bool { return true },
		func() int64 { return time.Now().Unix() },
	}

	executed := executor.Execute(Command{"1", 123, 0, Target{"testnetId"}, []string{}, Order{"status", "something"}})

	if !executed {
		t.Fatal("Should have run the command")
	}
}

func TestOneCommandForLater(t *testing.T) {

	executor := Executor{func(order Order) bool { return true },
		func(command Command) {},
		func(id string) bool { return true },
		func() int64 { return 122 },
	}

	executed := executor.Execute(Command{"1", 123, 0, Target{"testnetId"}, []string{}, Order{"status", "something"}})

	if executed {
		t.Fatal("Should not have run the command")
	}
}

func TestDependenciesCheck(t *testing.T) {

	executor := Executor{func(order Order) bool { return true },
		func(command Command) {},
		func(id string) bool { return false },
		func() int64 { return time.Now().Unix() },
	}

	executed := executor.Execute(Command{"1", 123, 0, Target{"testnetId"}, []string{"someDependency"}, Order{"status", "something"}})

	if executed {
		t.Fatal("Should not have run the command")
	}
}

func TestDependenciesSelection(t *testing.T) {

	executor := Executor{func(order Order) bool { return true },
		func(command Command) {},
		func(id string) bool { return id == "someOtherDependency" },
		func() int64 { return time.Now().Unix() },
	}

	executed := executor.Execute(Command{"1", 123, 0, Target{"testnetId"}, []string{"someDependency"}, Order{"status", "something"}})

	if executed {
		t.Fatal("Should not have run the command")
	}

	executed = executor.Execute(Command{"1", 123, 0, Target{"testnetId"}, []string{"someOtherDependency"}, Order{"status", "something"}})

	if !executed {
		t.Fatal("Should have run the command")
	}
}

func TestRescheduleOnFailure(t *testing.T) {
	var rescheduled Command
	executor := Executor{func(order Order) bool { return false },
		func(command Command) { rescheduled = command },
		func(id string) bool { return false },
		func() int64 { return 122 },
	}

	executed := executor.Execute(Command{"1", 121, 0, Target{"testnetId"}, []string{}, Order{"status", "something"}})

	if !executed {
		t.Fatal("Should have run the command")
	}

	if rescheduled.Timestamp != 122+10 {
		t.Fatalf("Did not reschedule 10 seconds later: %d", rescheduled.Timestamp)
	}

	if rescheduled.Retry != 1 {
		t.Fatal("Did not increment retry counter")
	}
}

func TestTooManyFailures(t *testing.T) {
	var rescheduled *Command
	executor := Executor{func(order Order) bool { return false },
		func(command Command) { rescheduled = &command },
		func(id string) bool { return false },
		func() int64 { return 122 },
	}

	executed := executor.Execute(Command{"1", 121, 4, Target{"testnetId"}, []string{}, Order{"status", "something"}})

	if !executed {
		t.Fatal("Should have run the command")
	}

	if rescheduled != nil {
		t.Fatal("Should not have rescheduled")
	}
}
