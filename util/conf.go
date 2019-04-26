package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

type Config struct {
	SSHUser            string  `json:"ssh-user"`
	SshKey             string  `json:"ssh-key"`
	ServerBits         uint32  `json:"server-bits"`
	ClusterBits        uint32  `json:"cluster-bits"`
	NodeBits           uint32  `json:"node-bits"`
	IPPrefix           uint32  `json:"ip-prefix"`
	Listen             string  `json:"listen"`
	Verbose            bool    `json:"verbose"`
	ThreadLimit        int64   `json:"thread-limit"`
	DockerOutputFile   string  `json:"docker-output-file"`
	Influx             string  `json:"influx"`
	InfluxUser         string  `json:"influx-user"`
	InfluxPassword     string  `json:"influx-password"`
	ServiceNetwork     string  `json:"service-network"`
	ServiceNetworkName string  `json:"service-network-name"`
	NodePrefix         string  `json:"node-prefix"`
	NodeNetworkPrefix  string  `json:"node-network-prefix"`
	ServicePrefix      string  `json:"service-prefix"`
	NodesPublicKey     string  `json:"nodes-public-key"`
	NodesPrivateKey    string  `json:"nodes-private-key"`
	HandleNodeSshKeys  bool    `json:"handle-node-ssh-keys"`
	MaxNodes           int     `json:"max-nodes"`
	MaxNodeMemory      string  `json:"max-node-memory"`
	MaxNodeCpu         float64 `json:"max-node-cpu"`
	BridgePrefix       string  `json:"bridge-prefix"`
}

/*
   Load the configuration from the Environment
*/
func (this *Config) LoadFromEnv() {
	var err error
	val, exists := os.LookupEnv("RSA_USER")
	if exists {
		this.SSHUser = val
	}
	val, exists = os.LookupEnv("LISTEN")
	if exists {
		this.Listen = val
	}
	val, exists = os.LookupEnv("RSA_KEY")
	if exists {
		this.SshKey = val
	}

	_, exists = os.LookupEnv("VERBOSE")
	if exists {
		this.Verbose = true
	}
	val, exists = os.LookupEnv("SERVER_BITS")
	if exists {
		tmp, err := strconv.ParseUint(val, 0, 32)
		this.ServerBits = uint32(tmp)
		if err != nil {
			fmt.Println("Invalid ENV value for SERVER_BITS")
			os.Exit(1)
		}
	}
	val, exists = os.LookupEnv("CLUSTER_BITS")
	if exists {
		tmp, err := strconv.ParseUint(val, 0, 32)
		this.ClusterBits = uint32(tmp)
		if err != nil {
			fmt.Println("Invalid ENV value for CLUSTER_BITS")
			os.Exit(1)
		}
	}
	val, exists = os.LookupEnv("NODE_BITS")
	if exists {
		tmp, err := strconv.ParseUint(val, 0, 32)
		this.NodeBits = uint32(tmp)
		if err != nil {
			fmt.Println("Invalid ENV value for NODE_BITS")
			os.Exit(1)
		}
	}
	val, exists = os.LookupEnv("THREAD_LIMIT")
	if exists {
		this.ThreadLimit, err = strconv.ParseInt(val, 0, 64)
		if err != nil {
			fmt.Println("Invalid ENV value for THREAD_LIMIT")
			os.Exit(1)
		}
	}
	val, exists = os.LookupEnv("IP_PREFIX")
	if exists {
		tmp, err := strconv.ParseUint(val, 0, 32)
		this.IPPrefix = uint32(tmp)
		if err != nil {
			fmt.Println("Invalid ENV value for IP_PREFIX")
			os.Exit(1)
		}
	}
	val, exists = os.LookupEnv("DOCKER_OUTPUT_FILE")
	if exists {
		this.DockerOutputFile = val
	}
	val, exists = os.LookupEnv("INFLUX")
	if exists {
		this.Influx = val
	}
	val, exists = os.LookupEnv("INFLUX_USER")
	if exists {
		this.InfluxUser = val
	}
	val, exists = os.LookupEnv("INFLUX_PASSWORD")
	if exists {
		this.InfluxPassword = val
	}
	val, exists = os.LookupEnv("SERVICE_NETWORK")
	if exists {
		this.ServiceNetwork = val
	}
	val, exists = os.LookupEnv("SERVICE_NETWORK_NAME")
	if exists {
		this.ServiceNetworkName = val
	}
	val, exists = os.LookupEnv("NODE_PREFIX")
	if exists {
		this.NodePrefix = val
	}
	val, exists = os.LookupEnv("NODE_NETWORK_PREFIX")
	if exists {
		this.NodeNetworkPrefix = val
	}
	val, exists = os.LookupEnv("SERVICE_PREFIX")
	if exists {
		this.ServicePrefix = val
	}

	val, exists = os.LookupEnv("NODES_PUBLIC_KEY")
	if exists {
		this.NodesPublicKey = val
	}

	val, exists = os.LookupEnv("NODES_PRIVATE_KEY")
	if exists {
		this.NodesPrivateKey = val
	}

	_, exists = os.LookupEnv("HANDLE_NODES_SSH_KEYS")
	if exists {
		this.HandleNodeSshKeys = true
	}

	val, exists = os.LookupEnv("MAX_NODES")
	if exists {
		tmp, err := strconv.ParseInt(val, 0, 32)
		this.MaxNodes = int(tmp)
		if err != nil {
			fmt.Println("Invalid ENV value for MAX_NODES")
			os.Exit(1)
		}
	}

	val, exists = os.LookupEnv("MAX_NODE_MEMORY")
	if exists {
		this.MaxNodeMemory = val
	}

	val, exists = os.LookupEnv("MAX_NODE_CPU")
	if exists {
		this.MaxNodeCpu, err = strconv.ParseFloat(val, 64)
		if err != nil {
			fmt.Println("Invalid ENV value for MAX_NODE_CPU")
			os.Exit(1)
		}
	}
	val, exists = os.LookupEnv("BRIDGE_PREFIX")
	if exists {
		this.BridgePrefix = val
	}
}

