package status

import (
	"../db"
	"../ssh"
	"../util"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
)

var conf *util.Config

func init() {
	conf = util.GetConfig()
}

type Comp struct {
	Cpu float64 `json:"cpu"`
	Vsz float64 `json:"virtualMemorySize"`
	Rss float64 `json:"residentSetSize"`
}

/*
   Represents the status of the node
*/
type NodeStatus struct {
	Name      string `json:"name"`
	Server    int    `json:"server"`
	Ip        string `json:"ip"`
	Up        bool   `json:"up"`
	Resources Comp   `json:"resourceUse"`
}

/*
   Finds the index of a node by name and server id
*/
func FindNodeIndex(status []NodeStatus, name string, serverId int) int {
	for i, stat := range status {
		if stat.Name == name && serverId == stat.Server {
			return i
		}
	}
	return -1
}

/*
   Gets the cpu usage of a node
*/
func SumResUsage(c *ssh.Client, name string) (Comp, error) {
	res, err := c.Run(fmt.Sprintf("docker exec %s ps aux --no-headers | grep -v nibbler | awk '{print $3,$5,$6}'", name))
	if err != nil {
		log.Println(err)
		return Comp{-1, -1, -1}, err
	}
	procs := strings.Split(res, "\n")
	//fmt.Printf("%#v\n", procs)
	var out Comp
	for _, proc := range procs {
		if len(proc) == 0 {
			continue
		}
		values := strings.Split(proc, " ")

		cpu, err := strconv.ParseFloat(values[0], 64)
		if err != nil {
			log.Println(err)
			return Comp{-1, -1, -1}, err
		}
		out.Cpu += cpu

		vsz, err := strconv.ParseFloat(values[1], 64)
		if err != nil {
			log.Println(err)
			return Comp{-1, -1, -1}, err
		}
		out.Vsz += vsz

		rss, err := strconv.ParseFloat(values[2], 64)
		if err != nil {
			log.Println(err)
			return Comp{-1, -1, -1}, err
		}
		out.Rss += rss

	}
	return out, nil
}

/*
   Checks the status of the nodes in the current testnet
*/
func CheckNodeStatus(nodes []db.Node) ([]NodeStatus, error) {

	serverIds := []int{}
	out := make([]NodeStatus, len(nodes))

	for _, node := range nodes {
		push := true
		for _, id := range serverIds {
			if id == node.Server {
				push = false
			}
		}
		if push {
			serverIds = append(serverIds, node.Server)
		}
		//fmt.Printf("ABS = %d; REL=%d;NAME=%s%d\n", node.AbsoluteNum, node.LocalID, conf.NodePrefix, node.LocalID)
		out[node.AbsoluteNum] = NodeStatus{
			Name:      fmt.Sprintf("%s%d", conf.NodePrefix, node.LocalID),
			Ip:        node.IP,
			Server:    node.Server,
			Up:        false,
			Resources: Comp{-1, -1, -1},
		} //local id to testnet

	}
	servers, err := db.GetServers(serverIds)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	mux := sync.Mutex{}
	wg := sync.WaitGroup{}

	for _, server := range servers {
		client, err := GetClient(server.Id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		res, err := client.Run(
			fmt.Sprintf("docker ps | egrep -o '%s[0-9]*' | sort", conf.NodePrefix))
		if err != nil {
			log.Println(err)
			return nil, err
		}
		names := strings.Split(res, "\n")
		for _, name := range names {
			if len(name) == 0 {
				continue
			}

			index := FindNodeIndex(out, name, server.Id)
			if index == -1 {
				log.Printf("name=\"%s\",server=%d\n", name, server.Id)
				continue
			}
			wg.Add(1)
			go func(client *ssh.Client, name string, index int) {
				defer wg.Done()
				resUsage, err := SumResUsage(client, name)
				if err != nil {
					log.Println(err)
				}
				mux.Lock()
				out[index].Up = true
				out[index].Resources = resUsage
				mux.Unlock()
			}(client, name, index)
		}
	}
	wg.Wait()
	return out, nil
}
