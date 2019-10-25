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
	SSHUser                 string  `mapstructure:"sshUser"`
	SSHKey                  string  `mapstructure:"sshKey"`
	SSHHost                 string  `mapstructure:"sshHost"`
	ServerBits              uint32  `mapstructure:"serverBits"`
	ClusterBits             uint32  `mapstructure:"clusterBits"`
	NodeBits                uint32  `mapstructure:"nodeBits"`
	IPPrefix                uint32  `mapstructure:"ipPrefix"`
	Listen                  string  `mapstructure:"listen"`
	Verbosity               string  `mapstructure:"verbosity"`
	DockerOutputFile        string  `mapstructure:"dockerOutputFile"`
	Influx                  string  `mapstructure:"influx"`         //No default
	InfluxUser              string  `mapstructure:"influxUser"`     //No default
	InfluxPassword          string  `mapstructure:"influxPassword"` //No default
	ServiceNetwork          string  `mapstructure:"serviceNetwork"`
	ServiceNetworkName      string  `mapstructure:"serviceNetworkName"`
	NodePrefix              string  `mapstructure:"nodePrefix"`
	NodeNetworkPrefix       string  `mapstructure:"nodeNetworkPrefix"`
	ServicePrefix           string  `mapstructure:"servicePrefix"`
	NodesPublicKey          string  `mapstructure:"nodesPublicKey"`  //No default
	NodesPrivateKey         string  `mapstructure:"nodesPrivateKey"` //No default
	HandleNodeSSHKeys       bool    `mapstructure:"handleNodeSshKeys"`
	MaxNodes                int     `mapstructure:"maxNodes"`
	MaxNodeMemory           string  `mapstructure:"maxNodeMemory"`
	MaxNodeCPU              float64 `mapstructure:"maxNodeCpu"`
	BridgePrefix            string  `mapstructure:"bridgePrefix"`
	LogJSON                 bool    `mapstructure:"logJson"`
	PrometheusConfig        string  `mapstructure:"prometheusConfig"`
	PrometheusPort          int     `mapstructure:"prometheusPort"`
	GanacheCLIOptions       string  `mapstructure:"ganacheCLIOptions"`
	GanacheRPCPort          int     `mapstructure:"ganacheRPCPort"`
	MaxRunAttempts          int     `mapstructure:"maxRunAttempts"`
	MaxConnections          int     `mapstructure:"maxConnections"`
	DataDirectory           string  `mapstructure:"datadir"`
	DisableTestnetReporting bool    `mapstructure:"disableTestnetReporting"`
	RequireAuth             bool    `mapstructure:"requireAuth"`
	MaxCommandOutputLogSize int     `mapstructure:"maxCommandOutputLogSize"`
	ResourceDir             string  `mapstructure:"resourceDir"`
	RemoveNodesOnFailure    bool    `mapstructure:"removeNodesOnFailure"`
	KillRetries             uint    `mapstructure:"killRetries"`
	EnablePortForwarding    bool    `mapstructure:"enablePortForwarding"`
	EnableDockerVolumes     bool    `mapstructure:"enableDockerVolumes"`
	EnableImageBuilding     bool    `mapstructure:"enableImageBuilding"`
	EnableDockerLogging     bool    `mapstructure:"enableDockerLogging"`
	TmpStoreDir             string  `mapstructure:"tmpStoreDir"`
}

//GetLogsOutputFile returns the path to the file where the nodes should output their logs
func (c Config) GetLogsOutputFile() string {
	return c.DockerOutputFile
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
	viper.BindEnv("logJson", "LOG_JSON")
	viper.BindEnv("prometheusConfig", "PROMETHEUS_CONFIG")
	viper.BindEnv("prometheusPort", "PROMETHEUS_PORT")
	viper.BindEnv("ganacheCLIOptions", "GANACHE_CLI_OPTIONS")
	viper.BindEnv("ganacheRPCPort", "GANACHE_RPC_PORT")
	viper.BindEnv("maxRunAttempts", "MAX_RUN_ATTEMPTS")
	viper.BindEnv("maxConnections", "MAX_CONNECTIONS")
	viper.BindEnv("datadir", "DATADIR")
	viper.BindEnv("disableTestnetReporting", "DISABLE_TESTNET_REPORTING")
	viper.BindEnv("requireAuth", "REQUIRE_AUTH")
	viper.BindEnv("maxCommandOutputLogSize", "MAX_COMMAND_OUTPUT_LOG_SIZE")
	viper.BindEnv("resourceDir", "RESOURCE_DIR")
	viper.BindEnv("removeNodesOnFailure", "REMOVE_NODES_ON_FAILURE")
	viper.BindEnv("killRetries", "KILL_RETRIES")
	viper.BindEnv("enablePortForwarding", "ENABLE_PORT_FORWARDING")
	viper.BindEnv("enableDockerVolumes", "ENABLE_DOCKER_VOLUMES")
	viper.BindEnv("enableImageBuilding", "ENABLE_IMAGE_BUILDING")
	viper.BindEnv("enableDockerLogging", "ENABLE_DOCKER_LOGGING")
	viper.BindEnv("tmpStoreDir", "TMP_STORE_DIR")
}
func setViperDefaults() {
	viper.SetDefault("sshUser", os.Getenv("USER"))
	viper.SetDefault("sshKey", os.Getenv("HOME")+"/.ssh/id_rsa")
	viper.SetDefault("sshHost", "127.0.0.1")
	viper.SetDefault("serverBits", 8)
	viper.SetDefault("clusterBits", 12)
	viper.SetDefault("nodeBits", 4)
	viper.SetDefault("ipPrefix", 10)
	viper.SetDefault("listen", "127.0.0.1:8000")
	viper.SetDefault("verbosity", "INFO")
	viper.SetDefault("dockerOutputFile", "/output.log")
	viper.SetDefault("serviceNetwork", "172.30.0.1/16")
	viper.SetDefault("serviceNetworkName", "wb_builtin_services")
	viper.SetDefault("nodePrefix", "whiteblock-node")
	viper.SetDefault("nodeNetworkPrefix", "wb_vlan")
	viper.SetDefault("servicePrefix", "wb_service")
	viper.SetDefault("maxNodes", 200)
	viper.SetDefault("maxNodeMemory", "")
	viper.SetDefault("maxNodeCpu", -1)
	viper.SetDefault("bridgePrefix", "wb_bridge")
	viper.SetDefault("logJson", false)
	viper.SetDefault("prometheusConfig", "/tmp/prometheus.yml")
	viper.SetDefault("prometheusPort", 9090)
	viper.SetDefault("prometheusInstrumentationPort", 8008)
	viper.SetDefault("maxRunAttempts", 30)
	viper.SetDefault("maxConnections", 50)
	viper.SetDefault("datadir", os.Getenv("HOME")+"/.config/whiteblock/")
	viper.SetDefault("disableTestnetReporting", false)
	viper.SetDefault("requireAuth", false)
	viper.SetDefault("maxCommandOutputLogSize", -1)
	viper.SetDefault("resourceDir", "./resources")
	viper.SetDefault("removeNodesOnFailure", true)
	viper.SetDefault("killRetries", 100)
	viper.SetDefault("ganacheRPCPort", 8545)
	viper.SetDefault("ganacheCLIOptions", "--gasLimit 4000000000000")
	viper.SetDefault("enablePortForwarding", true)
	viper.SetDefault("enableDockerVolumes", true)
	viper.SetDefault("enableImageBuilding", true)
	viper.SetDefault("enableDockerLogging", true)
	viper.SetDefault("tmpStoreDir", "/tmp")
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
