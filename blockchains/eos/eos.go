package eos

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"strings"
	"math/rand"
	"sync"
	db "../../db"
	util "../../util"
)

var conf *util.Config

func init(){
	conf = util.GetConfig()
}

/**
 * Setup the EOS test net
 * @param  int		nodes		The number of producers to make
 * @param  []Server servers		The list of relevant servers
 */
func Eos(data map[string]interface{},nodes int,servers []db.Server) ([]string,error) {
	
	eosconf,err := NewConf(data)
	if err != nil {
		return nil,err
	}
	eosconf.BlockProducers++

	fmt.Println("-------------Setting Up EOS-------------")
	sem := semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()

	masterIP := servers[0].Ips[0]
	masterServerIP := servers[0].Addr

	clientPasswords := make(map[string]string)

	fmt.Println("\n*** Get Key Pairs ***")

	

	contractAccounts := []string{
		"eosio.bpay",
		"eosio.msig",
		"eosio.names",
		"eosio.ram",
		"eosio.ramfee",
		"eosio.saving",
		"eosio.stake",
		"eosio.token",
		"eosio.vpay",
	}

	keyPairs := eos_getKeyPairs(servers)

	contractKeyPairs := eos_getContractKeyPairs(servers,contractAccounts)

	masterKeyPair := keyPairs[servers[0].Ips[0]]

	var accountNames []string
	for i := 0; i < int(eosconf.UserAccounts); i++{
		accountNames = append(accountNames,eos_getRegularName(i))
	}
	accountKeyPairs := eos_getUserAccountKeyPairs(masterServerIP,accountNames)

	util.Write("genesis.json", eosconf.GenerateGenesis(keyPairs[masterIP].PublicKey))
	util.Write("config.ini", eosconf.GenerateConfig())
	
	{
		for _, server := range servers {
			for i,ip := range server.Ips {
				/**Start keosd**/
				util.SshExec(server.Addr,fmt.Sprintf("docker exec -d whiteblock-node%d keosd --http-server-address 0.0.0.0:8900",i))
				clientPasswords[ip] = eos_createWallet(server.Addr, i)
				sem.Acquire(ctx,1)

				go func(serverIP string,accountKeyPairs map[string]util.KeyPair,accountNames []string,i int){
					defer sem.Release(1)
					for _,name := range  accountNames {
						util.SshExec(serverIP, fmt.Sprintf("docker exec whiteblock-node%d cleos wallet import --private-key %s", 
							i,accountKeyPairs[name].PrivateKey))
					}
					
				}(server.Addr,accountKeyPairs,accountNames,i)

			}
		}
		sem.Acquire(ctx,conf.ThreadLimit)
		sem.Release(conf.ThreadLimit)
	}
	password := clientPasswords[servers[0].Ips[0]]
	passwordNormal := clientPasswords[servers[0].Ips[1]]

	{
		
		node := 0
		for _, server := range servers {
			sem.Acquire(ctx,1)
			go func(serverIP string, ips []string){
				defer sem.Release(1)
				util.Scp(serverIP, "./genesis.json", "/home/appo/genesis.json")
				util.Scp(serverIP, "./config.ini", "/home/appo/config.ini")

				for i := 0; i < len(ips); i++ {
					util.SshExecIgnore(serverIP, fmt.Sprintf("docker exec whiteblock-node%d mkdir /datadir/", i))
					util.SshExec(serverIP, fmt.Sprintf("docker cp /home/appo/genesis.json whiteblock-node%d:/datadir/", i))
					util.SshExec(serverIP, fmt.Sprintf("docker cp /home/appo/config.ini whiteblock-node%d:/datadir/", i))
					node++
				}
				util.SshExec(serverIP,fmt.Sprintf("rm /home/appo/genesis.json"))
				util.SshExec(serverIP,fmt.Sprintf("rm /home/appo/config.ini"))
			}(server.Addr,server.Ips)
		}
		sem.Acquire(ctx,conf.ThreadLimit)
		sem.Release(conf.ThreadLimit)
	}
	defer func(){
		util.Rm("./genesis.json")
		util.Rm("./config.ini")
	}()

	/**Step 2d**/
	{
		fmt.Println(
			util.SshExec(masterServerIP, fmt.Sprintf("docker exec whiteblock-node0 cleos wallet import --private-key %s", 
				keyPairs[masterIP].PrivateKey)))

		fmt.Println(
			util.SshExec(masterServerIP,
				fmt.Sprintf(`docker exec -d whiteblock-node0 nodeos -e -p eosio --genesis-json /datadir/genesis.json --config-dir /datadir --data-dir /datadir %s %s`,
					eos_getKeyPairFlag(keyPairs[masterIP]),
					eos_getPTPFlags(servers, 0))))
	}
	

	/**Step 3**/
	{
		util.SshExecIgnore(masterServerIP, fmt.Sprintf("docker exec whiteblock-node0 cleos -u http://%s:8889 wallet unlock --password %s",
			masterIP, password))
		for _, account := range contractAccounts {
			sem.Acquire(ctx,1)
			go func(masterServerIP string,masterIP string,account string,masterKeyPair util.KeyPair,contractKeyPair util.KeyPair){
				defer sem.Release(1)
				
				
				util.SshExec(masterServerIP, fmt.Sprintf("docker exec whiteblock-node0 cleos wallet import --private-key %s", 
					contractKeyPair.PrivateKey))
				
				util.SshExec(masterServerIP, fmt.Sprintf("docker exec whiteblock-node0 cleos -u http://%s:8889 create account eosio %s %s %s",
					 masterIP, account,masterKeyPair.PublicKey,contractKeyPair.PublicKey))

				//log.Println("Finished creating account for "+account)
			}(masterServerIP,masterIP,account,masterKeyPair,contractKeyPairs[account])

		}
		sem.Acquire(ctx,conf.ThreadLimit)
		sem.Release(conf.ThreadLimit)
		
	}
	/**Steps 4 and 5**/
	{
		contracts := []string{"eosio.token","eosio.msig"}
		util.SshExecIgnore(masterServerIP, fmt.Sprintf("docker exec whiteblock-node0 cleos -u http://%s:8889 wallet unlock --password %s",
				masterIP, password))
		for _, contract := range contracts {
			
			fmt.Println(util.SshExec(masterServerIP, fmt.Sprintf("docker exec whiteblock-node0 cleos -u http://%s:8889 set contract %s /opt/eosio/contracts/%s",
				masterIP, contract, contract)))
		}
	}
	/**Step 6**/

	fmt.Println(util.SshExec(masterServerIP, fmt.Sprintf("docker exec whiteblock-node0 cleos -u http://%s:8889 push action eosio.token create '[ \"eosio\", \"10000000000.0000 SYS\" ]' -p eosio.token@active",
		masterIP)))

	fmt.Println(util.SshExec(masterServerIP, fmt.Sprintf("docker exec whiteblock-node0 cleos -u http://%s:8889 push action eosio.token issue '[ \"eosio\", \"1000000000.0000 SYS\", \"memo\" ]' -p eosio@active",
		masterIP)))


	util.SshExecIgnore(masterServerIP, fmt.Sprintf("docker exec whiteblock-node0 cleos -u http://%s:8889 wallet unlock --password %s",
				masterIP, password))


	/**Step 7**/
	for i := 0 ; i < 5; i++{
		res, err := util.SshExecCheck(masterServerIP, fmt.Sprintf("docker exec whiteblock-node0 cleos -u http://%s:8889 set contract -x 1000 eosio /opt/eosio/contracts/eosio.system",
		masterIP))
		if(err == nil){
			fmt.Println("SUCCESS!!!!!")
			fmt.Println(res)
			break
		}
		fmt.Println(res)
	}
	
	
	/**Step 8**/

	fmt.Println(
		util.SshExec(masterServerIP,
			fmt.Sprintf(`docker exec whiteblock-node0 cleos -u http://%s:8889 push action eosio setpriv '["eosio.msig", 1]' -p eosio@active`,
				masterIP)))

	/**Step 10a**/
	{
		node := 0
		for _, server := range servers {
			for _, ip := range server.Ips {
				
				if node == 0 {
					node++
					continue
				}
				sem.Acquire(ctx,1)
				go func(masterServerIP string, masterKeyPair util.KeyPair, keyPair util.KeyPair,node int){
					defer sem.Release(1)
					fmt.Println(
						util.SshExec(masterServerIP,
							fmt.Sprintf("docker exec whiteblock-node0 cleos wallet import --private-key %s",
								keyPair.PrivateKey)))
					if node < eosconf.BlockProducers {
						fmt.Println(
							util.SshExec(masterServerIP,
								fmt.Sprintf(`docker exec whiteblock-node0 cleos -u http://%s:8889 system newaccount eosio --transfer %s %s %s --stake-net "1000000.0000 SYS" --stake-cpu "1000000.0000 SYS" --buy-ram "1000000 SYS"`,
									masterIP,
									eos_getProducerName(node),
									masterKeyPair.PublicKey,
									keyPair.PublicKey)))
						fmt.Println(
							util.SshExec(masterServerIP,
								fmt.Sprintf(`docker exec whiteblock-node0 cleos -u http://%s:8889 transfer eosio %s "100000.0000 SYS"`,
									masterIP,
									eos_getProducerName(node))))
					}
					
				}(masterServerIP,masterKeyPair,keyPairs[ip],node)
				node++
			}
		}
		sem.Acquire(ctx,conf.ThreadLimit)
		sem.Release(conf.ThreadLimit)
	}
	
	/**Step 11c**/
	{
		node := 0
		for _, server := range servers {
			for i, ip := range server.Ips {
				
				if node == 0 {
					node++
					continue
				}
				sem.Acquire(ctx,1)

				go func(serverIP string,servers []db.Server,node int,i int,kp util.KeyPair){
					defer sem.Release(1)
					util.SshExecIgnore(serverIP, fmt.Sprintf("docker exec whiteblock-node%d mkdir -p /datadir/blocks", i))
					p2pFlags := eos_getPTPFlags(servers,node)
					if node > eosconf.BlockProducers {
						fmt.Println(
							util.SshExec(serverIP,
								fmt.Sprintf(`docker exec -d whiteblock-node%d nodeos --genesis-json /datadir/genesis.json --config-dir /datadir --data-dir /datadir %s %s`,
									i,
									eos_getKeyPairFlag(kp),
									p2pFlags)))
					}else{
						fmt.Println(
							util.SshExec(serverIP,
								fmt.Sprintf(`docker exec -d whiteblock-node%d nodeos --genesis-json /datadir/genesis.json --config-dir /datadir --data-dir /datadir -p %s %s %s`,
									i,
									eos_getProducerName(node),
									eos_getKeyPairFlag(kp),
									p2pFlags)))
					}
					
				}(server.Addr,servers,node,i,keyPairs[ip])
				node++
			}
		}
		sem.Acquire(ctx,conf.ThreadLimit)
		sem.Release(conf.ThreadLimit)
	}

	/**Step 11a**/
	{
		node := 0
		for _, server := range servers {
			for _, ip := range server.Ips {
				
				if node == 0 {
					node++
					continue
				}else if node >= eosconf.BlockProducers {
					break
				}

				if node % 5 == 0{
					util.SshExecIgnore(masterServerIP, fmt.Sprintf("docker exec whiteblock-node0 cleos -u http://%s:8889 wallet unlock --password %s",
						masterIP, password))
				}

				fmt.Println(
					util.SshExec(masterServerIP,
						fmt.Sprintf("docker exec whiteblock-node0 cleos --wallet-url http://%s:8900 -u http://%s:8889 system regproducer %s %s https://whiteblock.io/%s",
							masterIP,
							masterIP,
							eos_getProducerName(node),
							keyPairs[ip].PublicKey,
							keyPairs[ip].PublicKey)))
				
				node++
			}
		}
	}
	/**Step 11b**/
	fmt.Println(util.SshExec(masterServerIP,
						fmt.Sprintf("docker exec whiteblock-node0 cleos -u http://%s:8889 system listproducers",
							masterIP)))


	/**Create normal user accounts**/
	{
		for _, name := range accountNames {
			sem.Acquire(ctx,1)
			go func(masterServerIP string,name string,masterKeyPair util.KeyPair,accountKeyPair util.KeyPair){
				defer sem.Release(1)
				fmt.Println(
					util.SshExec(masterServerIP,
						fmt.Sprintf(`docker exec whiteblock-node0 cleos -u http://%s:8889 system newaccount eosio --transfer %s %s %s --stake-net "500000.0000 SYS" --stake-cpu "2000000.0000 SYS" --buy-ram "2000000 SYS"`,
							masterIP,
							name,
							masterKeyPair.PublicKey,
							accountKeyPair.PublicKey)))

				fmt.Println(
					util.SshExec(masterServerIP,
						fmt.Sprintf(`docker exec whiteblock-node0 cleos -u http://%s:8889 transfer eosio %s "100000.0000 SYS"`,
							masterIP,
							name)))
			}(masterServerIP,name,masterKeyPair,accountKeyPairs[name])
		}
		sem.Acquire(ctx,conf.ThreadLimit)
		sem.Release(conf.ThreadLimit)
	}
	/**Vote in block producers**/
	{	
		node := 0
		for _, server := range servers {
			for range server.Ips {			
				node++
			}
		}
		if(node > eosconf.BlockProducers){
			node = eosconf.BlockProducers
		}
		util.SshExecIgnore(masterServerIP, fmt.Sprintf("docker exec whiteblock-node1 cleos -u http://%s:8889 wallet unlock --password %s",
				masterIP, passwordNormal))
		n := 0
		for _, name := range accountNames {
			prod := 0
			if n != 0 {
				prod = rand.Intn(100) % n
			} 
		
			prod = prod % (node - 1)
			prod += 1
			sem.Acquire(ctx,1)
			go func(masterServerIP string,masterIP string,name string,prod int){
				defer sem.Release(1)
				fmt.Println(
						util.SshExec(masterServerIP,
							fmt.Sprintf("docker exec whiteblock-node1 cleos -u http://%s:8889 system voteproducer prods %s %s",
								masterIP,
								name,
								eos_getProducerName(prod))))
			}(masterServerIP,masterIP,name,prod)
			n++;
		}
		sem.Acquire(ctx,conf.ThreadLimit)
		sem.Release(conf.ThreadLimit)
	}
	
	/**Step 12**/
	fmt.Println(
		util.SshExec(masterServerIP,
			fmt.Sprintf(
				`docker exec whiteblock-node0 cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio.prods", "permission": "active"}}]}}' -p eosio@owner`,
				masterIP)))

	fmt.Println(
		util.SshExec(masterServerIP,
			fmt.Sprintf(
				`docker exec whiteblock-node0 cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio.prods", "permission": "active"}}]}}' -p eosio@active`,
				masterIP)))

	
	fmt.Println(
		util.SshExec(masterServerIP,
			fmt.Sprintf(
				`docker exec whiteblock-node0 cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.bpay", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.bpay@owner`,
				masterIP)))
	fmt.Println(
		util.SshExec(masterServerIP,
			fmt.Sprintf(
				`docker exec whiteblock-node0 cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.bpay", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.bpay@active`,
				masterIP)))
	fmt.Println(
		util.SshExec(masterServerIP,
			fmt.Sprintf(
				`docker exec whiteblock-node0 cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.msig", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.msig@owner`,
				masterIP)))
	fmt.Println(
		util.SshExec(masterServerIP,
			fmt.Sprintf(
				`docker exec whiteblock-node0 cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.msig", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.msig@active`,
				masterIP)))
	fmt.Println(
		util.SshExec(masterServerIP,
			fmt.Sprintf(
				`docker exec whiteblock-node0 cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.names", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.names@owner`,
				masterIP)))
	fmt.Println(
		util.SshExec(masterServerIP,
			fmt.Sprintf(
				`docker exec whiteblock-node0 cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.names", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.names@active`,
				masterIP)))
	fmt.Println(
		util.SshExec(masterServerIP,
			fmt.Sprintf(
				`docker exec whiteblock-node0 cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.ram", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.ram@owner`,
				masterIP)))

	fmt.Println(
		util.SshExec(masterServerIP,
			fmt.Sprintf(
				`docker exec whiteblock-node0 cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.ram", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.ram@active`,
				masterIP)))
	fmt.Println(
		util.SshExec(masterServerIP,
			fmt.Sprintf(
				`docker exec whiteblock-node0 cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.ramfee", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.ramfee@owner`,
				masterIP)))
	fmt.Println(
		util.SshExec(masterServerIP,
			fmt.Sprintf(
				`docker exec whiteblock-node0 cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.ramfee", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.ramfee@active`,
				masterIP)))



	out := []string{}

	for _, server := range servers {
		for _, ip := range server.Ips {
			out = append(out,clientPasswords[ip])
		}
	}

	return out,nil
}
func eos_getKeyPair(serverIP string) util.KeyPair {
	data := util.SshExec(serverIP, "docker exec whiteblock-node0 cleos create key --to-console | awk '{print $3}'")
	//fmt.Printf("RAW KEY DATA%s\n", data)
	keyPair := strings.Split(data, "\n")
	if(len(data) < 10){
		fmt.Printf("Unexpected create key output %s\n",keyPair)
		panic(1)
	}
	return util.KeyPair{PrivateKey: keyPair[0], PublicKey: keyPair[1]}
}


