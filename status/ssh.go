package status

import (
	db "../db"
	util "../util"
	"log"
	"sync"
)

var _clients = map[int]*util.SshClient{}

var _mux = sync.Mutex{}

/*
   GetClient retrieves the ssh client for running a command
   on a remote server based on server id. It will create one if it
   does not exist.
*/
func GetClient(id int) (*util.SshClient, error) {
	cli, ok := _clients[id]
	if !ok || cli == nil {
		_mux.Lock()
		defer _mux.Unlock()
		server, _, err := db.GetServer(id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		cli, err = util.NewSshClient(server.Addr, id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		_clients[id] = cli
	}
	return cli, nil
}

/*
   GetClients functions similar to GetClient, except that it takes in
   an array of server ids and outputs an array of clients
*/
func GetClients(servers []int) ([]*util.SshClient, error) {

	out := make([]*util.SshClient, len(servers))
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
