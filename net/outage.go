package netconf

import (
	db "../db"
	ssh "../ssh"
	status "../status"
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"log"
	"strings"
	"sync"
)

func RemoveAllOutages(client *ssh.Client) error {
	res, err := client.Run("sudo iptables --list-rules | grep wb_bridge | grep DROP | grep FORWARD || true")
	if err != nil {
		log.Println(err)
		return err
	}
	if len(res) == 0 {
		return nil
	}
	res = strings.Replace(res, "-A ", "", -1)
	cmds := strings.Split(res, "\n")
	wg := sync.WaitGroup{}

	sem := semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()

	for _, cmd := range cmds {
		if len(cmd) == 0 {
			continue
		}
		sem.Acquire(ctx, 1)
		wg.Add(1)
		go func(cmd string) {
			defer sem.Release(1)
			defer wg.Done()
			_, err = client.Run(fmt.Sprintf("sudo iptables -D %s", cmd))
			if err != nil {
				log.Println(err)
			}
		}(cmd)
	}
	sem.Acquire(ctx, conf.ThreadLimit)
	sem.Release(conf.ThreadLimit)

	wg.Wait()
	return nil
}

func MakeOutageCommands(node1 db.Node, node2 db.Node) []string {
	return []string{
		fmt.Sprintf("FORWARD -i %s%d -d %s -j DROP", conf.BridgePrefix, node1.AbsoluteNum, node2.Ip),
		fmt.Sprintf("FORWARD -i %s%d -d %s -j DROP", conf.BridgePrefix, node2.AbsoluteNum, node1.Ip),
	}
}

func MakeOutage(node1 db.Node, node2 db.Node) error {
	cmds := MakeOutageCommands(node1, node2)

	client, err := status.GetClient(node1.Server)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = client.Run(fmt.Sprintf("sudo iptables -I %s", cmds[0]))
	if err != nil {
		log.Println(err)
		return err
	}
	client, err = status.GetClient(node2.Server)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = client.Run(fmt.Sprintf("sudo iptables -I %s", cmds[1]))
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func RemoveOutage(node1 db.Node, node2 db.Node) error {

	cmds := MakeOutageCommands(node1, node2)

	client, err := status.GetClient(node1.Server)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = client.Run(fmt.Sprintf("sudo iptables -D %s", cmds[0]))
	if err != nil {
		log.Println(err)
		return err
	}
	client, err = status.GetClient(node2.Server)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = client.Run(fmt.Sprintf("sudo iptables -D %s", cmds[1]))
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func CreatePartitionOutage(side1 []db.Node, side2 []db.Node) { //Doesn't report errors yet
	wg := sync.WaitGroup{}
	for _, node1 := range side1 {
		for _, node2 := range side2 {
			wg.Add(1)
			go func(node1 db.Node, node2 db.Node) {
				defer wg.Done()
				err := MakeOutage(node1, node2)
				if err != nil {
					log.Println(err)
				}
			}(node1, node2)
		}
	}
	wg.Wait()
}
