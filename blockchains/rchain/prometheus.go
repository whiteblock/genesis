package rchain

import(
	"fmt"
	"log"
	util "../../util"
	db "../../db"
)


func SetupPrometheus(servers []db.Server,clients []*util.SshClient) error {
	for i,server := range servers{
		err := clients[i].Scp("./blockchains/rchain/prometheus.yml","/home/appo/prometheus.yml")
		if err != nil{
			log.Println(err)
			return err
		}
		for node,_ := range server.Ips{
			_,err = clients[i].Run(fmt.Sprintf("docker cp /home/appo/prometheus.yml whiteblock-node%d:/prometheus.yml",node))
			if err != nil{
				log.Println(err)
				return err
			}
			_,err = clients[i].DockerExecd(node,"prometheus --config.file=\"/prometheus.yml\"")
			if err != nil{
				log.Println(err)
				return err
			}
		}
		_,err = clients[i].Run("rm -f ~/prometheus.yml")
		if err != nil{
			log.Println(err)
			return err
		}
	}
	return nil
}