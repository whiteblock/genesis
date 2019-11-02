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

package db

import (
	"github.com/whiteblock/genesis/util"
)

// ServiceDetails represent the data of a testnet service.
type ServiceDetails struct {
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Env     map[string]string `json:"env"`
	Network string            `json:"network"`
	Ports   []string          `json:"ports"`
	Volumes []string          `json:"volumes"`
}

// DeploymentDetails represents the data for the construction of a testnet.
type DeploymentDetails struct {
	// ID will be included when it is queried from the database.
	ID string `json:"id,omitempty"`

	// TestNetID: The id of the provisioned testnet
	TestNetID string `json:"testnetID"`

	// OrgID: The id of the organization
	OrgID string `json:"orgID"`

	// Servers: The ids of the servers to build on
	Servers []int `json:"servers"`

	// Blockchain: The blockchain to build.
	Blockchain string `json:"blockchain"`

	// Nodes:  The number of nodes to build
	Nodes int `json:"nodes"`

	// Image: The docker image to build off of
	Images []string `json:"images"`

	// Params: The blockchain specific parameters
	Params map[string]interface{} `json:"params"`

	// Resources: The resources per node
	Resources []util.Resources `json:"resources"`

	// Environments is the environment variables to be passed to each node.
	// If it doesn't exist for a node, it defaults first to index 0.
	Environments []map[string]string `json:"environments"`

	// Custom files for each node
	Files []map[string]string `json:"files"`

	// Logs to keep track of for each node
	Logs []map[string]string `json:"logs"`

	// Fairly Arbitrary extras for when additional customizations are added.
	Extras map[string]interface{} `json:"extras"`
	jwt    string
	kid    string

	Services []ServiceDetails `json:"services"`
}
