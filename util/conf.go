package util

import(
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Builder 		string		`json:"builder"`
	SshUser			string		`json:"ssh-user"`
	SshPassword		string		`json:"ssh-password"`
	VyosHomeDir		string		`json:"vyos-home-dir"`
	Listen			string		`json:"listen"`
	RsaKey			string		`json:"rsa-key"`
	RsaUser			string		`json:"rsa-user"`
	Verbose			bool		`json:"verbose"`
	ServerBits		uint32		`json:"server-bits"`
	ClusterBits		uint32		`json:"cluster-bits"`
	NodeBits		uint32		`json:"node-bits"`
	ThreadLimit		int64		`json:"thread-limit"`
	BuildMode		string		`json:"build-mode"`
	IPPrefix		uint32		`json:"ip-prefix"`
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
		log.Println("Warning: Using default server bits")
		c.ServerBits = 8
	}
	if c.ClusterBits <= 0 {
		log.Println("Warning: Using default cluster bits")
		c.ClusterBits = 14
	}
	if c.NodeBits <= 0 {
		log.Println("Warning: Using default node bits")
		c.NodeBits = 2
	}
	if c.ThreadLimit <= 0 {
		log.Println("Warning: Using default thread limit")
		c.ThreadLimit = 10
	}
}

var	NodesPerCluster uint32

var conf *Config = nil


func init() {
	LoadConfig()
	conf.AutoFillMissing()
	NodesPerCluster	= (1 << conf.NodeBits) - ReservedIps
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