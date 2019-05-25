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

//Package netconf provides the basic functionality for the simulation of network conditions across nodes.
package netconf

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/status"
	"github.com/whiteblock/genesis/util"
	"regexp"
	"strconv"
	"strings"
)

/**
[ limit PACKETS ]
[ delay TIME [ JITTER [CORRELATION]]]
[ distribution {uniform|normal|pareto|paretonormal} ]
[ corrupt PERCENT [CORRELATION]]
[ duplicate PERCENT [CORRELATION]]
[ loss random PERCENT [CORRELATION]]
[ loss state P13 [P31 [P32 [P23 P14]]]
[ loss gemodel PERCENT [R [1-H [1-K]]]
[ ecn ]
[ reorder PRECENT [CORRELATION] [ gap DISTANCE ]]
[ rate RATE [PACKETOVERHEAD] [CELLSIZE] [CELLOVERHEAD]]
*/

var conf = util.GetConfig()

//Netconf is a representation of the impairments applied to a node
type Netconf struct {
	Node        int     `json:"node"`
	Limit       int     `json:"limit"`
	Loss        float64 `json:"loss"` //Loss % ie 100% = 100
	Delay       int     `json:"delay"`
	Rate        string  `json:"rate"`
	Duplication float64 `json:"duplicate"`
	Corrupt     float64 `json:"corrupt"`
	Reorder     float64 `json:"reorder"`
}

// CreateCommands generates the commands needed to obtain the desired
// network conditions
func CreateCommands(netconf Netconf, serverID int) []string {
	const offset int = 6
	out := []string{
		fmt.Sprintf("sudo -n tc qdisc del dev %s%d root", conf.BridgePrefix, netconf.Node),
		fmt.Sprintf("sudo -n tc qdisc add dev %s%d root handle 1: prio", conf.BridgePrefix, netconf.Node),
		fmt.Sprintf("sudo -n tc qdisc add dev %s%d parent 1:1 handle 2: netem", conf.BridgePrefix, netconf.Node), //unf
		fmt.Sprintf("sudo -n tc filter add dev %s%d parent 1:0 protocol ip pref 55 handle %d fw flowid 2:1",
			conf.BridgePrefix, netconf.Node, offset),
		fmt.Sprintf("sudo -n iptables -t mangle -A PREROUTING  ! -d %s -j MARK --set-mark %d",
			util.GetGateway(serverID, netconf.Node), offset),
	}

	if netconf.Limit > 0 {
		out[2] += fmt.Sprintf(" limit %d", netconf.Limit)
	}

	if netconf.Loss > 0 {
		out[2] += fmt.Sprintf(" loss %.4f", netconf.Loss)
	}

	if netconf.Delay > 0 {
		out[2] += fmt.Sprintf(" delay %dus", netconf.Delay)
	}

	if len(netconf.Rate) > 0 {
		out[2] += fmt.Sprintf(" rate %s", netconf.Rate)
	}

	if netconf.Duplication > 0 {
		out[2] += fmt.Sprintf(" duplicate %.4f", netconf.Duplication)
	}

	if netconf.Corrupt > 0 {
		out[2] += fmt.Sprintf(" corrupt %.4f", netconf.Duplication)
	}

	if netconf.Reorder > 0 {
		out[2] += fmt.Sprintf(" reorder %.4f", netconf.Reorder)
	}

	return out
}

//Apply applies the given network config.
func Apply(client ssh.Client, netconf Netconf, serverID int) error {
	cmds := CreateCommands(netconf, serverID)
	for i, cmd := range cmds {
		_, err := client.Run(cmd)
		if i == 0 {
			//Don't check the success of the first command which clears
			continue
		}
		if err != nil {
			return util.LogError(err)
		}
	}
	return nil
}

//ApplyAll applies all of the given netconfs
func ApplyAll(netconfs []Netconf, nodes []db.Node) error {
	for _, netconf := range netconfs {
		node, err := db.GetNodeByLocalID(nodes, netconf.Node)
		if err != nil {
			return util.LogError(err)
		}
		client, err := status.GetClient(node.Server)
		if err != nil {
			return util.LogError(err)
		}
		err = Apply(client, netconf, node.Server)
		if err != nil {
			return util.LogError(err)
		}
	}
	return nil
}

