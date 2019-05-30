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

package netconf

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/whiteblock/genesis/net/mocks"
)

func TestCreateCommands(t *testing.T) {
	var test = []struct {
		netconf Netconf
		serverID int
		expected []string
	}{
		{
			netconf:  Netconf{Node: 3, Limit: 5, Loss: 0.06, Delay: 1, Rate: "", Duplication: 0.02, Corrupt: 0, Reorder: 0.01},
			serverID: 2,
			expected: []string {
				"sudo -n tc qdisc del dev wb_bridge3 root",
				"sudo -n tc qdisc add dev wb_bridge3 root handle 1: prio",
				"sudo -n tc qdisc add dev wb_bridge3 parent 1:1 handle 2: netem limit 5 loss 0.0600 delay 1us duplicate 0.0200 reorder 0.0100",
				"sudo -n tc filter add dev wb_bridge3 parent 1:0 protocol ip pref 55 handle 6 fw flowid 2:1",
				"sudo -n iptables -t mangle -A PREROUTING  ! -d 10.2.0.49 -j MARK --set-mark 6",
			},
		},
		{netconf: Netconf{Node: 3, Limit: 0, Loss: 0, Delay: 0, Rate: "0", Duplication: 0, Corrupt: 0, Reorder: 0},
			serverID: 3,
			expected: []string {
				"sudo -n tc qdisc del dev wb_bridge3 root",
				"sudo -n tc qdisc add dev wb_bridge3 root handle 1: prio",
				"sudo -n tc qdisc add dev wb_bridge3 parent 1:1 handle 2: netem rate 0",
				"sudo -n tc filter add dev wb_bridge3 parent 1:0 protocol ip pref 55 handle 6 fw flowid 2:1",
				"sudo -n iptables -t mangle -A PREROUTING  ! -d 10.3.0.49 -j MARK --set-mark 6",
			},
		},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if !reflect.DeepEqual(CreateCommands(tt.netconf, tt.serverID), tt.expected) {
				t.Errorf("return value of CreateCommands does not match expected value")
			}
		})
	}
}

func TestApply(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	netconf := Netconf{Node: 3}
	serverId := 1
	client := mocks.NewMockClient(ctrl)

	expectations := []string{
		"sudo -n tc qdisc del dev wb_bridge3 root",
		"sudo -n tc qdisc add dev wb_bridge3 root handle 1: prio",
		"sudo -n tc qdisc add dev wb_bridge3 parent 1:1 handle 2: netem",
		"sudo -n tc filter add dev wb_bridge3 parent 1:0 protocol ip pref 55 handle 6 fw flowid 2:1",
		"sudo -n iptables -t mangle -A PREROUTING  ! -d 10.1.0.49 -j MARK --set-mark 6",
	}

	var previous *gomock.Call
	for _, expectation := range expectations {
		call := client.EXPECT().Run(expectation)
		if previous != nil {
			call = call.After(previous)
		}

		previous = call
	}

	Apply(client, netconf, serverId)
}

//TODO: test RemoveAllOnServer()

func Test_parseItems(t *testing.T) {
	var test = []struct {
		items []string
		nconf *Netconf
		expected *Netconf
	}{
		{
			items: []string{"limit", "3", "test2", "test3"},
			nconf: &Netconf{Node: 2, Limit: 0, Loss: 0, Delay: 0, Rate: "0", Duplication: 0, Corrupt: 0, Reorder: 0},
			expected: &Netconf{Node: 2, Limit: 3, Loss: 0, Delay: 0, Rate: "0", Duplication: 0, Corrupt: 0, Reorder: 0},
		},
		{
			items: []string{"limit", "2", "loss", "0.5%", "delay", "415.9s", "rate", "2", "duplicate", "1.7%", "corrupt", "0.5%", "reorder", "0.07%"},
			nconf: &Netconf{Node: 3, Limit: 0, Loss: 0, Delay: 0, Rate: "0", Duplication: 0, Corrupt: 0, Reorder: 0},
			expected: &Netconf{Node: 3, Limit: 2, Loss: 0.5, Delay: 415900000, Rate: "2", Duplication: 1.7, Corrupt: 0.5, Reorder: 0.07},
		},
	}

	for i, tt := range test {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			parseItems(tt.items, tt.nconf)

			if !reflect.DeepEqual(&tt.nconf, &tt.expected) {
				t.Errorf("parseItems did not successfully change the contents of nconf")
			}
		})
	}
}

func TestGetConfigOnServer_Successful(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	out := []Netconf{
		{Node: 3, Limit: 2, Loss: 0.5, Delay: 0, Rate: "", Duplication: 0, Corrupt: 0, Reorder: 0},
		{Node: 4, Limit: 2, Loss: 1.3, Delay: 415900000, Rate: "1", Duplication: 0, Corrupt: 0.4, Reorder: 0.7},
	}

	client.
		EXPECT().
		Run("sudo -n tc qdisc show | grep wb_bridge | grep netem || true").
		Return("some random words testing wb_bridge3 test test limit 2 loss 0.5%\nsome random words testing wb_bridge4 test test limit 2 delay 415.9s loss 1.3% corrupt 0.4% rate 1 reorder 0.7% duplication 0", nil)

	netconf, _ := GetConfigOnServer(client)

	if !reflect.DeepEqual(netconf, out) {
		t.Errorf("return value of GetConfigOnServer does not match expected value")
	}
}

//TODO finish this test func
func TestGetConfigOnServer_Unsuccessful1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	out := []Netconf{
		{Node: 3, Limit: 0, Loss: 0, Delay: 0, Rate: "", Duplication: 0, Corrupt: 0, Reorder: 0},
	}

	client.
		EXPECT().
		Run("sudo -n tc qdisc show | grep wb_bridge | grep netem || true").
		Return("some random words testing wb_bridge3\nsome words random test\n", nil)

	netconf, _ := GetConfigOnServer(client)
	if !reflect.DeepEqual(netconf, out) {
		t.Errorf("return value of GetConfigOnServer does not match expected value")
	}
}

func TestGetConfigOnServer_Unsuccessful2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	out := []Netconf{}

	client.
		EXPECT().
		Run("sudo -n tc qdisc show | grep wb_bridge | grep netem || true").
		Return("", nil)

	netconf, _ := GetConfigOnServer(client)
	if !reflect.DeepEqual(netconf, out) {
		t.Errorf("return value of GetConfigOnServer does not match expected value")
	}
}
