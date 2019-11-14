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
	"strconv"
	"testing"

	"github.com/docker/docker/api/types/mount"
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

			assert.Equal(t, tt.expected, num)
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
			assert.Equal(t, tt.expected, tt.res.NoLimits())
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
			assert.Equal(t, tt.expected, tt.res.NoCPULimits())
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
			assert.Equal(t, tt.expected, tt.res.NoMemoryLimits())
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
			assert.ElementsMatch(t, tt.expected, tt.c.GetEnv())
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

			assert.Equal(t, tt.expectedPortSet, portSet)
			assert.Equal(t, tt.expectedPortMap, portMap)
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
			assert.Equal(t, tt.expected, tt.c.GetEntryPoint())
		})
	}
}

func TestContainer_GetMounts(t *testing.T) {
	var tests = []struct {
		c        Container
		expected []mount.Mount
	}{
		{
			c:        Container{},
			expected: []mount.Mount{},
		},
		{
			c: Container{
				Volumes: []Mount{Mount{Name: "test", Directory: "/opt/whiteblock", ReadOnly: false}},
			},
			expected: []mount.Mount{mount.Mount{
				Type:     mount.TypeVolume,
				Source:   "test",
				Target:   "/opt/whiteblock",
				ReadOnly: false,
			}},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.c.GetMounts())
		})
	}
}
