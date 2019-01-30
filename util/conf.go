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
    SshUser             string      `json:"ssh-user"`
    SshPassword         string      `json:"ssh-password"`
    VyosHomeDir         string      `json:"vyos-home-dir"`
    Listen              string      `json:"listen"`
    RsaKey              string      `json:"rsa-key"`
    RsaUser             string      `json:"rsa-user"`
    Verbose             bool        `json:"verbose"`
    ServerBits          uint32      `json:"server-bits"`
    ClusterBits         uint32      `json:"cluster-bits"`
    NodeBits            uint32      `json:"node-bits"`
    ThreadLimit         int64       `json:"thread-limit"`
    BuildMode           string      `json:"build-mode"`
    IPPrefix            uint32      `json:"ip-prefix"`
    DockerOutputFile    string      `json:"docker-output-file"`
    Influx              string      `json:"influx"`
    InfluxUser          string      `json:"influx-user"`
    InfluxPassword      string      `json:"influx-password"`
    ServiceVlan         int         `json:"service-vlan"`
    ServiceNetwork      string      `json:"service-network"`
    ServiceNetworkName  string      `json:"service-network-name`
    NodePrefix          string      `json:"node-prefix"`
    NodeNetworkPrefix   string      `json:"node-network-prefix"`
    ServicePrefix       string      `json:"service-prefix"`
    NetworkVlanStart    int         `json:"network-vlan-start"`
    SetupMasquerade     bool        `json:"setup-masquerade"`

    NodesPublicKey      string      `json:"nodes-public-key"`
    NodesPrivateKey     string      `json:"nodes-private-key"`
    HandleNodeSshKeys   bool        `json:"handle-node-ssh-keys"`

    MaxNodes            int         `json:"max-nodes"`
    MaxNodeMemory       string      `json:"max-node-memory"`
    MaxNodeCpu          float64     `json:"max-node-cpu"`

    NeoBuild            bool        `json:"neo-build"`
    BridgePrefix        string      `json:"bridge-prefix"`            
}

