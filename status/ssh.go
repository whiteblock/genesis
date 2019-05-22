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

package status

import (
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"log"
	"sync"
)

var _clients = map[int]*ssh.Client{}

var _mux = sync.Mutex{}

// GetClient retrieves the ssh client for running a command
// on a remote server based on server id. It will create one if it
// does not exist.
func GetClient(id int) (*ssh.Client, error) {
	cli, ok := _clients[id]
	if !ok || cli == nil {
		_mux.Lock()
		defer _mux.Unlock()
		server, _, err := db.GetServer(id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		cli, err = ssh.NewClient(server.Addr, id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		_clients[id] = cli
	}
	return cli, nil
}

// GetClients functions similar to GetClient, except that it takes in
// an array of server ids and outputs an array of clients
func GetClients(servers []int) ([]*ssh.Client, error) {

	out := make([]*ssh.Client, len(servers))
	var err error
	for i := 0; i < len(servers); i++ {
		out[i], err = GetClient(servers[i])
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	return out, nil
}

// GetClientsFromNodes gets all of the ssh clients you need for
// communication with the given nodes
func GetClientsFromNodes(nodes []db.Node) ([]*ssh.Client, error) {
	serverIds := db.GetUniqueServerIDs(nodes)
	return GetClients(serverIds)
}
