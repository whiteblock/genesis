package main

import (
	util "./util"
	"log"
)

var conf *util.Config

func main() {
	conf = util.GetConfig()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	StartServer()
}