/*
   Fill in the missing essential values with the defaults.
*/
func (c *Config) AutoFillMissing() {
	if len(c.SSHUser) == 0 {
		c.SSHUser = "appo"
	}

	if len(c.Listen) == 0 {
		c.Listen = "127.0.0.1:8000"
	}
	if len(c.SshKey) == 0 {
		home := os.Getenv("HOME")
		c.SshKey = home + "/.ssh/id_rsa"
	}
	if c.ServerBits <= 0 {
		fmt.Println("Warning: Using default server bits")
		c.ServerBits = 8
	}
	if c.ClusterBits <= 0 {
		fmt.Println("Warning: Using default cluster bits")
		c.ClusterBits = 14
	}
	if c.NodeBits <= 0 {
		fmt.Println("Warning: Using default node bits")
		c.NodeBits = 2
	}
	if c.ThreadLimit <= 0 {
		fmt.Println("Warning: Using default thread limit")
		c.ThreadLimit = 10
	}

	if len(c.DockerOutputFile) == 0 {
		c.DockerOutputFile = "/output.log"
	}

	if len(c.ServiceNetwork) == 0 {
		c.ServiceNetwork = "172.30.0.0/16"
	}
	if len(c.ServiceNetworkName) == 0 {
		c.ServiceNetworkName = "wb_builtin_services"
	}

	if len(c.NodePrefix) == 0 {
		c.NodePrefix = "whiteblock-node"
	}

	if len(c.NodeNetworkPrefix) == 0 {
		c.NodeNetworkPrefix = "wb_vlan"
	}

	if len(c.ServicePrefix) == 0 {
		c.ServicePrefix = "wb_service"
	}

	if c.MaxNodes <= 0 {
		log.Println("Warning: No setting given for max nodes, defaulting to 200")
		c.MaxNodes = 200
	}

	if len(c.BridgePrefix) == 0 {
		c.BridgePrefix = "wb_bridge"
	}
}

var NodesPerCluster uint32

var conf *Config = nil

func init() {
	LoadConfig()
	conf.LoadFromEnv()
	conf.AutoFillMissing()
	NodesPerCluster = (1 << conf.NodeBits) - ReservedIps
}

/*
   The config from a file
*/
func LoadConfig() *Config {

	conf = new(Config)
	/**Load configuration**/
	dat, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Println("Warning: config.json not found, using defaults")
	} else {
		json.Unmarshal(dat, conf)
	}

	return conf
}

/*
   Get a pointer to the global config object.
   Do not modify this object
*/
func GetConfig() *Config {
	if conf == nil {
		LoadConfig()
	}
	return conf
}
