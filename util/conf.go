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
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config groups all of the global configuration parameters into
// a single struct
type Config struct {
	SSHUser                       string  `mapstructure:"sshUser"`
	SSHKey                        string  `mapstructure:"sshKey"`
	SSHHost                       string  `mapstructure:"sshHost"`
	ServerBits                    uint32  `mapstructure:"serverBits"`
	ClusterBits                   uint32  `mapstructure:"clusterBits"`
	NodeBits                      uint32  `mapstructure:"nodeBits"`
	IPPrefix                      uint32  `mapstructure:"ipPrefix"`
	Listen                        string  `mapstructure:"listen"`
	Verbosity                     string  `mapstructure:"verbosity"`
	DockerOutputFile              string  `mapstructure:"dockerOutputFile"`
	Influx                        string  `mapstructure:"influx"`
	InfluxUser                    string  `mapstructure:"influxUser"`
	InfluxPassword                string  `mapstructure:"influxPassword"`
	ServiceNetwork                string  `mapstructure:"serviceNetwork"`
	ServiceNetworkName            string  `mapstructure:"serviceNetworkName"`
	NodePrefix                    string  `mapstructure:"nodePrefix"`
	NodeNetworkPrefix             string  `mapstructure:"nodeNetworkPrefix"`
	ServicePrefix                 string  `mapstructure:"servicePrefix"`
	NodesPublicKey                string  `mapstructure:"nodesPublicKey"`
	NodesPrivateKey               string  `mapstructure:"nodesPrivateKey"`
	HandleNodeSSHKeys             bool    `mapstructure:"handleNodeSshKeys"`
	MaxNodes                      int     `mapstructure:"maxNodes"`
	MaxNodeMemory                 string  `mapstructure:"maxNodeMemory"`
	MaxNodeCPU                    float64 `mapstructure:"maxNodeCpu"`
	BridgePrefix                  string  `mapstructure:"bridgePrefix"`
	APIEndpoint                   string  `mapstructure:"apiEndpoint"`
	NibblerEndPoint               string  `mapstructure:"nibblerEndPoint"`
	LogJSON                       bool    `mapstructure:"logJson"`
	PrometheusConfig              string  `mapstructure:"prometheusConfig"`
	PrometheusPort                int     `mapstructure:"prometheusPort"`
	PrometheusInstrumentationPort int     `mapstructure:"prometheusInstrumentationPort"`
	MaxRunAttempts                int     `mapstructure:"maxRunAttempts"`
	MaxConnections                int     `mapstructure:"maxConnections"`
	DataDirectory                 string  `mapstructure:"datadir"`
	DisableNibbler                bool    `mapstructure:"disableNibbler"`
	DisableTestnetReporting       bool    `mapstructure:"disableTestnetReporting"`
	MaxCommandOutputLogSize       int     `mapstructure:"maxCommandOutputLogSize"`
}

//NodesPerCluster represents the maximum number of nodes allowed in a cluster
var NodesPerCluster uint32

var conf = new(Config)

