/*
   Allows the ability for the simulation of network conditions accross nodes.
*/
package netconf

import (
	db "../db"
	ssh "../ssh"
	status "../status"
	util "../util"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

/**
[ limit PACKETS ]
[ delay TIME [ JITTER [CORRELATION]]]
[ distribution {uniform|normal|pareto|paretonormal} ]
[ corrupt PERCENT [CORRELATION]]
[ duplicate PERCENT [CORRELATION]]
[ loss random PERCENT [CORRELATION]]
[ loss state P13 [P31 [P32 [P23 P14]]]
[ loss gemodel PERCENT [R [1-H [1-K]]]
[ ecn ]
[ reorder PRECENT [CORRELATION] [ gap DISTANCE ]]
[ rate RATE [PACKETOVERHEAD] [CELLSIZE] [CELLOVERHEAD]]
*/

var conf *util.Config = util.GetConfig()

type Netconf struct {
	Node        int     `json:"node"`
	Limit       int     `json:"limit"`
	Loss        float64 `json:"loss"` //Loss % ie 100% = 100
	Delay       int     `json:"delay"`
	Rate        string  `json:"rate"`
	Duplication float64 `json:"duplicate"`
	Corrupt     float64 `json:"corrupt"`
	Reorder     float64 `json:"reorder"`
}

/*
   CreateCommands generates the commands needed to obtain the desired
   network conditions
*/
func CreateCommands(netconf Netconf, serverId int) []string {
	const offset int = 6
	out := []string{
		fmt.Sprintf("sudo tc qdisc del dev %s%d root", conf.BridgePrefix, netconf.Node),
		fmt.Sprintf("sudo tc qdisc add dev %s%d root handle 1: prio", conf.BridgePrefix, netconf.Node),
		fmt.Sprintf("sudo tc qdisc add dev %s%d parent 1:1 handle 2: netem", conf.BridgePrefix, netconf.Node), //unf
		fmt.Sprintf("sudo tc filter add dev %s%d parent 1:0 protocol ip pref 55 handle %d fw flowid 2:1",
			conf.BridgePrefix, netconf.Node, offset),
		fmt.Sprintf("sudo iptables -t mangle -A PREROUTING  ! -d %s -j MARK --set-mark %d",
			util.GetGateway(serverId, netconf.Node), offset),
	}

	if netconf.Limit > 0 {
		out[2] += fmt.Sprintf(" limit %d", netconf.Limit)
	}

	if netconf.Loss > 0 {
		out[2] += fmt.Sprintf(" loss %.4f", netconf.Loss)
	}

	if netconf.Delay > 0 {
		out[2] += fmt.Sprintf(" delay %dus", netconf.Delay)
	}

	if len(netconf.Rate) > 0 {
		out[2] += fmt.Sprintf(" rate %s", netconf.Rate)
	}

	if netconf.Duplication > 0 {
		out[2] += fmt.Sprintf(" duplicate %.4f", netconf.Duplication)
	}

	if netconf.Corrupt > 0 {
		out[2] += fmt.Sprintf(" corrupt %.4f", netconf.Duplication)
	}

	if netconf.Reorder > 0 {
		out[2] += fmt.Sprintf(" reorder %.4f", netconf.Reorder)
	}

	return out
}

/*
   Apply applies the given network config.
*/
func Apply(client *ssh.Client, netconf Netconf, serverId int) error {
	cmds := CreateCommands(netconf, serverId)
	for i, cmd := range cmds {
		_, err := client.Run(cmd)
		if i == 0 {
			//Don't check the success of the first command which clears
			continue
		}
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

/*
   ApplyAll applies all of the given netconfs
*/
func ApplyAll(netconfs []Netconf, nodes []db.Node) error {
	for _, netconf := range netconfs {
		node, err := db.GetNodeByLocalId(nodes, netconf.Node)
		if err != nil {
			log.Println(err)
			return err
		}
		client, err := status.GetClient(node.Server)
		if err != nil {
			log.Println(err)
			return err
		}
		err = Apply(client, netconf, node.Server)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

/*
   ApplyToAll applies the given netconf to `nodes` nodes in the network on the given server
*/
func ApplyToAll(netconf Netconf, nodes []db.Node) error {
	for _, node := range nodes {
		netconf.Node = node.LocalId
		cmds := CreateCommands(netconf, node.Server)
		for i, cmd := range cmds {
			client, err := status.GetClient(node.Server)
			if err != nil {
				log.Println(err)
				return err
			}
			_, err = client.Run(cmd)
			if i == 0 {
				//Don't check the success of the first command which clears
				continue
			}
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}
	return nil
}

/*
   RemoveAll removes network conditions from the given number of nodes
*/
func RemoveAll(nodes []db.Node) error {
	for _, node := range nodes {
		client, err := status.GetClient(node.Server)
		if err != nil {
			log.Println(err)
			return err
		}
		client.Run(
			fmt.Sprintf("sudo tc qdisc del dev %s%d root", conf.BridgePrefix, node.LocalId))
	}
	return nil
}

/*
   RemoveAll removes network conditions from the given number of nodes
*/
func RemoveAllOnServer(client *ssh.Client, nodes int) {
	for i := 0; i < nodes; i++ {
		client.Run(
			fmt.Sprintf("sudo tc qdisc del dev %s%d root", conf.BridgePrefix, i))
	}
	RemoveAllOutages(client)
}

func parseItems(items []string, nconf *Netconf) error {

	for i := 0; i < len(items)/2; i++ {
		switch items[2*i] {
		case "limit":
			val, err := strconv.Atoi(items[2*i+1])
			if err != nil {
				log.Println(err)
				return err
			}
			nconf.Limit = val
		case "loss":
			val, err := strconv.ParseFloat(items[2*i+1][:len(items[2*i+1])-1], 64)
			if err != nil {
				log.Println(err)
				return err
			}
			nconf.Loss = val
		case "delay":
			re := regexp.MustCompile(`(?m)[0-9]+\.[0-9]+`)
			matches := re.FindAllString(items[2*i+1], -1)
			if len(matches) == 0 {
				return fmt.Errorf("Unexpected delay value \"%s\"", items[2*i+1])
			}

			val, err := strconv.ParseFloat(matches[0], 64)
			if err != nil {
				log.Println(err)
				return err
			}
			unit := items[2*i+1][len(matches[0]):]
			switch unit {
			case "s":
				val *= 1000
				fallthrough
			case "ms":
				val *= 1000
			}
			nconf.Delay = int(val)
		case "rate":
			nconf.Rate = items[2*i+1]
		case "duplicate":
			val, err := strconv.ParseFloat(items[2*i+1][:len(items[2*i+1])-1], 64)
			if err != nil {
				log.Println(err)
				return err
			}
			nconf.Duplication = val
		case "corrupt":
			val, err := strconv.ParseFloat(items[2*i+1][:len(items[2*i+1])-1], 64)
			if err != nil {
				log.Println(err)
				return err
			}
			nconf.Corrupt = val
		case "reorder":
			val, err := strconv.ParseFloat(items[2*i+1][:len(items[2*i+1])-1], 64)
			if err != nil {
				log.Println(err)
				return err
			}
			nconf.Reorder = val
		}
	}
	return nil
}

//5 start index
func GetConfigOnServer(client *ssh.Client) ([]Netconf, error) {
	res, err := client.Run("tc qdisc show | grep wb_bridge | grep netem || true")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if len(res) == 0 {
		return []Netconf{}, nil
	}
	out := []Netconf{}
	rawConfigs := strings.Split(res, "\n")

	for _, rawConfig := range rawConfigs { //4 for bridge name //7 for start of the shit
		rawItems := strings.Split(rawConfig, " ")
		if len(rawItems) < 5 {
			continue
		}
		bridgeName := rawItems[4]

		num, err := strconv.Atoi(bridgeName[len(conf.BridgePrefix):])
		if err != nil {
			log.Println(err)
			return nil, err
		}
		nconf := Netconf{Node: num}
		if len(rawItems) >= 8 {
			items := rawItems[7:]
			err = parseItems(items, &nconf)
			if err != nil {
				log.Println(err)
				return nil, err
			}
		}
		out = append(out, nconf)
	}
	return out, nil
}