func eos_getKeyPairs(servers []db.Server) map[string]util.KeyPair {
	keyPairs := make(map[string]util.KeyPair)
	/**Get the key pairs for each nodeos account**/
	
	var wg sync.WaitGroup
	var mutex = &sync.Mutex{}

	for _, server := range servers {
		wg.Add(1)
		go func(serverIP string,ips []string){
			defer wg.Done()
			for _, ip := range ips {
				data := util.SshExec(serverIP, "docker exec whiteblock-node0 cleos create key --to-console | awk '{print $3}'")
				//fmt.Printf("RAW KEY DATA%s\n", data)
				keyPair := strings.Split(data, "\n")
				if(len(data) < 10){
					fmt.Printf("Unexpected create key output %s\n",keyPair)
					panic(1)
				}
					
				mutex.Lock()
				keyPairs[ip] = util.KeyPair{PrivateKey: keyPair[0], PublicKey: keyPair[1]}
				mutex.Unlock()
			}
		}(server.Addr,server.Ips)
	}
	wg.Wait()
	
	return keyPairs
}


func eos_getContractKeyPairs(servers []db.Server,contractAccounts []string) map[string]util.KeyPair {

	keyPairs := make(map[string]util.KeyPair)
	server := servers[0]

	for _,contractAccount := range contractAccounts {
		
		keyPairs[contractAccount] = eos_getKeyPair(server.Addr)
	}
	return keyPairs
}

