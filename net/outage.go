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
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/util"
	"strconv"
	"strings"
	"sync"
)

//RemoveAllOutages removes all blocked connections on a server via the given client
func RemoveAllOutages(client ssh.Client) error {
	res, err := client.Run("sudo iptables --list-rules | grep wb_bridge | grep DROP | grep FORWARD || true")
	if err != nil {
		return util.LogError(err)
	}
	if len(res) == 0 {
		return nil
	}
	res = strings.Replace(res, "-A ", "", -1)
	cmds := strings.Split(res, "\n")
	wg := sync.WaitGroup{}

	for _, cmd := range cmds {
		if len(cmd) == 0 {
			continue
		}
		wg.Add(1)
		go func(cmd string) {
			defer wg.Done()
			_, err := client.Run(fmt.Sprintf("sudo iptables -D %s", cmd))
			if err != nil {
				log.Error(err)
			}
		}(cmd)
	}

	wg.Wait()
	return nil
}

func makeOutageCommands(node1 db.Node, node2 db.Node) []string {
	return []string{
		fmt.Sprintf("FORWARD -i %s%d -d %s -j DROP", conf.BridgePrefix, node1.AbsoluteNum, node2.IP),
		fmt.Sprintf("FORWARD -i %s%d -d %s -j DROP", conf.BridgePrefix, node2.AbsoluteNum, node1.IP),
	}
}

//GetCutConnections fetches the cut connections on a server
//TODO: Naive Implementation, does not yet take multiple servers into account
func GetCutConnections(client ssh.Client) ([]Connection, error) {
	res, err := client.Run("sudo iptables --list-rules | grep wb_bridge | grep DROP | grep FORWARD | awk '{print $4,$6}' | sed -e 's/\\/32//g' || true")
	if err != nil {
		return nil, util.LogError(err)
	}
	out := []Connection{}
	if len(res) == 0 { //No cut connections on this server
		return out, nil
	}

	cuts := strings.Split(res, "\n")

	for _, cut := range cuts {
		if len(cut) == 0 {
			continue
		}
		cutPair := strings.Split(cut, " ")
		if len(cutPair) != 2 {
			return nil, fmt.Errorf("unexpected result \"%s\" for cut pair", cut)
		}
		_, toNode, _ := util.GetInfoFromIP(cutPair[0])

		if len(cutPair[1]) <= len(conf.BridgePrefix) {
			return nil, fmt.Errorf("unexpected source interface, found \"%s\"", cutPair[1])
		}

		fromNode, err := strconv.Atoi(cutPair[1][len(conf.BridgePrefix):])
		if err != nil {
			return nil, util.LogError(err)
		}
		out = append(out, Connection{To: toNode, From: fromNode})
		log.WithFields(log.Fields{"to": toNode, "from": fromNode}).Debug("found a disconnection")
	}
	return out, nil
}
