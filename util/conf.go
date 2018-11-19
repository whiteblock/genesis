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
} 

var conf *Config = nil


func init() {
	LoadConfig()
	conf.AutoFillMissing()
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