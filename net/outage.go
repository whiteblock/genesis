package netconf

import (
	db "../db"
	ssh "../ssh"
	status "../status"
	"fmt"
	"log"
	"strings"
)

func RemoveAllOutages(client *ssh.Client) error {
	res, err := client.Run("sudo iptables --list-rules | grep wb_bridge | grep DROP | grep FORWARD")
	if err != nil {
		log.Println(err)
		return err
	}
	res = strings.Replace(res, "-A ", "", -1)
	cmds := strings.Split(res, "\n")

	for _, cmd := range cmds {
		_, err = client.Run(fmt.Sprintf("sudo iptables -D %s", cmd))
		if err != nil {
			log.Println(err)
			return err
		}
	}
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