func (this *Config) LoadFromEnv() {
    var err error
    val,exists := os.LookupEnv("SSH_USER")
    if exists {
        this.SshUser = val
    }
    val,exists = os.LookupEnv("SSH_PASSWORD")
    if exists {
        this.SshPassword = val
    }
    val,exists = os.LookupEnv("VYOS_HOME_DIR")
    if exists {
        this.VyosHomeDir = val
    }
    val,exists = os.LookupEnv("LISTEN")
    if exists {
        this.Listen = val
    }
    val,exists = os.LookupEnv("RSA_KEY")
    if exists {
        this.RsaKey = val
    }
    val,exists = os.LookupEnv("RSA_USER")
    if exists {
        this.RsaUser = val
    }
    _,exists = os.LookupEnv("VERBOSE")
    if exists {
        this.Verbose = true
    }
    val,exists = os.LookupEnv("SERVER_BITS")
    if exists {
        tmp,err := strconv.ParseUint(val,0,32)
        this.ServerBits = uint32(tmp)
        if err != nil{
            fmt.Println("Invalid ENV value for SERVER_BITS")
            os.Exit(1)
        }
    }
    val,exists = os.LookupEnv("CLUSTER_BITS")
    if exists {
        tmp,err := strconv.ParseUint(val,0,32)
        this.ClusterBits = uint32(tmp)
        if err != nil{
            fmt.Println("Invalid ENV value for CLUSTER_BITS")
            os.Exit(1)
        }
    }
    val,exists = os.LookupEnv("NODE_BITS")
    if exists {
        tmp,err := strconv.ParseUint(val,0,32)
        this.NodeBits = uint32(tmp)
        if err != nil{
            fmt.Println("Invalid ENV value for NODE_BITS")
            os.Exit(1)
        }
    }
    val,exists = os.LookupEnv("THREAD_LIMIT")
    if exists {
        this.ThreadLimit,err = strconv.ParseInt(val,0,64)
        if err != nil{
            fmt.Println("Invalid ENV value for THREAD_LIMIT")
            os.Exit(1)
        }
    }
    val,exists = os.LookupEnv("BUILD_MODE")
    if exists {
        this.BuildMode = val
    }
    val,exists = os.LookupEnv("IP_PREFIX")
    if exists {
        tmp,err := strconv.ParseUint(val,0,32)
        this.IPPrefix = uint32(tmp)
        if err != nil{
            fmt.Println("Invalid ENV value for IP_PREFIX")
            os.Exit(1)
        }
    }
    val,exists = os.LookupEnv("DOCKER_OUTPUT_FILE")
    if exists {
        this.DockerOutputFile = val
    }
    val,exists = os.LookupEnv("INFLUX")
    if exists {
        this.Influx = val
    }
    val,exists = os.LookupEnv("INFLUX_USER")
    if exists {
        this.InfluxUser = val
    }
    val,exists = os.LookupEnv("INFLUX_PASSWORD")
    if exists {
        this.InfluxPassword = val
    }
    val,exists = os.LookupEnv("SERVICE_VLAN")
    if exists {
        tmp,err := strconv.ParseInt(val,0,32)
        this.ServiceVlan = int(tmp)
        if err != nil{
            fmt.Println("Invalid ENV value for SERVICE_VLAN")
            os.Exit(1)
        }
    }
    val,exists = os.LookupEnv("SERVICE_NETWORK")
    if exists {
        this.ServiceNetwork = val
    }
    val,exists = os.LookupEnv("SERVICE_NETWORK_NAME")
    if exists {
        this.ServiceNetworkName = val
    }
    val,exists = os.LookupEnv("NODE_PREFIX")
    if exists {
        this.NodePrefix = val
    }
    val,exists = os.LookupEnv("NODE_NETWORK_PREFIX")
    if exists {
        this.NodeNetworkPrefix = val
    }
    val,exists = os.LookupEnv("SERVICE_PREFIX")
    if exists {
        this.ServicePrefix = val
    }
    val,exists = os.LookupEnv("NETWORK_VLAN_START")
    if exists {
        tmp,err := strconv.ParseInt(val,0,32)
        this.NetworkVlanStart = int(tmp)
        if err != nil{
            fmt.Println("Invalid ENV value for NETWORK_VLAN_START")
            os.Exit(1)
        }
    }
    _,exists = os.LookupEnv("SETUP_MASQUERADE")
    if exists {
        this.SetupMasquerade = true
    }

    val,exists = os.LookupEnv("NODES_PUBLIC_KEY")
    if exists {
        this.NodesPublicKey = val
    }

    val,exists = os.LookupEnv("NODES_PRIVATE_KEY")
    if exists {
        this.NodesPrivateKey = val
    }

    _,exists = os.LookupEnv("HANDLE_NODES_SSH_KEYS")
    if exists {
        this.HandleNodeSshKeys = true
    }

    val,exists = os.LookupEnv("MAX_NODES")
    if exists {
        tmp,err := strconv.ParseInt(val,0,32)
        this.MaxNodes = int(tmp)
        if err != nil{
            fmt.Println("Invalid ENV value for MAX_NODES")
            os.Exit(1)
        }
    }

    val,exists = os.LookupEnv("MAX_NODE_MEMORY")
    if exists {
        this.MaxNodeMemory = val
    }

    val,exists = os.LookupEnv("MAX_NODE_CPU")
    if exists {
        this.MaxNodeCpu,err = strconv.ParseFloat(val,64)
        if err != nil{
            fmt.Println("Invalid ENV value for MAX_NODE_CPU")
            os.Exit(1)
        }
    }

    _,exists = os.LookupEnv("NEO_BUILD")
    if exists {
        this.NeoBuild = true
    }

    val,exists = os.LookupEnv("BRIDGE_PREFIX")
    if exists {
        this.BridgePrefix = val
    }
}

func (c *Config) AutoFillMissing() {
    if len(c.SshUser) == 0 {
        c.SshUser = "appo"
    }
    if len(c.SshPassword) == 0 {
        c.SshPassword = "w@ntest"
    }
    if len(c.VyosHomeDir) == 0 {
        c.VyosHomeDir = "/home/appo"
    }
    if len(c.Listen) == 0 {
        c.Listen = "127.0.0.1:8000"
    }
    if len(c.RsaKey) == 0 {
        home := os.Getenv("HOME")
        c.RsaKey = home+"/.ssh/id_rsa"
    }
    if len(c.RsaUser) == 0 {
        c.RsaUser = "appo"
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
    if c.AllowExec {
        fmt.Println("Warning: exec call is enabled. This is unsafe!")
    }
    if len(c.DockerOutputFile) == 0 {
        c.DockerOutputFile = "/output.log"
    }

    if c.ServiceVlan == 0 {
        c.ServiceVlan = 4094
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

    if c.NetworkVlanStart <= 0 {
        c.NetworkVlanStart = 101
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

func GetConfig() *Config {
	if conf == nil {
		LoadConfig()
	}
	return conf
}
