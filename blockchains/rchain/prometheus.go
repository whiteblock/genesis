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

package rchain

import (
	"../../db"
	"../../ssh"
	//"log"
)

func setupPrometheus(servers []db.Server, clients []*ssh.Client) error {
	/*for i, server := range servers {
		for node := range server.Ips {
			_, err := clients[i].DockerExecd(node, "prometheus --config.file=\"/prometheus.yml\"")
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}*/
	return nil
}
