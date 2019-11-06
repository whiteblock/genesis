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

package entity

import (
	"strconv"
	"strings"
)

// Resources represents the maximum amount of resources
// that a node can use.
type Resources struct {
	// Cpus should be a floating point value represented as a string, and
	// is  equivalent to the percentage of a single cores time which can be used
	// by a node. Can be more than 1.0, meaning the node can use multiple cores at
	// a time.
	Cpus string `json:"cpus"`
	// Memory supports values up to Terrabytes (tb). If the unit is omitted, then it
	// is assumed to be bytes. This is not case sensitive.
	Memory string `json:"memory"`
}

func memconv(mem string) (int64, error) {

	m := strings.ToLower(mem)

	var multiplier int64 = 1

	if strings.HasSuffix(m, "kb") || strings.HasSuffix(m, "k") {
		multiplier = 1000
	} else if strings.HasSuffix(m, "mb") || strings.HasSuffix(m, "m") {
		multiplier = 1000000
	} else if strings.HasSuffix(m, "gb") || strings.HasSuffix(m, "g") {
		multiplier = 1000000000
	} else if strings.HasSuffix(m, "tb") || strings.HasSuffix(m, "t") {
		multiplier = 1000000000000
	}

	i, err := strconv.ParseInt(strings.Trim(m, "bgkmt"), 10, 64)
	if err != nil {
		return -1, err
	}

	return i * multiplier, nil
}

// GetMemory gets the memory value as an integer.
func (res Resources) GetMemory() (int64, error) {
	return memconv(res.Memory)
}

// NoLimits checks if the resources object doesn't specify any limits
func (res Resources) NoLimits() bool {
	return len(res.Memory) == 0 && len(res.Cpus) == 0
}

// NoCPULimits checks if the resources object doesn't specify any cpu limits
func (res Resources) NoCPULimits() bool {
	return len(res.Cpus) == 0
}

// NoMemoryLimits checks if the resources object doesn't specify any memory limits
func (res Resources) NoMemoryLimits() bool {
	return len(res.Memory) == 0
}
