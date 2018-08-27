package main

import "fmt"

/**
 * Calculate the exponential
 * @param  uint32	base	The exponent base
 * @param  uint32	exp		The exponent power
 * @return uint32			The result
 */
func pow(base uint32, exp uint32) uint32 {
	if exp <= 1 {
		return base	
	}
	return base * pow(base,exp - 1)
}

/**
 * Converts the IP address, given in network byte order,
 *  to a string in IPv4 dotted-decimal notation.
 * @see		inet_ntoa(3)
 * @param  	uint32	ip	The IP address in binary
 * @return 	string		The IP address in ddnv4
 */
func inet_ntoa(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
				(ip & (0x0FF << 0x018)) >> 0x018,
				(ip & (0x0FF << 0x010)) >> 0x010,
				(ip & (0x0FF << 0x08)) >> 0x08,
				(ip & 0x0FF))
}

/**
 * Calculate the IP address of a node, based on
 * the current IP scheme
 * @param  int		server	The server number
 * @param  int		node	The relative node number
 * @return string			The IP address of the node
 */
func getNodeIP(server int,node int) string {
	var ip uint32 = 10 << 24
	var clusterShift uint32 = NODE_BITS
	var serverShift uint32 = NODE_BITS + CLUSTER_BITS
	var clusterLast uint32 =  (1 << CLUSTER_BITS) - 1
	//set server bits
	ip += uint32(server) << serverShift
	//set cluster bits 
	cluster := uint32(uint32(node)/NODES_PER_CLUSTER)
	//fmt.Printf("CLUSTER IS %d\n",cluster)
	ip += cluster << clusterShift
	//set the node bits
	if(cluster == clusterLast){
		if(node != 0){
			ip += uint32(node)%NODES_PER_CLUSTER + 1
		}
	}else{
		ip += uint32(node)%NODES_PER_CLUSTER + 2
	}
	return inet_ntoa(ip)
}

/**
 * Calculate the gateway IP address for a node,
 * base on the current IP scheme
 * @param  int		server	The server number
 * @param  int		node	The relative node number
 * @return string			The node's gateway address
 */
func getGateway(server int, node int) string {
	var ip uint32 = 10 << 24
	clusterShift := NODE_BITS
	serverShift := NODE_BITS + CLUSTER_BITS
	//set server bits
	ip += uint32(server) << serverShift
	//set cluster bits 
	cluster := uint32(uint32(node)/NODES_PER_CLUSTER)
	ip += cluster << clusterShift
	ip += 1
	return inet_ntoa(ip)
}

/**
 * Calculate the gateway IP addresses for all of the nodes
 * on a server
 * @param  int		server	The server number
 * @param  int		nodes	The number of nodes on that server
 * @return []string			A list of gateways for all of the nodes on that server
 */
func getGateways(server int, nodes int) []string {
	clusters := uint32((uint32(nodes) - (uint32(nodes)%NODES_PER_CLUSTER))/NODES_PER_CLUSTER) + 1
	out := []string{}
	var i uint32;
	for i = 0; i < clusters ; i++ {
		out = append(out,getGateway(server,int(i * NODES_PER_CLUSTER)))
	}

	return out;
}

/**
 * Calculate the subnet based on the IP scheme
 * @return int	The subnet for all of the nodes
 */
func getSubnet() int {
	return 32 - int(NODE_BITS)
}