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

package status

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/util"
	"strconv"
	"strings"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

// Comp represents the compuational resources currently in use
// by a node
type Comp struct {
	CPU float64 `json:"cpu"`
	VSZ float64 `json:"virtualMemorySize"`
	RSS float64 `json:"residentSetSize"`
}

// NodeStatus represents the status of the node
type NodeStatus struct {
	Name         string            `json:"name"`
	Server       int               `json:"server"`
	IP           string            `json:"ip"`
	Up           bool              `json:"up"`
	Resources    Comp              `json:"resourceUse"`
	ID           string            `json:"id"`
	Protocol     string            `json:"protocol"`
	Image        string            `json:"image"`
	PortMappings map[string]string `json:"portMappings,omitonempty"`
	Timestamp    int64             `json:"timestamp"`
}

// FindNodeIndex finds the index of a node by name and server id
func FindNodeIndex(status []NodeStatus, name string, serverID int) int {
	for i, stat := range status {
		if stat.Name == name && serverID == stat.Server {
			return i
		}
	}
	return -1
}

// SumResUsage gets the cpu usage of a node
func SumResUsage(c ssh.Client, name string) (Comp, error) {
	res, err := c.Run(fmt.Sprintf("docker exec %s ps aux --no-headers | awk '{print $3,$5,$6}'", name))
	if err != nil {
		return Comp{-1, -1, -1}, util.LogError(err)
	}
	procs := strings.Split(res, "\n")
	log.WithFields(log.Fields{"name": name, "nprocs": len(res)}).Trace("found processes")
	var out Comp
	for _, proc := range procs {
		if len(proc) == 0 {
			continue
		}
		values := strings.Split(proc, " ")

		cpu, err := strconv.ParseFloat(values[0], 64)
		if err != nil {
			return Comp{-1, -1, -1}, util.LogError(err)
		}
		out.CPU += cpu

		vsz, err := strconv.ParseFloat(values[1], 64)
		if err != nil {
			return Comp{-1, -1, -1}, util.LogError(err)
		}
		out.VSZ += vsz

		rss, err := strconv.ParseFloat(values[2], 64)
		if err != nil {
			return Comp{-1, -1, -1}, util.LogError(err)
		}
		out.RSS += rss

	}
	return out, nil
}
