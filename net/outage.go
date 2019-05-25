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
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/status"
	"github.com/whiteblock/genesis/util"
	"log"
	"strconv"
	"strings"
	"sync"
)

//RemoveAllOutages removes all blocked connections on a server via the given client
func RemoveAllOutages(client ssh.Client) error {
	res, err := client.Run("sudo iptables --list-rules | grep wb_bridge | grep DROP | grep FORWARD || true")
	if err != nil {
		log.Println(err)
		return err
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
			_, err = client.Run(fmt.Sprintf("sudo iptables -D %s", cmd))
			if err != nil {
				log.Println(err)
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

func mkrmOutage(node1 db.Node, node2 db.Node, create bool) error {
	flag := "-I"
	if !create {
		flag = "-D"
	}
	cmds := makeOutageCommands(node1, node2)

	client, err := status.GetClient(node1.Server)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = client.Run(fmt.Sprintf("sudo iptables %s %s", flag, cmds[0]))
	if err != nil {
		log.Println(err)
		return err
	}
	client, err = status.GetClient(node2.Server)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = client.Run(fmt.Sprintf("sudo iptables %s %s", flag, cmds[1]))
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

//MakeOutage removes the ability for the given nodes to connect
func MakeOutage(node1 db.Node, node2 db.Node) error {
	return mkrmOutage(node1, node2, true)
}

//RemoveOutage returns the ability for the given nodes to connect
func RemoveOutage(node1 db.Node, node2 db.Node) error {
	return mkrmOutage(node1, node2, false)
}

//CreatePartitionOutage causes the two sides to be unable to communicate with one and the other
func CreatePartitionOutage(side1 []db.Node, side2 []db.Node) { //Doesn't report errors yet
	wg := sync.WaitGroup{}
	for _, node1 := range side1 {
		for _, node2 := range side2 {
			wg.Add(1)
			go func(node1 db.Node, node2 db.Node) {
				defer wg.Done()
				err := MakeOutage(node1, node2)
				if err != nil {
					log.Println(err)
				}
			}(node1, node2)
		}
	}
	wg.Wait()
}

//GetCutConnections fetches the cut connections on a server
//TODO: Naive Implementation, does not yet take multiple servers into account
func GetCutConnections(client ssh.Client) ([]Connection, error) {
	res, err := client.Run("sudo iptables --list-rules | grep wb_bridge | grep DROP | grep FORWARD | awk '{print $4,$6}' | sed -e 's/\\/32//g' || true")
	if err != nil {
		log.Println(err)
		return nil, err
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
			log.Println(err)
			return nil, err
		}
		out = append(out, Connection{To: toNode, From: fromNode})
	}
	return out, nil
}

//CalculatePartitions calculates the current partitions in the network
func CalculatePartitions(nodes []db.Node) ([][]int, error) {
	clients, err := status.GetClientsFromNodes(nodes)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	cutConnections := []Connection{}
	for _, client := range clients {
		conns, err := GetCutConnections(client)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		cutConnections = append(cutConnections, conns...)
	}

	conns := NewConnections(len(nodes))

	conns.RemoveAll(cutConnections)

	return conns.Networks(), nil
}
