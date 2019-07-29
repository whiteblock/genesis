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

	"github.com/golang/mock/gomock"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/ssh/mocks"
	"github.com/whiteblock/genesis/testnet"
)

func CreateMockClient(t *testing.T) *mocks.MockClient {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	return mocks.NewMockClient(ctrl)
}

func TestKillNode(t *testing.T) {
	client := CreateMockClient(new(testing.T))

	expectation := fmt.Sprintf("docker rm -f %s%d", conf.NodePrefix, 0)
	client.EXPECT().Run(expectation)

	err := KillNode(client, 0)
	if err != nil {
		t.Error("return value of KillNode does not match expected value")
	}
}

func TestKill(t *testing.T) {
	client := CreateMockClient(new(testing.T))

	expectation := fmt.Sprintf("docker rm -f $(docker ps -aq -f name=\"%s%d\")", conf.NodePrefix, 0)
	client.EXPECT().Run(expectation)

	err := Kill(client, 0)
	if err != nil {
		t.Error("return value of Kill does not match expected value")
	}
}

func TestKillAll(t *testing.T) {
	client := CreateMockClient(new(testing.T))

	expectation := fmt.Sprintf("docker rm -f $(docker ps -aq -f name=\"%s\")", conf.NodePrefix)
	client.EXPECT().Run(expectation)

	err := KillAll(client)
	if err != nil {
		t.Error("return value of Kill does not match expected value")
	}
}

func Test_dockerNetworkCreateCmd(t *testing.T) {
	var tests = []struct {
		subnet string
		gateway string
		network int
		name string
		expected string
	}{
		{
			subnet: "blah",
			gateway: "blah",
			network: 0,
			name: "blah",
			expected: fmt.Sprintf("docker network create --subnet %s --gateway %s -o \"com.docker.network.bridge.name=%s%d\" %s",
				"blah",
				"blah",
				conf.BridgePrefix,
				0,
				"blah"),
		},
		{
			subnet: "test",
			gateway: "test",
			network: 1000,
			name: "test",
			expected: fmt.Sprintf("docker network create --subnet %s --gateway %s -o \"com.docker.network.bridge.name=%s%d\" %s",
				"test",
				"test",
				conf.BridgePrefix,
				1000,
				"test"),
		},
		{
			subnet: "",
			gateway: "",
			network: 0,
			name: "",
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

func TestNetworkCreate(t *testing.T) {
	client := CreateMockClient(new(testing.T))

	command := "docker network create --subnet 10.10.0.0/28 --gateway 10.10.0.1 -o \"com.docker.network.bridge.name=wb_bridge0\" wb_vlan0"
	client.EXPECT().Run(command)

	testNet := new(testnet.TestNet)
	testNet.Clients = map[int]ssh.Client{0: client}

	if err := NetworkCreate(testNet, 0, 10, 0); err != nil {
		t.Error("return value of NetworkCreate does not match expected value")
	}
}

