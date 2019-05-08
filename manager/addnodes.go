/*
	Copyright 2019 Whiteblock Inc.
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

package manager

import (
	"../blockchains/registrar"
	"../db"
	"../deploy"
	"../state"
	"../testnet"
	"../util"
	"fmt"
)

// AddNodes allows for nodes to be added to the network.
// The nodes don't need to be of the same type of the original build.
// It is worth noting that any missing information from the given
// deployment details will be filled in from the origin build.
func AddNodes(details *db.DeploymentDetails, testnetID string) error {
	buildState, err := state.GetBuildStateByID(testnetID)
	if err != nil {
		return util.LogError(err)
	}

	tn, err := testnet.RestoreTestNet(testnetID)
	if err != nil {
		buildState.ReportError(err)
		return err
	}
	defer tn.FinishedBuilding()

	err = tn.AddDetails(*details)
	if err != nil {
		buildState.ReportError(err)
		return err
	}

	//STEP 2: VALIDATE
	for i, res := range details.Resources {
		err = res.ValidateAndSetDefaults()
		if err != nil {
			err = fmt.Errorf("%s. For node %d", err.Error(), i)
			buildState.ReportError(err)
			return err
		}
	}

	if details.Nodes > conf.MaxNodes {
		buildState.ReportError(fmt.Errorf("too many nodes"))
		return fmt.Errorf("too many nodes")
	}

	err = deploy.AddNodes(tn)
	if err != nil {
		buildState.ReportError(err)
		return err
	}

	addNodesFn, err := registrar.GetAddNodeFunc(details.Blockchain)
	if err != nil {
		buildState.ReportError(err)
		return err
	}
	err = addNodesFn(tn)
	if err != nil {
		buildState.ReportError(err)
		return err
	}

	err = handleSideCars(tn, true)
	if err != nil {
		buildState.ReportError(err)
		return err
	}

	err = tn.StoreNodes()
	if err != nil {
		buildState.ReportError(err)
		return err
	}

	return nil
}
