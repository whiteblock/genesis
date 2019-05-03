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
