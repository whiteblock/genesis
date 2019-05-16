/*
	Copyright 2019 whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Genesis is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package util

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
)

// Config groups all of the global configuration parameters into
// a single struct
type Config struct {
	SSHUser            string  `json:"ssh-user"`
	SSHKey             string  `json:"ssh-key"`
	ServerBits         uint32  `json:"server-bits"`
	ClusterBits        uint32  `json:"cluster-bits"`
	NodeBits           uint32  `json:"node-bits"`
	IPPrefix           uint32  `json:"ip-prefix"`
	Listen             string  `json:"listen"`
	Verbosity          string  `json:"verbosity"`
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
	HandleNodeSSHKeys  bool    `json:"handle-node-ssh-keys"`
	MaxNodes           int     `json:"max-nodes"`
	MaxNodeMemory      string  `json:"max-node-memory"`
	MaxNodeCPU         float64 `json:"max-node-cpu"`
	BridgePrefix       string  `json:"bridge-prefix"`
	APIEndpoint        string  `json:"api-endpoint"`
}

// LoadFromEnv loads the configuration from the Environment
func (c *Config) LoadFromEnv() {
	var err error
	val, exists := os.LookupEnv("RSA_USER")
	if exists {
		c.SSHUser = val
	}
	val, exists = os.LookupEnv("LISTEN")
	if exists {
		c.Listen = val
	}
	val, exists = os.LookupEnv("RSA_KEY")
	if exists {
		c.SSHKey = val
	}

	val, exists = os.LookupEnv("VERBOSITY")
	if exists {
		c.Verbosity = val
	}
	val, exists = os.LookupEnv("SERVER_BITS")
	if exists {
		tmp, err := strconv.ParseUint(val, 0, 32)
		c.ServerBits = uint32(tmp)
		if err != nil {
			fmt.Println("Invalid ENV value for SERVER_BITS")
			os.Exit(1)
		}
	}
	val, exists = os.LookupEnv("CLUSTER_BITS")
	if exists {
		tmp, err := strconv.ParseUint(val, 0, 32)
		c.ClusterBits = uint32(tmp)
		if err != nil {
			fmt.Println("Invalid ENV value for CLUSTER_BITS")
			os.Exit(1)
		}
	}
	val, exists = os.LookupEnv("NODE_BITS")
	if exists {
		tmp, err := strconv.ParseUint(val, 0, 32)
		c.NodeBits = uint32(tmp)
		if err != nil {
			fmt.Println("Invalid ENV value for NODE_BITS")
			os.Exit(1)
		}
	}
	val, exists = os.LookupEnv("THREAD_LIMIT")
	if exists {
		c.ThreadLimit, err = strconv.ParseInt(val, 0, 64)
		if err != nil {
			fmt.Println("Invalid ENV value for THREAD_LIMIT")
			os.Exit(1)
		}
	}
	val, exists = os.LookupEnv("IP_PREFIX")
	if exists {
		tmp, err := strconv.ParseUint(val, 0, 32)
		c.IPPrefix = uint32(tmp)
		if err != nil {
			fmt.Println("Invalid ENV value for IP_PREFIX")
			os.Exit(1)
		}
	}
	val, exists = os.LookupEnv("DOCKER_OUTPUT_FILE")
	if exists {
		c.DockerOutputFile = val
	}
	val, exists = os.LookupEnv("INFLUX")
	if exists {
		c.Influx = val
	}
	val, exists = os.LookupEnv("INFLUX_USER")
	if exists {
		c.InfluxUser = val
	}
	val, exists = os.LookupEnv("INFLUX_PASSWORD")
	if exists {
		c.InfluxPassword = val
	}
	val, exists = os.LookupEnv("SERVICE_NETWORK")
	if exists {
		c.ServiceNetwork = val
	}
	val, exists = os.LookupEnv("SERVICE_NETWORK_NAME")
	if exists {
		c.ServiceNetworkName = val
	}
	val, exists = os.LookupEnv("NODE_PREFIX")
	if exists {
		c.NodePrefix = val
	}
	val, exists = os.LookupEnv("NODE_NETWORK_PREFIX")
	if exists {
		c.NodeNetworkPrefix = val
	}
	val, exists = os.LookupEnv("SERVICE_PREFIX")
	if exists {
		c.ServicePrefix = val
	}

	val, exists = os.LookupEnv("NODES_PUBLIC_KEY")
	if exists {
		c.NodesPublicKey = val
	}

	val, exists = os.LookupEnv("NODES_PRIVATE_KEY")
	if exists {
		c.NodesPrivateKey = val
	}

	_, exists = os.LookupEnv("HANDLE_NODES_SSH_KEYS")
	if exists {
		c.HandleNodeSSHKeys = true
	}

	val, exists = os.LookupEnv("MAX_NODES")
	if exists {
		tmp, err := strconv.ParseInt(val, 0, 32)
		c.MaxNodes = int(tmp)
		if err != nil {
			fmt.Println("Invalid ENV value for MAX_NODES")
			os.Exit(1)
		}
	}

	val, exists = os.LookupEnv("MAX_NODE_MEMORY")
	if exists {
		c.MaxNodeMemory = val
	}

	val, exists = os.LookupEnv("MAX_NODE_CPU")
	if exists {
		c.MaxNodeCPU, err = strconv.ParseFloat(val, 64)
		if err != nil {
			fmt.Println("Invalid ENV value for MAX_NODE_CPU")
			os.Exit(1)
		}
	}
	val, exists = os.LookupEnv("BRIDGE_PREFIX")
	if exists {
		c.BridgePrefix = val
	}

	val, exists = os.LookupEnv("API_ENDPOINT")
	if exists {
		c.APIEndpoint = val
	}
}

// AutoFillMissing fills in the missing essential values with the defaults.
func (c *Config) AutoFillMissing() {
	if len(c.SSHUser) == 0 {
		c.SSHUser = "appo"
	}

	if len(c.Listen) == 0 {
		c.Listen = "127.0.0.1:8000"
	}
	if len(c.SSHKey) == 0 {
		home := os.Getenv("HOME")
		c.SSHKey = home + "/.ssh/id_rsa"
	}
	if c.ServerBits <= 0 {
		log.Warn("Using default server bits")
		c.ServerBits = 8
	}
	if c.ClusterBits <= 0 {
		log.Warn("Using default cluster bits")
		c.ClusterBits = 12
	}
	if c.NodeBits <= 0 {
		log.Warn("Using default node bits")
		c.NodeBits = 4
	}
	if c.ThreadLimit <= 0 {
		log.Warn("Using default thread limit")
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
		log.Warn("No setting given for max nodes, defaulting to 200")
		c.MaxNodes = 200
	}

	if len(c.BridgePrefix) == 0 {
		c.BridgePrefix = "wb_bridge"
	}

	if len(c.APIEndpoint) == 0 {
		c.APIEndpoint = "https://api.whiteblock.io"
	}
}

//NodesPerCluster represents the maximum number of nodes allowed in a cluster
var NodesPerCluster uint32

var conf *Config

func init() {
	LoadConfig()
	conf.LoadFromEnv()
	conf.AutoFillMissing()
	//log.SetReportCaller(true)
	lvl, err := log.ParseLevel(conf.Verbosity)
	if err != nil {
		panic(err)
	}
	log.SetLevel(lvl)
	NodesPerCluster = (1 << conf.NodeBits) - ReservedIps
}

// LoadConfig loads the config from the configuration file
func LoadConfig() *Config {

	conf = new(Config)
	/**Load configuration**/
	dat, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Warn("config.json not found, using defaults")
	} else {
		json.Unmarshal(dat, conf)
	}

	return conf
}

// GetConfig gets a pointer to the global config object.
// Do not modify c object
func GetConfig() *Config {
	if conf == nil {
		LoadConfig()
	}
	return conf
}
