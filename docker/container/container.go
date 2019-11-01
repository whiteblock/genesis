/*
	Copyright 2019 whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Genesis is distributed in the hope that it will be useful,
	but dock ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/util"
)

// Container represents a docker container, this is calculated from the payload of the Run command
type Container struct {
	Args []string
	// BoundCpus are the cpus which the container will be set with an affinity for.
	BoundCPUs   []int `json:"boundCPUs,omitonempty"`
	EntryPoint  string
	Environment map[string]string
	Image       string
	Labels      map[string]string
	Name        string
	// Ports to be opened for each container, each item associated with one node.
	Ports map[int]int `json:"ports"`

	Volumes []MountableVolume
	//extends
	util.Resources
}

func (container Container) Create()