func eos_getUserAccountKeyPairs(masterServerIP string,accountNames []string) map[string]util.KeyPair {

	keyPairs := make(map[string]util.KeyPair)
	sem := semaphore.NewWeighted(conf.ThreadLimit)
	var mutex = &sync.Mutex{}
	ctx := context.TODO()

	for _,name := range accountNames {
		sem.Acquire(ctx,1)
		go func(serverIP string,name string){
			defer sem.Release(1)
			data := util.SshExec(serverIP, "docker exec whiteblock-node0 cleos create key --to-console | awk '{print $3}'")
			//fmt.Printf("RAW KEY DATA%s\n", data)
			keyPair := strings.Split(data, "\n")
			if(len(data) < 10){
				fmt.Printf("Unexpected create key output %s\n",keyPair)
				panic(1)
			}
			mutex.Lock()
			keyPairs[name] = util.KeyPair{PrivateKey: keyPair[0], PublicKey: keyPair[1]}
			mutex.Unlock()
		}(masterServerIP,name)
	}
	sem.Acquire(ctx,conf.ThreadLimit)
	sem.Release(conf.ThreadLimit)
	return keyPairs
}


func eos_createWallet(serverIP string, id int) string {
	data := util.SshExec(serverIP, fmt.Sprintf("docker exec whiteblock-node%d cleos wallet create --to-console | tail -n 1",id))
	//fmt.Printf("CREATE WALLET DATA %s\n",data)
	offset := 0
	for data[len(data) - (offset + 1)] != '"' {
		offset++
	}
	offset++
	data = data[1 : len(data) - offset]
	fmt.Printf("CREATE WALLET DATA %s\n",data)
	return data
}

