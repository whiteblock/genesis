package syscoin

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	util "../../util"
	db "../../db"
	state "../../state"
)

var conf *util.Config

func init(){
	conf = util.GetConfig()
}

func RegTest(data map[string]interface{},nodes int,servers []db.Server) error {
	sem3 := semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()
	sysconf,err := NewConf(data)
	if err != nil {
		return err
	}
	defer func(){
		fmt.Printf("Cleaning up...")
		util.Rm("config.boot")
		util.Rm("regtest.conf")
		fmt.Printf("done\n")
	}()
	state.SetBuildSteps(1+(1*nodes))

	fmt.Println("-------------Setting Up Syscoin-------------")
	{
		fmt.Printf("Creating the syscoin conf file...")
		createSyscoinConf(servers,sysconf)
		state.IncrementBuildProgress()
		fmt.Printf("done\n")
	}

	fmt.Printf("Launching the nodes")
	for _,server := range servers {
		sem3.Acquire(ctx,1)
		go func(server db.Server){
			util.SshExecIgnore(server.Addr,"mkdir ~/datadir")
			util.Scp(server.Addr,"./regtest.conf","/home/appo/regtest.conf")
			util.SshExec(server.Addr,"rm -rf ~/datadir && mkdir -p ~/datadir")
			util.SshExec(server.Addr,"cp /home/appo/regtest.conf /home/appo/datadir/regtest.conf")
			util.SshExecIgnore(server.Addr,"killall syscoind")
			util.SshExec(server.Addr,"syscoind -daemon -conf=\"/home/appo/datadir/regtest.conf\" -datadir=\"/home/appo/datadir/\"")
			for j,_ := range server.Ips {
				fmt.Printf(".")
				container := fmt.Sprintf("whiteblock-node%d",j)
				mkdirCmd := fmt.Sprintf("docker exec %s mkdir -p /syscoin && docker exec %s mkdir -p /syscoin/datadir",container,container)
				cpCmd := fmt.Sprintf("docker cp /home/appo/regtest.conf %s:/syscoin/datadir/regtest.conf",container)
				execCmd := fmt.Sprintf("docker exec %s syscoind -daemon -conf=\"/syscoin/datadir/regtest.conf\" -datadir=\"/syscoin/datadir/\"",container)
				util.SshExec(server.Addr,
					fmt.Sprintf("%s&&%s&&%s",mkdirCmd,cpCmd,execCmd))
				state.IncrementBuildProgress()
			}
			sem3.Release(1)

		}(server)
	}

	err = sem3.Acquire(ctx,conf.ThreadLimit)
	if err != nil{
		return err
	}
	fmt.Printf("done\n")
	sem3.Release(conf.ThreadLimit)


	return nil 
}

func createSyscoinConf(servers []db.Server, sysconf *SysConf){
	confData := sysconf.Generate()
	confData += "rpcport=8369\nport=8370\n"

	maxConns := 1
	for _,server := range servers {
		for _,ip := range server.Ips {
			//confData += fmt.Sprintf("connect=%s:8370\n",ip)
			//confData += fmt.Sprintf("whitelist=%s\n",ip)
			confData += fmt.Sprintf("connect=%s:8370\n",ip)
			//confData += fmt.Sprintf("rpcconnect=%s:8369\n",ip)
			maxConns += 4
		}
	}
	for _,server := range servers {
		//confData += fmt.Sprintf("connect=%s:8370\n",server.iaddr.ip)
		//confData += fmt.Sprintf("whitelist=%s\n",server.iaddr.ip)
		confData += fmt.Sprintf("connect=%s:8370\n",server.Iaddr.Ip)
		//confData += fmt.Sprintf("rpcconnect=%s:8369\n",server.iaddr.ip)
		maxConns += 4
	}
	confData += "rpcallowip=10.0.0.0/8\n"
	confData += fmt.Sprintf("maxconnections=%d\n",maxConns)
	util.Write("./regtest.conf",confData)
}

