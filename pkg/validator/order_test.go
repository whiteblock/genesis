/*
	Copyright 2019 Whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Genesis is distributed in the hope that it will be useful,
	but dock ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/whiteblock/definition/command"
)

func TestOrderValidator_ValidateContainer_Success(t *testing.T) {
	testContainer := command.Container{
		Name:   "t",
		Ports:  map[int]int{8000: 8000},
		Cpus:   "2.0",
		Memory: "2GB",
		Image:  "t",
	}
	assert.NoError(t, Container(testContainer))
}

func TestOrderValidator_ValidateContainer_BadName(t *testing.T) {
	testContainer := command.Container{
		Name:   "",
		Ports:  map[int]int{8000: 8000},
		Cpus:   "2.0",
		Memory: "2GB",
		Image:  "t",
	}
	assert.Error(t, Container(testContainer))
}

func TestOrderValidator_ValidateContainer_BadPorts_Host(t *testing.T) {
	testContainer := command.Container{
		Name:   "t",
		Ports:  map[int]int{80000: 8000},
		Cpus:   "2.0",
		Memory: "2GB",
		Image:  "t",
	}
	assert.Error(t, Container(testContainer))
}

func TestOrderValidator_ValidateContainer_BadPorts_Container(t *testing.T) {
	testContainer := command.Container{
		Name:   "t",
		Ports:  map[int]int{8000: 80000},
		Cpus:   "2.0",
		Memory: "2GB",
		Image:  "t",
	}
	assert.Error(t, Container(testContainer))
}

func TestOrderValidator_ValidateContainer_BadCPUs(t *testing.T) {
	testContainer := command.Container{
		Name:   "t",
		Ports:  map[int]int{8000: 8000},
		Cpus:   "fdsfsfsfe",
		Memory: "2GB",
		Image:  "t",
	}
	assert.Error(t, Container(testContainer))
}

func TestOrderValidator_ValidateContainer_BadMem(t *testing.T) {
	testContainer := command.Container{
		Name:   "t",
		Ports:  map[int]int{8000: 8000},
		Cpus:   "2.0",
		Memory: "fdwe2",
		Image:  "t",
	}
	assert.Error(t, Container(testContainer))
}

func TestOrderValidator_ValidateContainer_BadImage(t *testing.T) {
	testContainer := command.Container{
		Name:   "t",
		Ports:  map[int]int{8000: 8000},
		Cpus:   "2.0",
		Memory: "2GB",
		Image:  "",
	}
	assert.Error(t, Container(testContainer))
}