func eos_getKeyPairFlag(keyPair util.KeyPair) string {
	return fmt.Sprintf("--private-key '[ \"%s\",\"%s\" ]'", keyPair.PublicKey, keyPair.PrivateKey)
}

func eos_getProducerName(num int) string {
	if num == 0 {
		return "eosio"
	}
	out := ""

	for i := num; i > 0; i = (i - (i % 4)) / 4{
		place := i % 4
		place++
		out = fmt.Sprintf("%d%s",place,out)//I hate this
	}
	for i := len(out); i < 5; i++ {
		out = "x"+out
	}

	return "prod"+out
}

func eos_getRegularName(num int) string {

	out := ""
	//num -= blockProducers

	for i := num; i > 0; i = (i - (i % 4)) / 4{
		place := i % 4
		place++
		out = fmt.Sprintf("%d%s",place,out)//I hate this
	}
	for i := len(out); i < 8; i++ {
		out = "x"+out
	}

	return "user"+out
}


func eos_getPTPFlags(servers []db.Server, exclude int) string {
	flags := ""
	node := 0
	for _, server := range servers {
		for _, ip := range server.Ips {
			if(node == exclude){
				node++
				continue
			}
			flags += fmt.Sprintf("--p2p-peer-address %s:8999 ", ip)

		}
	}
	return flags
}
