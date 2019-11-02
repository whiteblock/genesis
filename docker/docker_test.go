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

package docker

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/whiteblock/genesis/util"
)


func Test_dockerNetworkCreateCmd(t *testing.T) {
	var tests = []struct {
		subnet   string
		gateway  string
		network  int
		name     string
		expected string
	}{
		{
			subnet:  "blah",
			gateway: "blah",
			network: 0,
			name:    "blah",
			expected: fmt.Sprintf("docker network create --subnet %s --gateway %s -o \"com.docker.network.bridge.name=%s%d\" %s",
				"blah",
				"blah",
				conf.BridgePrefix,
				0,
				"blah"),
		},
		{
			subnet:  "test",
			gateway: "test",
			network: 1000,
			name:    "test",
			expected: fmt.Sprintf("docker network create --subnet %s --gateway %s -o \"com.docker.network.bridge.name=%s%d\" %s",
				"test",
				"test",
				conf.BridgePrefix,
				1000,
				"test"),
		},
		{
			subnet:  "",
			gateway: "",
			network: 0,
			name:    "",
			expected: fmt.Sprintf("docker network create --subnet %s --gateway %s -o \"com.docker.network.bridge.name=%s%d\" %s",
				"",
				"",
				conf.BridgePrefix,
				0,
				""),
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(dockerNetworkCreateCmd(tt.subnet, tt.gateway, tt.network, tt.name), tt.expected) {
				t.Error("return value of dockerNetworkCreateCmd does not match expected value")
			}
		})
	}
}

func Test_getFlagsFromResources(t *testing.T) {
	var tests = []struct {
		res      util.Resources
		expected string
	}{
		{
			res: util.Resources{
				Cpus:    "4",
				Memory:  "5GB",
				Volumes: []string{},
				Ports:   []string{},
			},
			expected: " --cpus 4 --memory 5000000000",
		},
		{
			res: util.Resources{
				Cpus:    "6",
				Memory:  "5MB",
				Volumes: []string{},
				Ports:   []string{},
			},
			expected: " --cpus 6 --memory 5000000",
		},
		{
			res: util.Resources{
				Cpus:    "2",
				Memory:  "10KB",
				Volumes: []string{},
				Ports:   []string{},
			},
			expected: " --cpus 2 --memory 10000",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			out, err := getFlagsFromResources(tt.res)
			if err != nil {
				t.Error("error running getFlagsFromResources", err)
			}

			if !reflect.DeepEqual(out, tt.expected) {
				t.Error("return value of getFlagsFromResources does not match expected value")
			}
		})
	}
}

func Test_dockerRunCmd(t *testing.T) {
	var tests = []struct {
		container Container
		expected  string
	}{
		{
			container: &ContainerDetails{
				Environment: map[string]string{},
				Image:       "test",
				Node:        0,
				Resources: util.Resources{
					Cpus:   "4",
					Memory: "5GB",
				},
				SubnetID:     10,
				NetworkIndex: 0,
				Type:         ContainerType(0),
				EntryPoint:   "/bin/sh",
				Args:         []string{},
			},
			expected: "docker run -itd --entrypoint /bin/sh --network wb_vlan0  --cpus 4 --memory 5000000000 --ip 10.10.0.2 --hostname whiteblock-node0 --name whiteblock-node0 test",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			out, err := dockerRunCmd(tt.container)
			if err != nil {
				t.Error("error running dockerRunCmd", err)
			}

			if out != tt.expected {
				t.Error("return value of dockerRunCmd does not match expected value")
			}
		})
	}
}

func Test_serviceDockerRunCmd(t *testing.T) {
	var tests = []struct {
		network  string
		ip       string
		name     string
		env      map[string]string
		volumes  []string
		ports    []string
		image    string
		cmd      string
		expected string
	}{
		{
			network:  "10",
			ip:       "10.128.13.14",
			name:     "eos",
			env:      map[string]string{},
			volumes:  []string{},
			ports:    []string{"3000"},
			image:    "testImage",
			cmd:      "test",
			expected: "docker run -itd --network 10 --ip 10.128.13.14 --hostname eos --name eos -e \"BIND_ADDR=10.128.13.14\"  -p 3000  testImage test",
		},
		{
			network:  "0",
			ip:       "10.128.13.14",
			name:     "geth",
			env:      map[string]string{},
			volumes:  []string{},
			ports:    []string{},
			image:    "testImage",
			cmd:      "test",
			expected: "docker run -itd --network 0 --ip 10.128.13.14 --hostname geth --name geth -e \"BIND_ADDR=10.128.13.14\"   testImage test",
		},
		{
			network:  "10",
			ip:       "10.128.01.01",
			name:     "artemis",
			env:      map[string]string{},
			volumes:  []string{},
			ports:    []string{"4444"},
			image:    "testImage",
			cmd:      "test",
			expected: "docker run -itd --network 10 --ip 10.128.01.01 --hostname artemis --name artemis -e \"BIND_ADDR=10.128.01.01\"  -p 4444  testImage test",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			out := serviceDockerRunCmd(tt.network, tt.ip, tt.name, tt.env, tt.volumes, tt.ports, tt.image, tt.cmd)

			if out != tt.expected {
				t.Error("return value of serviceDockerRunCmd does not match expected value")
			}

		})
	}
}