//ApplyToAll applies the given netconf to `nodes` nodes in the network on the given server
func ApplyToAll(netconf Netconf, nodes []db.Node) error {
	for _, node := range nodes {
		netconf.Node = node.LocalID
		cmds := CreateCommands(netconf, node.Server)
		for i, cmd := range cmds {
			client, err := status.GetClient(node.Server)
			if err != nil {
				log.WithFields(log.Fields{"i": i, "cmd": cmd, "error": err}).Error("error running netem command")
				return util.LogError(err)
			}
			_, err = client.Run(cmd)
			if i == 0 {
				//Don't check the success of the first command which clears
				continue
			}
			if err != nil {
				return util.LogError(err)
			}
		}
	}
	return nil
}

//RemoveAll removes network conditions from the given nodes
func RemoveAll(nodes []db.Node) error {
	for _, node := range nodes {
		client, err := status.GetClient(node.Server)
		if err != nil {
			return util.LogError(err)
		}
		_, err = client.Run(
			fmt.Sprintf("sudo -n tc qdisc del dev %s%d root", conf.BridgePrefix, node.LocalID))
		if err != nil {
			log.Error(err)
		}
	}
	return nil
}

//RemoveAllOnServer removes network conditions from the given number of nodes on the given client
func RemoveAllOnServer(client ssh.Client, nodes int) {
	for i := 0; i < nodes; i++ {
		client.Run(
			fmt.Sprintf("sudo tc qdisc del dev %s%d root", conf.BridgePrefix, i))
	}
	RemoveAllOutages(client)
}

func parseItems(items []string, nconf *Netconf) error {

	for i := 0; i < len(items)/2; i++ {
		switch items[2*i] {
		case "limit":
			val, err := strconv.Atoi(items[2*i+1])
			if err != nil {
				return util.LogError(err)
			}
			nconf.Limit = val
		case "loss":
			val, err := strconv.ParseFloat(items[2*i+1][:len(items[2*i+1])-1], 64)
			if err != nil {
				return util.LogError(err)
			}
			nconf.Loss = val
		case "delay":
			re := regexp.MustCompile(`(?m)[0-9]+\.[0-9]+`)
			matches := re.FindAllString(items[2*i+1], -1)
			if len(matches) == 0 {
				return fmt.Errorf("unexpected delay value \"%s\"", items[2*i+1])
			}

			val, err := strconv.ParseFloat(matches[0], 64)
			if err != nil {
				return util.LogError(err)
			}
			unit := items[2*i+1][len(matches[0]):]
			switch unit {
			case "s":
				val *= 1000
				fallthrough
			case "ms":
				val *= 1000
			}
			nconf.Delay = int(val)
		case "rate":
			nconf.Rate = items[2*i+1]
		case "duplicate":
			val, err := strconv.ParseFloat(items[2*i+1][:len(items[2*i+1])-1], 64)
			if err != nil {
				return util.LogError(err)
			}
			nconf.Duplication = val
		case "corrupt":
			val, err := strconv.ParseFloat(items[2*i+1][:len(items[2*i+1])-1], 64)
			if err != nil {
				return util.LogError(err)
			}
			nconf.Corrupt = val
		case "reorder":
			val, err := strconv.ParseFloat(items[2*i+1][:len(items[2*i+1])-1], 64)
			if err != nil {
				return util.LogError(err)
			}
			nconf.Reorder = val
		}
	}
	return nil
}

//GetConfigOnServer gets the network impairments present on a server
func GetConfigOnServer(client ssh.Client) ([]Netconf, error) {
	res, err := client.Run("sudo -n tc qdisc show | grep wb_bridge | grep netem || true")
	if err != nil {
		return nil, util.LogError(err)
	}
	if len(res) == 0 {
		return []Netconf{}, nil
	}
	out := []Netconf{}
	rawConfigs := strings.Split(res, "\n")

	for _, rawConfig := range rawConfigs { //4 for bridge name //7 for start of the shit
		rawItems := strings.Split(rawConfig, " ")
		if len(rawItems) < 5 {
			continue
		}
		bridgeName := rawItems[4]

		num, err := strconv.Atoi(bridgeName[len(conf.BridgePrefix):])
		if err != nil {
			return nil, util.LogError(err)
		}
		nconf := Netconf{Node: num}
		if len(rawItems) >= 8 {
			items := rawItems[7:]
			err = parseItems(items, &nconf)
			if err != nil {
				return nil, util.LogError(err)
			}
		}
		out = append(out, nconf)
	}
	return out, nil
}
