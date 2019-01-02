package deploy

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"errors"
	"log"
	db "../db"
	util "../util"
)

var conf *util.Config = util.GetConfig()
/**
 * Builds out the Docker Network on pre-setup servers
 * Returns a string of all of the IP addresses 
 */
func Build(buildConf *Config,servers []db.Server,resources Resources,clients []*util.SshClient) ([]db.Server,error) {
	var sem	= semaphore.NewWeighted(conf.ThreadLimit)
	
	ctx := context.TODO()
	Prepare(buildConf.Nodes,servers)
	fmt.Println("-------------Building The Docker Containers-------------")
	n := buildConf.Nodes
	i := 0

	for n > 0 && i < len(servers){
		fmt.Printf("-------------Building on Server %d-------------\n",i)
		
		max_nodes := int(servers[i].Max - servers[i].Nodes)
		var nodes int
		if max_nodes > n {
			nodes = n
		}else{
			nodes = max_nodes
		}
		for j := 0; j < nodes; j++ {
			servers[i].Ips = append(servers[i].Ips,util.GetNodeIP(servers[i].ServerID,j))
		}

		prepareVlans(servers[i], nodes,clients[i])
		var startCmd string
		
		fmt.Printf("Creating the docker containers on server %d\n",i)

		switch(conf.Builder){
			case "local deploy legacy":
				startCmd = fmt.Sprintf("~/local_deploy/whiteblock -n %d -i %s -s %d -a %d -b %d -c %d -S",
					nodes,
					buildConf.Image,
					servers[i].ServerID,
					conf.ServerBits,
					conf.ClusterBits,
					conf.NodeBits)

			case "local deploy":
				extra_args := ""
				if !resources.NoCpuLimits() {
					extra_args += fmt.Sprintf(" -C %s",resources.Cpus)
				}
				if !resources.NoMemoryLimits() {
					mem,err := resources.GetMemory()
					if err != nil {
						log.Println(err)
						return nil, errors.New("Invalid value for memory")
					}
					extra_args += fmt.Sprintf(" -M %d",mem)
				}
				startCmd = fmt.Sprintf("~/local_deploy/deploy -n %d -i %s -s %d -a %d -b %d -c %d%s -S",
					nodes,
					buildConf.Image,
					servers[i].ServerID,
					conf.ServerBits,
					conf.ClusterBits,
					conf.NodeBits,
					extra_args)
			
			case "umba":
				startCmd = fmt.Sprintf("~/umba/umba -n %d -i %s -s %d -I %s",
				nodes,
				buildConf.Image,
				servers[i].ServerID,
				servers[i].Iface)

			case "genesis":
				fallthrough
			default:
				err := DockerRunAll(servers[i],clients[i],resources,nodes,buildConf.Image)
				if err != nil{
					log.Println(err)
					return nil,err
				}

		}
		if len(startCmd) > 0 {
			//Acquire resources
			err := sem.Acquire(ctx,1)
			if err != nil {
				log.Println(err)
				return nil,err
			}
			go func(server int,startCmd string){
				defer sem.Release(1)
				clients[server].Run(startCmd)
			}(i,startCmd)
		}
		
		n -= nodes
		i++
	}
	//Acquire all of the resources here, then release and destroy
	err := sem.Acquire(ctx,conf.ThreadLimit)
	if err != nil {
		return servers, nil
	}
	sem.Release(conf.ThreadLimit)
	if n != 0 {
		return servers, errors.New(fmt.Sprintf("ERROR: Only able to build %d/%d nodes\n",(buildConf.Nodes - n),buildConf.Nodes))
	}

	return servers, nil
}


func prepareVlans(server db.Server, nodes int,client *util.SshClient) error {
	switch(conf.Builder){
		case "local deploy":
			client.Run("~/local_deploy/deploy -k")
			if(conf.BuildMode == "stand alone"){
				cmd := fmt.Sprintf("cd ~/local_deploy && ./vlan -k && ./vlan -s %d -n %d -a %d -b %d -c %d -i %s --stand-alone", 
						server.ServerID, nodes, conf.ServerBits, conf.ClusterBits, conf.NodeBits, server.Iface)
				_,err := client.Run(cmd)
				if err != nil {
					log.Println(err)
					return err
				}
			}else{
				cmd := fmt.Sprintf("cd ~/local_deploy && ./vlan -k && ./vlan -s %d -n %d -a %d -b %d -c %d -i %s", 
						server.ServerID, nodes, conf.ServerBits, conf.ClusterBits, conf.NodeBits, server.Iface)
				_,err := client.Run(cmd)
				if err != nil {
					log.Println(err)
					return err
				}
			}
		case "local deploy legacy":
			client.Run("~/local_deploy/whiteblock -k")
			cmd := fmt.Sprintf("cd ~/local_deploy && ./vlan -B && ./vlan -s %d -n %d -a %d -b %d -c %d -i %s", 
					server.ServerID, nodes, conf.ServerBits, conf.ClusterBits, conf.NodeBits, server.Iface)
			_,err := client.Run(cmd)
			if err != nil {
				log.Println(err)
				return err
			}
		case "umba":
			//Nothing to do
		case "genesis":
			fallthrough
		default:
			DockerKillAll(client)
			DockerNetworkDestroyAll(client)
			err := DockerNetworkCreateAll(server,client,nodes)
			if err != nil {
				return err
			}
	}
	return nil
}
