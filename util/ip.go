package util

import (
    "fmt"
    "net"
    "errors"
    "log"
)

/**
 * Converts the IP address, given in network byte order,
 *  to a string in IPv4 dotted-decimal notation.
 * @see     InetNtoa(3)
 * @param   uint32  ip  The IP address in binary
 * @return  string      The IP address in ddnv4
 */
func InetNtoa(ip uint32) string {
    return fmt.Sprintf("%d.%d.%d.%d",
                (ip & (0x0FF << 0x018)) >> 0x018,
                (ip & (0x0FF << 0x010)) >> 0x010,
                (ip & (0x0FF << 0x08)) >> 0x08,
                (ip & 0x0FF))
}

/**
 * Calculate the IP address of a node, based on
 * the current IP scheme
 * @param  int      server  The server number
 * @param  int      node    The relative node number
 * @return string           The IP address of the node
 */
func GetNodeIP(server int,node int) string {
    var ip uint32 = conf.IPPrefix << (conf.NodeBits+conf.ClusterBits+conf.ServerBits)
    var clusterShift uint32 = conf.NodeBits
    var serverShift uint32 = conf.NodeBits + conf.ClusterBits
    var clusterLast uint32 =  (1 << conf.ClusterBits) - 1
    //set server bits
    ip += uint32(server) << serverShift
    //set cluster bits 
    cluster := uint32(uint32(node)/NodesPerCluster)
    //fmt.Printf("CLUSTER IS %d\n",cluster)
    ip += cluster << clusterShift
    //set the node bits
    if(cluster != clusterLast){
        ip += uint32(node)%NodesPerCluster + 2
    }
    return InetNtoa(ip)
}

/**
 * Calculate the gateway IP address for a node,
 * base on the current IP scheme
 * @param  int      server  The server number
 * @param  int      node    The relative node number
 * @return string           The node's gateway address
 */
func GetGateway(server int, node int) string {
    var ip uint32 = conf.IPPrefix << (conf.NodeBits+conf.ClusterBits+conf.ServerBits)
    clusterShift := conf.NodeBits
    serverShift := conf.NodeBits + conf.ClusterBits
    //set server bits
    ip += uint32(server) << serverShift
    //set cluster bits 
    cluster := uint32(uint32(node)/NodesPerCluster)
    ip += cluster << clusterShift
    ip += 1
    return InetNtoa(ip)
}

/**
 * Calculate the gateway IP addresses for all of the nodes
 * on a server
 * @param  int      server  The server number
 * @param  int      nodes   The number of nodes on that server
 * @return []string         A list of gateways for all of the nodes on that server
 */
func GetGateways(server int, nodes int) []string {
    clusters := uint32((uint32(nodes) - (uint32(nodes)%NodesPerCluster))/NodesPerCluster)
    out := []string{}
    var i uint32;
    for i = 0; i < clusters ; i++ {
        out = append(out,GetGateway(server,int(i * NodesPerCluster)))
    }

    return out;
}

/**
 * Calculate the subnet based on the IP scheme
 * @return int  The subnet for all of the nodes
 */
func GetSubnet() int {
    return 32 - int(conf.NodeBits)
}


func GetNetworkAddress(server int, node int) string {
    var ip uint32 = conf.IPPrefix << (conf.NodeBits+conf.ClusterBits+conf.ServerBits)
    clusterShift := conf.NodeBits
    serverShift := conf.NodeBits + conf.ClusterBits
    //set server bits
    ip += uint32(server) << serverShift
    //set cluster bits 
    cluster := uint32(uint32(node)/NodesPerCluster)
    ip += cluster << clusterShift
    return fmt.Sprintf("%s/%d",InetNtoa(ip),GetSubnet())
}

/**Increment ip address by 1**/
func inc(ip net.IP) {
    for i := len(ip) - 1; i >= 0; i-- {
        ip[i]++
        if ip[i] > 0 {
            return
        }
    }
}

/**Get a map of the service name to ips**/
func GetServiceIps(services []Service) (map[string]string,error) {
    out := make(map[string]string)
    _, ipnet, err := net.ParseCIDR(conf.ServiceNetwork)
    ip := ip.Mask(ipnet.Mask)
    if err != nil {
        log.Println(err)
        return nil,err
    }
    inc(ip)//skip first ip

    for _,service := range services {
        inc(ip)
        if !ipnet.Contains(ip) {
            return nil,errors.New("CIDR range too small")
        }
        out[service.Name] = ip
    }
    return out    
}

/**Get the gateway and the CIDR subnet**/
func GetServiceNetwork() (string, string, error){
    ip, _, err := net.ParseCIDR(conf.ServiceNetwork)
    if err != nil {
        log.Println(err)
        return "","",err
    }
    inc(ip)

    return ip.String(),conf.ServiceNetwork,nil
}
