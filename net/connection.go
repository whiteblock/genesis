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

//Connection represents a uni-directional connection
type Connection struct {
	To   int `json:"to"`
	From int `json:"from"`
}

//Connections represents a graph of the network connections
type Connections struct {
	cons [][]bool //[from][to]
}

//NewConnections creates a Connections object with all connections marked
//as open
func NewConnections(nodes int) *Connections {
	out := new(Connections)
	out.cons = make([][]bool, nodes)

	for i := range out.cons {
		out.cons[i] = make([]bool, nodes)
		for j := range out.cons[i] {
			out.cons[i][j] = true
		}
	}

	return out
}

//RemoveAll will mark all of the given connections as not
//connected
func (mesh *Connections) RemoveAll(conns []Connection) {
	for _, conn := range conns {
		mesh.cons[conn.From][conn.To] = false
	}
}

func findPossiblePeers(cons [][]bool, node int) []int {
	var out []int

	for i, con := range cons[node] {
		if node == i {
			continue
		}
		if con || cons[i][node] {
			out = append(out, i)
		}
	}

	return out
}

func filterPeers(peers []int, alreadyDone []int) []int {
	if len(alreadyDone) == 0 {
		return peers
	}

	var out []int

	for _, peer := range peers {
		if !containsPeer(alreadyDone, peer) {
			out = append(out, peer)
		}
	}

	return out
}

func mergeUniquePeers(peers1 []int, peers2 []int) []int {
	out := make([]int, len(peers1))
	copy(out, peers1)

	for _, peer := range peers2 {
		if !containsPeer(out, peer) {
			out = append(out, peer)
		}
	}

	return out
}

func containsPeer(peers []int, peer int) bool {
	for _, p := range peers {
		if peer == p {
			return true
		}
	}

	return false
}

//Networks calculates the distinct, completely separate partitions in the network
func (mesh Connections) Networks() [][]int {
	nodes := []int{}
	nodesFinalized := []int{}
	nodesToTry := []int{}

	out := [][]int{}

	for len(nodesFinalized) < len(mesh.cons) {
		//fmt.Printf("\n\nNodes : %#v\n Nodes Finalized: %#v\nNodes To Try%#v\n\n",nodes,nodesFinalized,nodesToTry)
		if len(nodesToTry) == 0 {
			if len(nodes) > 0 {
				nodesFinalized = mergeUniquePeers(nodes, nodesFinalized)
				out = append(out, nodes)
				nodes = []int{}
			}
			for i := 0; i < len(mesh.cons); i++ {
				if !containsPeer(nodesFinalized, i) {
					nodesToTry = findPossiblePeers(mesh.cons, i)
					nodes = []int{i}
					nodes = mergeUniquePeers(nodes, nodesToTry)
					break
				}
			}

		} else {
			newPeers := findPossiblePeers(mesh.cons, nodesToTry[0])
			nodes = mergeUniquePeers(nodes, []int{nodesToTry[0]})
			newPeers = filterPeers(newPeers, nodes)
			if len(nodesToTry) > 1 {
				nodesToTry = nodesToTry[1:]
			} else {
				nodesToTry = []int{}
			}
			nodesToTry = mergeUniquePeers(nodesToTry, newPeers)

		}
	}
	return out
}
