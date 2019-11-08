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

//VolumeConfig is the configuration options for volumes
type VolumeConfig struct {
	//Driver is the docker volume to use
	Driver string
	//DriverOpts are the options to supply to the driver
	DriverOpts map[string]string
}

//Volume represents a docker volume which may be shared among multiple containers
type Volume struct {
	//Name is the name of the docker volume
	Name  string `json:"name"`
	//Labels to be attached to the volume
	Labels map[string]string
	VolumeConfig
}
