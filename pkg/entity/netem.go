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

package entity

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
