package main

import (
	"log"
	util "./util"
    rest "./rest"
)

var conf *util.Config

func main() {
	conf = util.GetConfig()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	rest.StartServer()
}
