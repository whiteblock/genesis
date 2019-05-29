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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/whiteblock/genesis/net/mocks"
)

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
