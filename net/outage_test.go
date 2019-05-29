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

func TestRemoveAllOutages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)

	client.
		EXPECT().
		Run("sudo iptables --list-rules | grep wb_bridge | grep DROP | grep FORWARD || true").
		Return("sudo blah -A test test\nsudo this is a test\nsudo this is another test for the loop\nsudo test\n\n", nil)

	expectations := []string{
		"sudo iptables -D sudo blah test test",
		"sudo iptables -D sudo this is a test",
		"sudo iptables -D sudo this is another test for the loop",
		"sudo iptables -D sudo test",
	}

	for _, expectation := range expectations {
		 client.EXPECT().Run(expectation)
	}

	RemoveAllOutages(client)
}
