package util

import(
    "encoding/json"
    "io/ioutil"
    "log"
    "os"
    "strconv"
    "fmt"
)

type Config struct {
    Builder             string      `json:"builder"`
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
    AllowExec           bool        `json:"allow-exec"`
    DockerOutputFile    string      `json:"docker-output-file"`

    Influx              string      `json:"influx"`
    InfluxUser          string      `json:"influx-user"`
    InfluxPassword      string      `json:"influx-password"`
}

func (this *Config) LoadFromEnv() {
    var err error
    val,exists := os.LookupEnv("BUILDER")
    if exists {
        this.Builder = val
    }
    val,exists = os.LookupEnv("SSH_USER")
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
    _,exists = os.LookupEnv("ALLOW_EXEC")
    if exists {
        this.AllowExec = true
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
}

func (c *Config) AutoFillMissing() {
    if len(c.Builder) == 0 {
        c.Builder = "local deploy"
    }
    if len(c.SshUser) == 0 {
        c.SshUser = "appo"
    }
    if len(c.SshPassword) == 0 {
        c.SshPassword = "w@ntest"
    }
    if len(c.VyosHomeDir) == 0 {
        c.VyosHomeDir = "/home/appo/"
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
        println("Warning: Using default server bits")
        c.ServerBits = 8
    }
    if c.ClusterBits <= 0 {
        println("Warning: Using default cluster bits")
        c.ClusterBits = 14
    }
    if c.NodeBits <= 0 {
        println("Warning: Using default node bits")
        c.NodeBits = 2
    }
    if c.ThreadLimit <= 0 {
        println("Warning: Using default thread limit")
        c.ThreadLimit = 10
    }
    if c.AllowExec {
        println("Warning: exec call is enabled. This is unsafe!")
    }
    if len(c.DockerOutputFile) == 0 {
        c.DockerOutputFile = "/output.log"
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
    }else{
        json.Unmarshal(dat,conf)
    }

    return conf
}


func GetConfig() *Config {
    if(conf == nil){
        LoadConfig()
    }
    return conf
}