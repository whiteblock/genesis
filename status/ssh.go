package status

import (
	"../db"
	"../ssh"
	"log"
	"sync"
)

var _clients = map[int]*ssh.Client{}

var _mux = sync.Mutex{}

/*
   GetClient retrieves the ssh client for running a command
   on a remote server based on server id. It will create one if it
   does not exist.
*/
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

/*
   GetClients functions similar to GetClient, except that it takes in
   an array of server ids and outputs an array of clients
*/
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

func GetClientsFromNodes(nodes []db.Node) ([]*ssh.Client, error) {
	serverIds := db.GetUniqueServerIds(nodes)
	return GetClients(serverIds)
}
