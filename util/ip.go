package util

import (
	"fmt"
	"log"
	"net"
)

// ReservedIps indicates the number of ip addresses reserved in
// a cluster's subnet
const ReservedIps uint32 = 3

// InetNtoa converts the IP address, given in network byte order,
// to a string in IPv4 dotted-decimal notation.
func InetNtoa(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		(ip&(0x0FF<<0x018))>>0x018,
		(ip&(0x0FF<<0x010))>>0x010,
		(ip&(0x0FF<<0x08))>>0x08,
		ip&0x0FF)
}

// GetNodeIP calculates the IP address of a node, based on
// the current IP scheme
func GetNodeIP(server int, node int) string {
	var ip = conf.IPPrefix << (conf.NodeBits + conf.ClusterBits + conf.ServerBits)
	var clusterShift = conf.NodeBits
	var serverShift = conf.NodeBits + conf.ClusterBits
	var clusterLast uint32 = (1 << conf.ClusterBits) - 1
	//set server bits
	ip += uint32(server) << serverShift
	//set cluster bits
	cluster := uint32(uint32(node) / NodesPerCluster)
	//fmt.Printf("CLUSTER IS %d\n",cluster)
	ip += cluster << clusterShift
	//set the node bits
	if cluster != clusterLast {
		ip += uint32(node)%NodesPerCluster + 2
	}
	return InetNtoa(ip)
}

// GetInfoFromIP returns the server number and the node number calculated from the given
// IPv4 address based on the current IP scheme.
func GetInfoFromIP(ipStr string) (int, int) {
	ipBytes := net.ParseIP(ipStr).To4()
	var rawIP uint32
	for _, ipByte := range ipBytes {
		rawIP = rawIP << 8
		rawIP += uint32(ipByte)
	}
	var clusterLast uint32 = (1 << conf.ClusterBits) - 1
	server := (rawIP >> (conf.NodeBits + conf.ClusterBits)) & ((1 << conf.ServerBits) - 1)
	cluster := (rawIP >> conf.NodeBits) & ((1 << conf.ClusterBits) - 1)
	var node uint32

	if cluster != clusterLast {
		node = (rawIP & ((1 << conf.NodeBits) - 1)) - 2
	} else {
		node = 0
	}
	node = node + (cluster * NodesPerCluster)
	return int(server), int(node)
}

// GetGateway calculates the gateway IP address for a node,
// base on the current IP scheme
func GetGateway(server int, node int) string {
	var ip = conf.IPPrefix << (conf.NodeBits + conf.ClusterBits + conf.ServerBits)
	clusterShift := conf.NodeBits
	serverShift := conf.NodeBits + conf.ClusterBits
	//set server bits
	ip += uint32(server) << serverShift
	//set cluster bits
	cluster := uint32(uint32(node) / NodesPerCluster)
	ip += cluster << clusterShift
	ip++
	return InetNtoa(ip)
}

// GetGateways calculates the gateway IP addresses for all of the nodes
// on a server.
func GetGateways(server int, nodes int) []string {
	clusters := uint32((uint32(nodes) - (uint32(nodes) % NodesPerCluster)) / NodesPerCluster)
	out := []string{}
	var i uint32
	for i = 0; i < clusters; i++ {
		out = append(out, GetGateway(server, int(i*NodesPerCluster)))
	}

	return out
}

// GetSubnet calculates the subnet based on the IP scheme
func GetSubnet() int {
	return 32 - int(conf.NodeBits)
}

// GetWholeNetworkIP gets the network ip of the whole network for a server.
func GetWholeNetworkIP(server int) string {
	var ip = conf.IPPrefix << (conf.NodeBits + conf.ClusterBits + conf.ServerBits)
	var serverShift = conf.NodeBits + conf.ClusterBits
	//set server bits
	ip += uint32(server) << serverShift
	return InetNtoa(ip)
}

// GetNetworkAddress gets the network address of the cluster the given node belongs to.
func GetNetworkAddress(server int, node int) string {
	var ip = conf.IPPrefix << (conf.NodeBits + conf.ClusterBits + conf.ServerBits)
	clusterShift := conf.NodeBits
	serverShift := conf.NodeBits + conf.ClusterBits
	//set server bits
	ip += uint32(server) << serverShift
	//set cluster bits
	cluster := uint32(uint32(node) / NodesPerCluster)
	ip += cluster << clusterShift
	return fmt.Sprintf("%s/%d", InetNtoa(ip), GetSubnet())
}

// inc increments an ip address by 1
func inc(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			return
		}
	}
}

// GetServiceIps creates a map of the service names to their ip addresses. Useful
// for determining the ip address of a service.
func GetServiceIps(services []Service) (map[string]string, error) {
	out := make(map[string]string)
	ip, ipnet, err := net.ParseCIDR(conf.ServiceNetwork)
	ip = ip.Mask(ipnet.Mask)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	inc(ip) //skip first ip

	for _, service := range services {
		inc(ip)
		if !ipnet.Contains(ip) {
			return nil, fmt.Errorf("CIDR range too small")
		}
		out[service.Name] = ip.String()
	}
	return out, nil
}

// GetServiceNetwork gets the network address in CIDR of the service network
func GetServiceNetwork() (string, string, error) {
	ip, ipnet, err := net.ParseCIDR(conf.ServiceNetwork)
	if err != nil {
		log.Println(err)
		return "", "", err
	}

	return ip.String(), ipnet.String(), nil
}
