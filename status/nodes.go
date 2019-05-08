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

package status

import (
	"../db"
	"../ssh"
	"../util"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
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
	Name      string `json:"name"`
	Server    int    `json:"server"`
	IP        string `json:"ip"`
	Up        bool   `json:"up"`
	Resources Comp   `json:"resourceUse"`
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
func SumResUsage(c *ssh.Client, name string) (Comp, error) {
	res, err := c.Run(fmt.Sprintf("docker exec %s ps aux --no-headers | grep -v nibbler | awk '{print $3,$5,$6}'", name))
	if err != nil {
		log.Println(err)
		return Comp{-1, -1, -1}, err
	}
	procs := strings.Split(res, "\n")
	//fmt.Printf("%#v\n", procs)
	var out Comp
	for _, proc := range procs {
		if len(proc) == 0 {
			continue
		}
		values := strings.Split(proc, " ")

		cpu, err := strconv.ParseFloat(values[0], 64)
		if err != nil {
			log.Println(err)
			return Comp{-1, -1, -1}, err
		}
		out.CPU += cpu

		vsz, err := strconv.ParseFloat(values[1], 64)
		if err != nil {
			log.Println(err)
			return Comp{-1, -1, -1}, err
		}
		out.VSZ += vsz

		rss, err := strconv.ParseFloat(values[2], 64)
		if err != nil {
			log.Println(err)
			return Comp{-1, -1, -1}, err
		}
		out.RSS += rss

	}
	return out, nil
}

// CheckNodeStatus checks the status of the nodes in the current testnet
func CheckNodeStatus(nodes []db.Node) ([]NodeStatus, error) {

	serverIDs := db.GetUniqueServerIDs(nodes)
	out := make([]NodeStatus, len(nodes))

	for _, node := range nodes {
		//fmt.Printf("ABS = %d; REL=%d;NAME=%s%d\n", node.AbsoluteNum, node.LocalID, conf.NodePrefix, node.LocalID)
		out[node.AbsoluteNum] = NodeStatus{
			Name:      fmt.Sprintf("%s%d", conf.NodePrefix, node.LocalID),
			IP:        node.IP,
			Server:    node.Server,
			Up:        false,
			Resources: Comp{-1, -1, -1},
		}

	}
	servers, err := db.GetServers(serverIDs)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	mux := sync.Mutex{}
	wg := sync.WaitGroup{}

	for _, server := range servers {
		client, err := GetClient(server.ID)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		res, err := client.Run(
			fmt.Sprintf("docker ps | egrep -o '%s[0-9]*' | sort", conf.NodePrefix))
		if err != nil {
			log.Println(err)
			return nil, err
		}
		names := strings.Split(res, "\n")
		for _, name := range names {
			if len(name) == 0 {
				continue
			}

			index := FindNodeIndex(out, name, server.ID)
			if index == -1 {
				log.Printf("name=\"%s\",server=%d\n", name, server.ID)
				continue
			}
			wg.Add(1)
			go func(client *ssh.Client, name string, index int) {
				defer wg.Done()
				resUsage, err := SumResUsage(client, name)
				if err != nil {
					log.Println(err)
				}
				mux.Lock()
				out[index].Up = true
				out[index].Resources = resUsage
				mux.Unlock()
			}(client, name, index)
		}
	}
	wg.Wait()
	return out, nil
}
