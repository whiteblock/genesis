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

package netconf

import (
	"../util"
	"math"
)

//Calculator contains functions used to calculate network impairments based on a distance
type Calculator struct {
	Loss        func(float64) float64
	Delay       func(float64) int
	Rate        func(float64) string
	Duplication func(float64) float64
	Corrupt     func(float64) float64
	Reorder     func(float64) float64
}

//Link is a naive representation of a network link
type Link struct {
	EgressNode  int     `json:"egressNode"`  //redundant info
	IngressNode int     `json:"ingressNode"` //redundant info
	Loss        float64 `json:"loss"`        //Loss % ie 100% = 100
	Delay       int     `json:"delay"`
	Rate        string  `json:"rate"`
	Duplication float64 `json:"duplicate"`
	Corrupt     float64 `json:"corrupt"`
	Reorder     float64 `json:"reorder"`
}

//GetDefaultCalculator creates a calculator which can be used to calculate latency
func GetDefaultCalculator() *Calculator { //TODO: improve the equations
	return &Calculator{
		Loss: func(dist float64) float64 {
			if dist == 0 {
				return 0
			}
			return math.Min(math.Log(dist), 100.0)
		},
		Delay: func(dist float64) int {
			return int(dist * 10)
		},
		Rate: func(dist float64) string {
			return ""
		},
		Duplication: func(dist float64) float64 {
			if dist == 0 {
				return 0
			}
			return math.Min(math.Log10(dist), 100.0)
		},
		Corrupt: func(dist float64) float64 {
			if dist == 0 {
				return 0
			}
			return math.Min(math.Log10(dist), 100.0)
		},
		Reorder: func(dist float64) float64 {
			if dist == 0 {
				return 0
			}
			return math.Min(math.Log10(dist), 100.0)
		},
	}
}

// CreateLinks generates a naive representation of a mesh network, which includes
// basic network impairments based on the distance between the points
func CreateLinks(pnts []util.Point, c *Calculator) [][]Link {
	if c == nil {
		c = GetDefaultCalculator()
	}
	dists := util.Distances(pnts)
	out := make([][]Link, len(pnts))

	for i, distArr := range dists {
		for j, dist := range distArr {
			out[i] = append(out[i], Link{
				EgressNode:  i,
				IngressNode: j,
				Loss:        c.Loss(dist),
				Delay:       c.Delay(dist),
				Rate:        c.Rate(dist),
				Duplication: c.Duplication(dist),
				Corrupt:     c.Corrupt(dist),
				Reorder:     c.Reorder(dist),
			})

		}
	}
	return out
}
