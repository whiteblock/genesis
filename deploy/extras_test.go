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

package deploy

import (
	"fmt"
	"testing"

	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/testnet"
)

func Test_handleDockerBuildRequest(t *testing.T) {
	//ctrl := gomock.NewController(t)
	//defer ctrl.Finish()

	dB := db.DeploymentDetails{}
	tn, _ := testnet.NewTestNet(dB, "1")

	tn.LDD.Blockchain = "geth"

	fmt.Println(tn.LDD.Blockchain)


	//prebuild := map[string]interface{}{"dockerfile": "/test_test/Dockerfile", "testTestTest": "doesn't matter"}
	//expected := fmt.Sprintf("docker build /tmp/test -t %s -f %s", imageName, path)

	//client := mocks.NewMockClient(ctrl)
	//client.EXPECT().Run()

}