func setViperEnvBindings() {
	viper.BindEnv("sshUser", "SSH_USER")
	viper.BindEnv("listen", "LISTEN")
	viper.BindEnv("sshKey", "SSH_KEY")
	viper.BindEnv("verbosity", "VERBOSITY")
	viper.BindEnv("serverBits", "SERVER_BITS")
	viper.BindEnv("clusterBits", "CLUSTER_BITS")
	viper.BindEnv("nodeBits", "NODE_BITS")
	viper.BindEnv("ipPrefix", "IP_PREFIX")
	viper.BindEnv("dockerOutputFile", "DOCKER_OUTPUT_FILE")
	viper.BindEnv("influx", "INFLUX")
	viper.BindEnv("influxUser", "INFLUX_USER")
	viper.BindEnv("influxPassword", "INFLUX_PASSWORD")
	viper.BindEnv("serviceNetwork", "SERVICE_NETWORK")
	viper.BindEnv("serviceNetworkName", "SERVICE_NETWORK_NAME")
	viper.BindEnv("nodePrefix", "NODE_PREFIX")
	viper.BindEnv("nodeNetworkPrefix", "NODE_NETWORK_PREFIX")
	viper.BindEnv("servicePrefix", "SERVICE_PREFIX")
	viper.BindEnv("nodesPublicKey", "NODES_PUBLIC_KEY")
	viper.BindEnv("nodesPrivateKey", "NODES_PRIVATE_KEY")
	viper.BindEnv("handleNodeSshKeys", "HANDLE_NODES_SSH_KEYS")
	viper.BindEnv("maxNodes", "MAX_NODES")
	viper.BindEnv("maxNodeMemory", "MAX_NODE_MEMORY")
	viper.BindEnv("maxNodeCPU", "MAX_NODE_CPU")
	viper.BindEnv("bridgePrefix", "BRIDGE_PREFIX")
	viper.BindEnv("apiEndpoint", "API_ENDPOINT")
	viper.BindEnv("nibblerEndPoint", "NIBBLER_END_POINT")
	viper.BindEnv("logJson", "LOG_JSON")
	viper.BindEnv("maxRunAttempts", "MAX_RUN_ATTEMPTS")
	viper.BindEnv("maxConnections", "MAX_CONNECTIONS")
	viper.BindEnv("datadir", "DATADIR")
	viper.BindEnv("disableNibbler", "DISABLE_NIBBLER")
	viper.BindEnv("disableTestnetReporting", "DISABLE_TESTNET_REPORTING")
	viper.BindEnv("maxCommandOutputLogSize", "MAX_COMMAND_OUTPUT_LOG_SIZE")
}
func setViperDefaults() {
	viper.SetDefault("sshUser", os.Getenv("USER"))
	viper.SetDefault("listen", "127.0.0.1:8000")
	viper.SetDefault("sshKey", os.Getenv("HOME")+"/.ssh/id_rsa")
	viper.SetDefault("serverBits", 8)
	viper.SetDefault("clusterBits", 12)
	viper.SetDefault("nodeBits", 4)
	viper.SetDefault("dockerOutputFile", "/output.log")
	viper.SetDefault("serviceNetwork", "172.30.0.1/16")
	viper.SetDefault("serviceNetworkName", "wb_builtin_services")
	viper.SetDefault("nodePrefix", "whiteblock-node")
	viper.SetDefault("nodeNetworkPrefix", "wb_vlan")
	viper.SetDefault("servicePrefix", "wb_service")
	viper.SetDefault("sshHost", "127.0.0.1")
	viper.SetDefault("verbosity", "INFO")
	viper.SetDefault("maxNodes", 200)
	viper.SetDefault("bridgePrefix", "wb_bridge")
	viper.SetDefault("apiEndpoint", "https://api.whiteblock.io")
	viper.SetDefault("nibblerEndPoint", "https://storage.googleapis.com/genesis-public/nibbler/master/bin/linux/amd64/nibbler")
	viper.SetDefault("prometheusConfig", "/tmp/prometheus.yml")
	viper.SetDefault("prometheusPort", 8088)
	viper.SetDefault("prometheusInstrumentationPort", 8008)
	viper.SetDefault("maxRunAttempts", 30)
	viper.SetDefault("maxConnections", 50)
	viper.SetDefault("datadir", os.Getenv("HOME")+"/.config/whiteblock/")
	viper.SetDefault("disableNibbler", false)
	viper.SetDefault("disableTestnetReporting", false)
	viper.SetDefault("maxCommandOutputLogSize", -1)
}

// GCPFormatter enables the ability to use genesis logging with Stackdriver
type GCPFormatter struct {
	JSON           *log.JSONFormatter
	ConstantFields log.Fields
}

// Format takes in the entry and processes it into the appropiate log entry
func (gf GCPFormatter) Format(entry *log.Entry) ([]byte, error) {
	for k, v := range gf.ConstantFields {
		entry.Data[k] = v
	}
	return gf.JSON.Format(entry)
}

func init() {
	setViperDefaults()
	setViperEnvBindings()
	viper.AddConfigPath("/etc/whiteblock/")          // path to look for the config file in
	viper.AddConfigPath("$HOME/.config/whiteblock/") // call multiple times to add many search paths
	viper.SetConfigName("genesis")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()

	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("could not find the config file")
	}
	err = viper.Unmarshal(&conf)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	lvl, err := log.ParseLevel(conf.Verbosity)
	if err != nil {
		log.SetLevel(log.InfoLevel)
		log.Warn(err)
	}
	log.SetLevel(lvl)
	NodesPerCluster = (1 << conf.NodeBits) - ReservedIps

	if conf.LogJSON {
		log.SetFormatter(&GCPFormatter{
			JSON: &log.JSONFormatter{
				FieldMap: log.FieldMap{
					log.FieldKeyTime:  "eventTime",
					log.FieldKeyLevel: "severity",
					log.FieldKeyMsg:   "message",
				},
			},
			ConstantFields: log.Fields{
				"serviceContext": map[string]string{"service": "genesis", "version": "1.8.2"},
			},
		})
	}

	err = os.MkdirAll(conf.DataDirectory, 0776)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "dir": conf.DataDirectory}).Fatal("could not create data directory")
	}
}

// GetConfig gets a pointer to the global config object.
// Do not modify conf object
func GetConfig() *Config {
	return conf
}
