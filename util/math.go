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

package util

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

//Point represents a basic 2D coordinate point
type Point struct {
	//X represents the position on the x axis
	X int `json:"x"`
	//Y represents the position on the y axis
	Y int `json:"y"`
}

// Distances creates a distance matrix, of all the distances between the given points
func Distances(pnts []Point) [][]float64 {
	out := make([][]float64, len(pnts))
	for i := range pnts {
		out[i] = make([]float64, len(pnts))
	}

	for i, ipnt := range pnts {
		for j, jpnt := range pnts {
			if j == i {
				continue
			}
			diffX := math.Abs(float64(ipnt.X - jpnt.X))
			diffY := math.Abs(float64(ipnt.Y - jpnt.Y))
			out[i][j] = math.Sqrt(math.Pow(diffX, 2) + math.Pow(diffY, 2))
		}
	}
	return out
}

// Distribute generates a roughly uniform random distribution for connections
// among nodes.
func Distribute(nodes []string, dist []int) ([][]string, error) {
	if len(nodes) < 2 {
		return nil, fmt.Errorf("cannot distribute a series smaller than 1")
	}
	for _, d := range dist {
		if d >= len(nodes) {
			return nil, fmt.Errorf("cannot distribute among more nodes than those that are provided")
		}
	}
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	out := [][]string{}
	for i := range nodes {
		conns := []string{}

		DISTRIBUTE:
		for j := 0; j < dist[i]; j++ {
			newConnIndex := r1.Intn(len(nodes))
			if newConnIndex == i {
				j--
				continue
			}
			newConn := nodes[newConnIndex]
			for _, conn := range conns {
				if newConn == conn {
					j--
					continue DISTRIBUTE
				}
			}

			conns = append(conns, newConn)

		}
		out = append(out, conns)
	}
	return out, nil
}

// GenerateWorstCaseNetwork generates a random path through all nodes
func GenerateWorstCaseNetwork(nodes int) [][]int {
	return generateWorstCaseNetwork(nodes, time.Now().UnixNano())
}

// private test function of exported function GenerateWorstCaseNetwork()
func generateWorstCaseNetwork(nodes int, seed int64) [][]int {
	out := make([][]int, nodes)

	s1 := rand.NewSource(seed)
	rng := rand.New(s1)

	nodePool := make([]int, nodes)
	for i := 0; i < nodes; i++ {
		nodePool[i] = i
	}
	node := 0
	for i := 0; i < nodes; i++ {
		newNodeIndex := rng.Intn(len(nodePool))
		newNode := nodePool[newNodeIndex]
		out[node] = []int{newNode}
		node = newNode
		nodePool = append(nodePool[:newNodeIndex], nodePool[newNodeIndex+1:]...)
	}
	out[node] = []int{0}
	return out
}

// GenerateUniformRandMeshNetwork generates a random mesh network that ensures
// that there is always a path between all the nodes
func GenerateUniformRandMeshNetwork(nodes int, conns int) ([][]int, error) {
	return generateUniformRandMeshNetwork(nodes, conns, time.Now().UnixNano())
}

// private func for GenerateUniformRandMeshNetwork for testing purposes
func generateUniformRandMeshNetwork(nodes int, conns int, seed int64) ([][]int, error) {
	if conns < 1 {
		return nil, fmt.Errorf("each node must have at least one connection")
	}
	if conns >= nodes {
		return nil, fmt.Errorf("too many connection to distribute without duplicates")
	}
	s1 := rand.NewSource(seed)
	rng := rand.New(s1)
	out := generateWorstCaseNetwork(nodes, seed)

	for i := 0; i < nodes; i++ {
		for j := 1; j < conns; j++ {
			node := rng.Intn(nodes)
			add := true
			for _, node2 := range out[i] {
				if node == node2 {
					j--
					add = false
					break
				}
			}
			if node == i {
				j--
				continue
			}
			if add {
				out[i] = append(out[i], node)
			}
		}
	}
	return out, nil
}

// GenerateNoDuplicateMeshNetwork generates a random mesh network that ensures
// that peering there is always a path between all the nodes, without any duplication.
// That is, if 1 contains 3, 3 won't contain 1 by elimination
func GenerateNoDuplicateMeshNetwork(nodes int, conns int) ([][]int, error) {
	return generateNoDuplicateMeshNetwork(nodes, conns, time.Now().UnixNano())
}

// private func of GenerateNoDuplicateMeshNetwork for testing purposes
func generateNoDuplicateMeshNetwork(nodes int, conns int, seed int64) ([][]int, error) {
	out, err := generateUniformRandMeshNetwork(nodes, conns, seed)
	if err != nil {
		return nil, err
	}
	for i := range out {
		for j := 0; j < len(out[i]); j++ {
			remove := false
			for _, k := range out[out[i][j]] {
				if k == i {
					remove = true
					break
				}
			}
			if remove {
				out[i] = append(out[i][:j], out[i][j+1:]...)
			}
		}
	}
	return out, nil
}

// GenerateDependentMeshNetwork generates a random mesh network that ensures
// the if built in order, each node will be given a list of peers which is already up and running.
// Note: This means that the first node will have an empty list
func GenerateDependentMeshNetwork(nodes int, conns int) ([][]int, error) {
	if conns < 1 {
		return nil, fmt.Errorf("each node must have at least one connection")
	}
	if conns >= nodes {
		return nil, fmt.Errorf("too many connection to distribute without duplicates")
	}
	s1 := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(s1)
	out := make([][]int, nodes)
	nodeToEnsure := 0
	for i := 0; i < nodes; i++ {
		for j := 1; j <= conns && j <= i; j++ {
			var node int
			if nodeToEnsure < i {
				node = nodeToEnsure
				nodeToEnsure++
			} else {
				node = rng.Intn(i)
			}

			add := true
			for _, node2 := range out[i] {
				if node == node2 {
					j--
					add = false
					break
				}
			}
			if node == i {
				j--
				continue
			}
			if add {
				out[i] = append(out[i], node)
			}
		}
	}
	return out, nil
}
