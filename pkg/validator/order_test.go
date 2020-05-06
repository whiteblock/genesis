/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/whiteblock/definition/command"
)

func TestOrderValidator_ValidateContainer_Success(t *testing.T) {
	testContainer := command.Container{
		Name:     "t",
		TCPPorts: map[int]int{8000: 8000},
		Cpus:     "2.0",
		Memory:   "2GB",
		Image:    "t",
	}
	assert.NoError(t, Container(testContainer))
}

func TestOrderValidator_ValidateContainer_BadName(t *testing.T) {
	testContainer := command.Container{
		Name:     "",
		TCPPorts: map[int]int{8000: 8000},
		Cpus:     "2.0",
		Memory:   "2GB",
		Image:    "t",
	}
	assert.Error(t, Container(testContainer))
}

func TestOrderValidator_ValidateContainer_BadCPUs(t *testing.T) {
	testContainer := command.Container{
		Name:     "t",
		TCPPorts: map[int]int{8000: 8000},
		Cpus:     "fdsfsfsfe",
		Memory:   "2GB",
		Image:    "t",
	}
	assert.Error(t, Container(testContainer))
}

func TestOrderValidator_ValidateContainer_BadMem(t *testing.T) {
	testContainer := command.Container{
		Name:     "t",
		TCPPorts: map[int]int{8000: 8000},
		Cpus:     "2.0",
		Memory:   "fdwe2",
		Image:    "t",
	}
	assert.Error(t, Container(testContainer))
}

func TestOrderValidator_ValidateContainer_BadImage(t *testing.T) {
	testContainer := command.Container{
		Name:     "t",
		TCPPorts: map[int]int{8000: 8000},
		Cpus:     "2.0",
		Memory:   "2GB",
		Image:    "",
	}
	assert.Error(t, Container(testContainer))
}
