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
	"testing"

	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
)

func TestContainer_GetMemory_Successful(t *testing.T) {
	var tests = []struct {
		res      Container
		expected int64
	}{
		{res: Container{
			Cpus:   "",
			Memory: "45",
		}, expected: int64(45)},
		{res: Container{
			Cpus:   "",
			Memory: "1",
		}, expected: int64(1)},
		{res: Container{
			Cpus:   "",
			Memory: "92233720368547",
		}, expected: int64(92233720368547)},
		{res: Container{
			Cpus:   "",
			Memory: "3gb",
		}, expected: int64(3000000000)},
		{res: Container{
			Cpus:   "",
			Memory: "6KB",
		}, expected: int64(6000)},
		{res: Container{
			Cpus:   "",
			Memory: "4mb",
		}, expected: int64(4000000)},
		{res: Container{
			Cpus:   "",
			Memory: "1tb",
		}, expected: int64(1000000000000)},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			num, err := tt.res.GetMemory()
			assert.NoError(t, err)

			assert.Equal(t, num, tt.expected)
		})
	}
}

func TestContainer_GetMemory_Unsuccessful(t *testing.T) {
	var tests = []struct {
		res Container
	}{
		{res: Container{
			Cpus:   "",
			Memory: "45.46",
		}},
		{res: Container{
			Cpus:   "",
			Memory: "35273409857203948572039458720349857",
		}},
		{res: Container{
			Cpus:   "",
			Memory: "s",
		}},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := tt.res.GetMemory()
			assert.Error(t, err)
		})
	}
}

func TestContainer_NoLimits(t *testing.T) {
	var tests = []struct {
		res      Container
		expected bool
	}{
		{
			res: Container{
				Memory: "",
				Cpus:   "",
			},
			expected: true,
		},
		{
			res: Container{
				Memory: "5gb",
				Cpus:   "",
			},
			expected: false,
		},
		{
			res: Container{
				Memory: "",
				Cpus:   "5",
			},
			expected: false,
		},
		{
			res: Container{
				Memory: "4gb",
				Cpus:   "6",
			},
			expected: false,
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.res.NoLimits(), tt.expected)
		})
	}
}

func TestContainer_NoCPULimits(t *testing.T) {
	var tests = []struct {
		res      Container
		expected bool
	}{
		{
			res: Container{
				Cpus: "",
			},
			expected: true,
		},
		{
			res: Container{
				Cpus: "5",
			},
			expected: false,
		},
		{
			res: Container{
				Cpus: " ",
			},
			expected: false,
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.res.NoCPULimits(), tt.expected)
		})
	}
}

func TestContainer_NoMemoryLimits(t *testing.T) {
	var tests = []struct {
		res      Container
		expected bool
	}{
		{
			res: Container{
				Memory: "",
			},
			expected: true,
		},
		{
			res: Container{
				Memory: " ",
			},
			expected: false,
		},
		{
			res: Container{
				Memory: "5GB",
			},
			expected: false,
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.res.NoMemoryLimits(), tt.expected)
		})
	}
}

func TestContainer_GetEnv(t *testing.T) {
	var tests = []struct {
		c        Container
		expected []string
	}{
		{
			c: Container{
				Environment: map[string]string{"test": "env"},
			},
			expected: []string{"test=env"},
		},
		{
			c: Container{
				Environment: map[string]string{},
			},
			expected: []string{},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.ElementsMatch(t, tt.c.GetEnv(), tt.expected)
		})
	}
}

func TestContainer_GetPortBindings(t *testing.T) {
	var tests = []struct {
		c               Container
		expectedPortSet nat.PortSet
		expectedPortMap nat.PortMap
	}{
		{
			c: Container{
				Ports: nil,
			},
			expectedPortSet: nil,
			expectedPortMap: nil,
		},
		{
			c: Container{
				Ports: map[int]int{4000: 3000, 8000: 4444},
			},
			expectedPortSet: map[nat.Port]struct{}{"3000/tcp": struct{}{}, "4444/tcp": struct{}{}},
			expectedPortMap: map[nat.Port][]nat.PortBinding{"3000/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "4000"}}, "4444/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "8000"}}},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			portSet, portMap, err := tt.c.GetPortBindings()
			assert.NoError(t, err)

			assert.Equal(t, portSet, tt.expectedPortSet)
			assert.Equal(t, portMap, tt.expectedPortMap)
		})
	}
}

func TestContainer_GetEntryPoint(t *testing.T) {
	var tests = []struct {
		c        Container
		expected strslice.StrSlice
	}{
		{
			c: Container{
				EntryPoint: "",
			},
			expected: nil,
		},
		{
			c: Container{
				EntryPoint: "/test/path",
				Args:       []string{"test", "arguments", "as", "flags"},
			},
			expected: strslice.StrSlice{"/test/path", "test", "arguments", "as", "flags"},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.c.GetEntryPoint(), tt.expected)
		})
	}
}
