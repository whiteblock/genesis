package rchain

import (
	db "../../db"
	util "../../util"
	"log"
)

func SetupPrometheus(servers []db.Server, clients []*util.SshClient) error {
	for i, server := range servers {
		for node, _ := range server.Ips {
			_, err := clients[i].DockerExecd(node, "prometheus --config.file=\"/prometheus.yml\"")
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}
	return nil
}
