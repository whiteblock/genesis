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

package command

//Netconf represents network impairments which are to be applied to a network
type Netconf struct {
	OrderPayload
	//Container is the target container
	Container string `json:"container"`
	//Network is the target network
	Network string `json:"network"`
	//Limit is the max number of packets to hold the in queue
	Limit int `json:"limit"`
	//Loss represents packet loss % ie 100% = 100
	Loss float64 `json:"loss"`
	//Delay represents the latency to be applied in microseconds
	Delay int `json:"delay"`
	//Rate represents the bandwidth constraint to be applied to the network
	Rate string `json:"rate"`
	//Duplication represents the percentage of packets to duplicate
	Duplication float64 `json:"duplicate"`
	//Corrupt represents the percentage of packets to corrupt
	Corrupt float64 `json:"corrupt"`
	//Reorder represents the percentage of packets that get reordered
	Reorder float64 `json:"reorder"`
}
