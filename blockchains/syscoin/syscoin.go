package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
)

var (
	sem3 = semaphore.NewWeighted(THREAD_LIMIT)
)

func syscoinRegTest(nodes int,servers []Server){
	ctx := context.TODO()
	fmt.Println("-------------Setting Up Syscoin-------------")
	fmt.Printf("Creating the syscoin conf file...")
	createSyscoinConf(servers)
	fmt.Printf("done\n")
	fmt.Printf("Launching the nodes")
	for _,server := range servers {
		sem3.Acquire(ctx,1)
		go setupSyscoinRegTest(server)
	}

	err := sem3.Acquire(ctx,THREAD_LIMIT)
	check_fatal(err)
	fmt.Printf("done\n")
	sem3.Release(THREAD_LIMIT)


	fmt.Printf("Cleaning up...")
	rm("config.boot")
	rm("regtest.conf")
	fmt.Printf("done\n")
}

func createSyscoinConf(servers []Server){
	confData := "rpcuser=appo\nrpcpassword=w@ntest\nrpcport=8369\nserver=1\nregtest=1\nport=8370\nlisten=1\nrest=1\ndebug=1\nunittest=1\naddressindex=1\nassetallocationindex=1\ntpstest=1\n"
	maxConns := 1
	for _,server := range servers {
		for _,ip := range server.ips {
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
		confData += fmt.Sprintf("connect=%s:8370\n",server.iaddr.ip)
		//confData += fmt.Sprintf("rpcconnect=%s:8369\n",server.iaddr.ip)
		maxConns += 4
	}
	confData += "rpcallowip=10.0.0.0/8\n"
	confData += fmt.Sprintf("maxconnections=%d\n",maxConns)
	write("./regtest.conf",confData)
}

func setupSyscoinRegTest(server Server){
	sshExecIgnore(server.addr,"mkdir ~/datadir")
	_scp(server.addr,"./regtest.conf","/home/appo/regtest.conf")
	sshExec(server.addr,"rm -rf ~/datadir && mkdir -p ~/datadir")
	sshExec(server.addr,"cp /home/appo/regtest.conf /home/appo/datadir/regtest.conf")
	sshExecIgnore(server.addr,"killall syscoind")
	sshExec(server.addr,"syscoind -daemon -conf=\"/home/appo/datadir/regtest.conf\" -datadir=\"/home/appo/datadir/\"")
	for j,_ := range server.ips {
		fmt.Printf(".")
		container := fmt.Sprintf("whiteblock-node%d",j)
		mkdirCmd := fmt.Sprintf("docker exec %s mkdir -p /syscoin && docker exec %s mkdir -p /syscoin/datadir",container,container)
		cpCmd := fmt.Sprintf("docker cp /home/appo/regtest.conf %s:/syscoin/datadir/regtest.conf",container)
		execCmd := fmt.Sprintf("docker exec %s syscoind -daemon -conf=\"/syscoin/datadir/regtest.conf\" -datadir=\"/syscoin/datadir/\"",container)
		sshExec(server.addr,
			fmt.Sprintf("%s&&%s&&%s",mkdirCmd,cpCmd,execCmd))
	}
	sem3.Release(1)
}