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

// Package status handles functions related to the current state of the network
package status

import (
	"github.com/Whiteblock/genesis/state"
	"log"
)

// CheckBuildStatus checks the current status of the build relating to the
// given build id
func CheckBuildStatus(buildID string) (string, error) {
	bs, err := state.GetBuildStateByID(buildID)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return bs.Marshal(), nil
}
