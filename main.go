package main

import (
	"log"
	util "./util"
)

var conf *util.Config

func main() {
	conf = util.GetConfig()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	StartServer()
}
