package main

import (
	rest "./rest"
	util "./util"
	"log"
)

var conf *util.Config

func main() {
	conf = util.GetConfig()
	log.SetFlags(log.LstdFlags | log.Llongfile)
	rest.StartServer()
}
