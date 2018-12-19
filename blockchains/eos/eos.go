package eos

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	//"strings"
	"math/rand"
	//"sync"
	//"errors"
	"log"
	db "../../db"
	util "../../util"
	state "../../state"
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
func Eos(data map[string]interface{},nodes int,servers []db.Server,clients []*util.SshClient) ([]string,error) {
	
	eosconf,err := NewConf(data)
	if err != nil {
		log.Println(err)
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
	km,err := NewKeyMaster()
	if err != nil{
		log.Println(err)
		return nil,err
	}
	keyPairs,err := km.GetServerKeyPairs(servers)
	if err != nil{
		log.Println(err)
		return nil,err
	}

	contractKeyPairs,err := km.GetMappedKeyPairs(contractAccounts)
	if err != nil {
		log.Println(err)
		return nil,err
	}

	masterKeyPair := keyPairs[servers[0].Ips[0]]

	var accountNames []string
	for i := 0; i < int(eosconf.UserAccounts); i++{
		accountNames = append(accountNames,eos_getRegularName(i))
	}
	accountKeyPairs,err := km.GetMappedKeyPairs(accountNames)
	if err != nil {
		log.Println(err)
		return nil,err
	}
	err = util.Write("genesis.json", eosconf.GenerateGenesis(keyPairs[masterIP].PublicKey))
	if err != nil{
		log.Println(err)
		return nil,err
	}
	err = util.Write("config.ini", eosconf.GenerateConfig())
	if err != nil{
		log.Println(err)
		return nil,err
	}
	/**Start keos and add all the key pairs for all the nodes**/
	{
		for i, server := range servers {
			for localId,ip := range server.Ips {
				/**Start keosd**/
				_,err = clients[i].DockerExecd(localId,"keosd --http-server-address 0.0.0.0:8900")
				if err != nil{
					log.Println(err)
					return nil,err
				}
				clientPasswords[ip],err = eos_createWallet(clients[i], localId)
				if err != nil {
					log.Println(err)
					return nil,err
				}
				sem.Acquire(ctx,1)

				go func(accountKeyPairs map[string]util.KeyPair,accountNames []string,localId int,server int){
					defer sem.Release(1)
					cmds := []string{}
					for _,name := range accountNames {
						if len(cmds) > 50 {
							_,err := clients[server].DockerMultiExec(localId,cmds)
							if err != nil {
								log.Println(err)
								state.ReportError(err)
								return
							}
							cmds = []string{}
						}
						cmds = append(cmds,fmt.Sprintf("cleos wallet import --private-key %s", accountKeyPairs[name].PrivateKey))
					}
					if len(cmds) > 0 {
						_,err := clients[server].DockerMultiExec(localId,cmds)
						if err != nil {
							log.Println(err)
							state.ReportError(err)
							return
						}
					}
					
				}(accountKeyPairs,accountNames,localId,i)

			}
		}
		sem.Acquire(ctx,conf.ThreadLimit)
		sem.Release(conf.ThreadLimit)
		if !state.ErrorFree(){
			return nil, state.GetError()
		}
	}
	password := clientPasswords[servers[0].Ips[0]]
	passwordNormal := clientPasswords[servers[0].Ips[1]]

	{
		
		node := 0
		for i, server := range servers {
			sem.Acquire(ctx,1)
			go func(i int, ips []string){
				defer sem.Release(1)
				err := clients[i].Scp("./genesis.json", "/home/appo/genesis.json")
				if err != nil {
					log.Println(err)
					state.ReportError(err)
					return
				}
				err = clients[i].Scp("./config.ini", "/home/appo/config.ini")
				if err != nil {
					log.Println(err)
					state.ReportError(err)
					return
				}

				for j := 0; j < len(ips); j++ {
					_,err = clients[i].DockerExec(j,"mkdir /datadir/")
					if err != nil {
						log.Println(err)
						state.ReportError(err)
						return
					}
					_,err = clients[i].FastMultiRun(
									fmt.Sprintf("docker cp /home/appo/genesis.json whiteblock-node%d:/datadir/", j),
									fmt.Sprintf("docker cp /home/appo/config.ini whiteblock-node%d:/datadir/", j))
					if err != nil {
						log.Println(err)
						state.ReportError(err)
						return
					}
					node++
				}
				_,err = clients[i].Run("rm /home/appo/genesis.json")
				if err != nil {
					log.Println(err)
					state.ReportError(err)
					return
				}
				_,err = clients[i].Run("rm /home/appo/config.ini")
				if err != nil {
					log.Println(err)
					state.ReportError(err)
					return
				}
			}(i,server.Ips)
		}
		sem.Acquire(ctx,conf.ThreadLimit)
		sem.Release(conf.ThreadLimit)
		if !state.ErrorFree(){
			return nil, state.GetError()
		}
	}
	defer func(){
		util.Rm("./genesis.json")
		util.Rm("./config.ini")
	}()

	/**Step 2d**/
	{
		
		res,err := clients[0].DockerExec(0,fmt.Sprintf("cleos wallet import --private-key %s", 
				keyPairs[masterIP].PrivateKey))

		if err != nil {
			log.Println(err)
			return nil,err
		}
		println(res)
		
		res,err = clients[0].DockerExecd(0,
					fmt.Sprintf(`nodeos -e -p eosio --genesis-json /datadir/genesis.json --config-dir /datadir --data-dir /datadir %s %s`,
						eos_getKeyPairFlag(keyPairs[masterIP]),
						eos_getPTPFlags(servers, 0)))
		if err != nil {
			log.Println(err)
			return nil,err
		}
		println(res)
	}
	

	/**Step 3**/
	{
		clients[0].Run(fmt.Sprintf("docker exec whiteblock-node0 cleos -u http://%s:8889 wallet unlock --password %s",
			masterIP, password))//Can fail

		for _, account := range contractAccounts {
			sem.Acquire(ctx,1)
			go func(masterIP string,account string,masterKeyPair util.KeyPair,contractKeyPair util.KeyPair){
				defer sem.Release(1)
				
				
				_,err = clients[0].DockerExec(0,fmt.Sprintf("cleos wallet import --private-key %s", 
							contractKeyPair.PrivateKey))
				if err != nil {
					state.ReportError(err)
					log.Println(err)
					return
				}
				_,err = clients[0].DockerExec(0,fmt.Sprintf("cleos -u http://%s:8889 create account eosio %s %s %s",
							masterIP, account,masterKeyPair.PublicKey,contractKeyPair.PublicKey))
				if err != nil {
					state.ReportError(err)
					log.Println(err)
					return
				}

				//log.Println("Finished creating account for "+account)
			}(masterIP,account,masterKeyPair,contractKeyPairs[account])

		}
		sem.Acquire(ctx,conf.ThreadLimit)
		sem.Release(conf.ThreadLimit)

		if !state.ErrorFree(){
			return nil, state.GetError()
		}
		
	}
	/**Steps 4 and 5**/
	{
		contracts := []string{"eosio.token","eosio.msig"}
		util.SshExecIgnore(masterServerIP, fmt.Sprintf("docker exec whiteblock-node0 cleos -u http://%s:8889 wallet unlock --password %s",
				masterIP, password))
		for _, contract := range contracts {
			
			_,err = util.DockerExec(masterServerIP,0, fmt.Sprintf("cleos -u http://%s:8889 set contract %s /opt/eosio/contracts/%s",
				masterIP, contract, contract))
			if err != nil {
				log.Println(err)
				return nil,err
			}
		}
	}
	/**Step 6**/

	res,err := clients[0].DockerExec(0, fmt.Sprintf("cleos -u http://%s:8889 push action eosio.token create '[ \"eosio\", \"10000000000.0000 SYS\" ]' -p eosio.token@active",
		masterIP))
	if err != nil {
		log.Println(err)
		return nil,err
	}
	println(res)

	res,err = clients[0].DockerExec(0, fmt.Sprintf("cleos -u http://%s:8889 push action eosio.token issue '[ \"eosio\", \"1000000000.0000 SYS\", \"memo\" ]' -p eosio@active",
		masterIP))
	
	if err != nil{
		log.Println(err)
		return nil,err
	}

	println(res)

	clients[0].Run(fmt.Sprintf("docker exec whiteblock-node0 cleos -u http://%s:8889 wallet unlock --password %s",
				masterIP, password))//Ignore fail


	/**Step 7**/
	for i := 0 ; i < 5; i++{
		res, err := clients[0].DockerExec(0, fmt.Sprintf("cleos -u http://%s:8889 set contract -x 1000 eosio /opt/eosio/contracts/eosio.system",
		masterIP))
		if(err == nil){
			fmt.Println("SUCCESS!!!!!")
			fmt.Println(res)
			break
		}
		fmt.Println(res)
	}
	
	
	/**Step 8**/

	
		res,err = clients[0].DockerExec(0,
			fmt.Sprintf(`cleos -u http://%s:8889 push action eosio setpriv '["eosio.msig", 1]' -p eosio@active`,
				masterIP))
		if err != nil{
			log.Println(err)
			return nil,err
		}

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
					
					_,err = clients[0].DockerExec(0,fmt.Sprintf("cleos wallet import --private-key %s",keyPair.PrivateKey))
					if err != nil {
						log.Println(err)
						state.ReportError(err)
						return
					}

					if node < eosconf.BlockProducers {
						_,err = clients[0].DockerExec(0,
									fmt.Sprintf(`cleos -u http://%s:8889 system newaccount eosio --transfer %s %s %s --stake-net "1000000.0000 SYS" --stake-cpu "1000000.0000 SYS" --buy-ram "1000000 SYS"`,
										masterIP,
										eos_getProducerName(node),
										masterKeyPair.PublicKey,
										keyPair.PublicKey))
						if err != nil {
							log.Println(err)
							state.ReportError(err)
							return
						}
						
						_,err = clients[0].DockerExec(0,fmt.Sprintf(`cleos -u http://%s:8889 transfer eosio %s "100000.0000 SYS"`,
										masterIP,
										eos_getProducerName(node)))
						if err != nil {
							log.Println(err)
							state.ReportError(err)
							return
						}
					}
					
				}(masterServerIP,masterKeyPair,keyPairs[ip],node)
				node++
			}
		}
		sem.Acquire(ctx,conf.ThreadLimit)
		sem.Release(conf.ThreadLimit)
		if !state.ErrorFree() {
			return nil,state.GetError()
		}
	}
	
	/**Step 11c**/
	{
		node := 0
		for i, server := range servers {
			for j, ip := range server.Ips {
				
				if node == 0 {
					node++
					continue
				}
				sem.Acquire(ctx,1)

				go func(server int,servers []db.Server,node int,j int,kp util.KeyPair){
					defer sem.Release(1)
					clients[server].DockerExec(j,"mkdir -p /datadir/blocks")
					p2pFlags := eos_getPTPFlags(servers,node)
					if node > eosconf.BlockProducers {
		
						res,err := clients[server].DockerExecd(j,
									fmt.Sprintf(`nodeos --genesis-json /datadir/genesis.json --config-dir /datadir --data-dir /datadir %s %s`,
										eos_getKeyPairFlag(kp),
										p2pFlags))
						if err != nil{
							log.Println(err)
							state.ReportError(err)
							return
						}
						println(res)
					}else{

						res,err := clients[server].DockerExecd(j,
									fmt.Sprintf(`nodeos --genesis-json /datadir/genesis.json --config-dir /datadir --data-dir /datadir -p %s %s %s`,
										eos_getProducerName(node),
										eos_getKeyPairFlag(kp),
										p2pFlags))
						if err != nil{
							log.Println(err)
							state.ReportError(err)
							return
						}
						println(res)
					}
					
				}(i,servers,node,j,keyPairs[ip])
				node++
			}
		}
		sem.Acquire(ctx,conf.ThreadLimit)
		sem.Release(conf.ThreadLimit)
		if !state.ErrorFree(){
			return nil,state.GetError()
		}
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
					clients[0].DockerExec(0,fmt.Sprintf("cleos -u http://%s:8889 wallet unlock --password %s",
						masterIP, password))//ignore
				}

				
				res,err = clients[0].DockerExec(0,
							fmt.Sprintf("cleos --wallet-url http://%s:8900 -u http://%s:8889 system regproducer %s %s https://whiteblock.io/%s",
								masterIP,
								masterIP,
								eos_getProducerName(node),
								keyPairs[ip].PublicKey,
								keyPairs[ip].PublicKey))
				if err != nil{
					log.Println(err)
					return nil,err
				}
				println(res)
				
				node++
			}
		}
	}
	/**Step 11b**/
	res,err = clients[0].DockerExec(0,fmt.Sprintf("cleos -u http://%s:8889 system listproducers",
									masterIP))
	if err != nil{
		log.Println(err)
		return nil,err
	}
	println(res)
	/**Create normal user accounts**/
	{
		for _, name := range accountNames {
			sem.Acquire(ctx,1)
			go func(masterServerIP string,name string,masterKeyPair util.KeyPair,accountKeyPair util.KeyPair){
				defer sem.Release(1)
				
					res,err := clients[0].DockerExec(0,
								fmt.Sprintf(`cleos -u http://%s:8889 system newaccount eosio --transfer %s %s %s --stake-net "500000.0000 SYS" --stake-cpu "2000000.0000 SYS" --buy-ram "2000000 SYS"`,
											masterIP,
											name,
											masterKeyPair.PublicKey,
											accountKeyPair.PublicKey))
					if err != nil{
						log.Println(err)
						state.ReportError(err)
						return
					}
					println(res)

				
					res,err = clients[0].DockerExec(0,
							fmt.Sprintf(`cleos -u http://%s:8889 transfer eosio %s "100000.0000 SYS"`,
										masterIP,
										name))
					if err != nil{
						log.Println(err)
						state.ReportError(err)
						return
					}
					println(res)

			}(masterServerIP,name,masterKeyPair,accountKeyPairs[name])
		}
		sem.Acquire(ctx,conf.ThreadLimit)
		sem.Release(conf.ThreadLimit)
		if !state.ErrorFree() {
			return nil, state.GetError()
		}
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
		_,err = clients[0].DockerExec(1, fmt.Sprintf("cleos -u http://%s:8889 wallet unlock --password %s",
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
				
					res,err := clients[0].DockerExec(1,
							fmt.Sprintf("cleos -u http://%s:8889 system voteproducer prods %s %s",
										masterIP,
										name,
										eos_getProducerName(prod)))
						if err != nil{
							log.Println(err)
							state.ReportError(err)
							return
						}
					println(res)
			}(masterServerIP,masterIP,name,prod)
			n++;
		}
		sem.Acquire(ctx,conf.ThreadLimit)
		sem.Release(conf.ThreadLimit)
		if !state.ErrorFree() {
			return nil, state.GetError()
		}
	}
	
	/**Step 12**/
	
	_,err = clients[0].DockerExec(0,
			fmt.Sprintf(
				`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio.prods", "permission": "active"}}]}}' -p eosio@owner`,
				masterIP))
	if err != nil{
		log.Println(err)
		return nil,err
	}

	
	_,err = clients[0].DockerExec(0,
			fmt.Sprintf(
				`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio.prods", "permission": "active"}}]}}' -p eosio@active`,
				masterIP))
	if err != nil {
		log.Println(err)
		return nil,err
	}
	
	
	_,err = clients[0].DockerExec(0,
		fmt.Sprintf(
			`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.bpay", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.bpay@owner`,
			masterIP))

	if err != nil {
		log.Println(err)
		return nil,err
	}
	
	_,err = clients[0].DockerExec(0,
		fmt.Sprintf(
			`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.bpay", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.bpay@active`,
			masterIP))

	if err != nil{
		log.Println(err)
		return nil,err
	}

	_,err = clients[0].DockerExec(0,
			fmt.Sprintf(
				`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.msig", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.msig@owner`,
				masterIP))

	if err != nil {
		log.Println(err)
		return nil,err
	}

	_,err = clients[0].DockerExec(0,
			fmt.Sprintf(
				`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.msig", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.msig@active`,
				masterIP))

	if err != nil {
		return nil,err
	}

	
	_,err = clients[0].DockerExec(0,
			fmt.Sprintf(
				`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.names", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.names@owner`,
				masterIP))

	if err != nil {
		return nil,err
	}

	_,err = clients[0].DockerExec(0,fmt.Sprintf(
				`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.names", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.names@active`,
				masterIP))

	if err != nil {
		log.Println(err)
		return nil,err
	}

	
	_,err = clients[0].DockerExec(0,fmt.Sprintf(
				`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.ram", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.ram@owner`,
				masterIP))

	if err != nil {
		log.Println(err)
		return nil,err
	}

	_,err = clients[0].DockerExec(0,fmt.Sprintf(
				`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.ram", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.ram@active`,
				masterIP))

	if err != nil {
		log.Println(err)
		return nil,err
	}
	_,err = clients[0].DockerExec(0,fmt.Sprintf(
				`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.ramfee", "permission": "owner", "parent": "", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.ramfee@owner`,
				masterIP))
	if err != nil {
		log.Println(err)
		return nil,err
	}

	
	_,err = clients[0].DockerExec(0,fmt.Sprintf(
				`cleos -u http://%s:8889 push action eosio updateauth '{"account": "eosio.ramfee", "permission": "active", "parent": "owner", "auth": {"threshold": 1, "keys": [], "waits": [], "accounts": [{"weight": 1, "permission": {"actor": "eosio", "permission": "active"}}]}}' -p eosio.ramfee@active`,
				masterIP))
	if err != nil {
		log.Println(err)
		return nil,err
	}

	out := []string{}

	for _, server := range servers {
		for _, ip := range server.Ips {
			out = append(out,clientPasswords[ip])
		}
	}

	return out,nil
}

/*func eos_getKeyPair(serverIP string) (util.KeyPair,error){
	data,err := util.SshExec(serverIP, "docker exec whiteblock-node0 cleos create key --to-console | awk '{print $3}'")
	if err != nil {
		return util.KeyPair{},err
	}
	//fmt.Printf("RAW KEY DATA%s\n", data)
	keyPair := strings.Split(data, "\n")
	if(len(data) < 10){
		return util.KeyPair{},errors.New(fmt.Sprintf("Unexpected create key output %s\n",keyPair))
		panic(1)
	}
	return util.KeyPair{PrivateKey: keyPair[0], PublicKey: keyPair[1]},nil
}*/

/**
func eos_getKeyPairs(servers []db.Server,clients []*util.SshClient) (map[string]util.KeyPair,error) {
	keyPairs := make(map[string]util.KeyPair)
	//Get the key pairs for each nodeos account
	
	var wg sync.WaitGroup
	var mutex = &sync.Mutex{}

	for i, server := range servers {
		wg.Add(1)
		go func(server int,ips []string){
			defer wg.Done()
			for _, ip := range ips {
				data,err := clients[server].DockerExec(0,"cleos create key --to-console | awk '{print $3}'")
				if err != nil {
					state.ReportError(err)
					return
				}
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
		}(i,server.Ips)
	}
	wg.Wait()
	if !state.ErrorFree() {
		return nil, state.GetError()
	}
	return keyPairs,nil
}


func eos_getContractKeyPairs(servers []db.Server,contractAccounts []string) (map[string]util.KeyPair,error) {

	keyPairs := make(map[string]util.KeyPair)
	server := servers[0]
	var err error

	for _,contractAccount := range contractAccounts {
		
		keyPairs[contractAccount],err = eos_getKeyPair(server.Addr)
		if err != nil {
			return keyPairs,err
		}
	}
	return keyPairs,nil
}

func eos_getUserAccountKeyPairs(client *util.SshClient,accountNames []string) (map[string]util.KeyPair,error) {

	keyPairs := make(map[string]util.KeyPair)
	sem := semaphore.NewWeighted(conf.ThreadLimit)
	var mutex = &sync.Mutex{}
	ctx := context.TODO()

	for _,name := range accountNames {
		sem.Acquire(ctx,1)
		go func(name string){
			defer sem.Release(1)
			data,err := client.DockerExec(0, "cleos create key --to-console | awk '{print $3}'")
			if err != nil{
				state.ReportError(err)
				return
			}
			//fmt.Printf("RAW KEY DATA%s\n", data)
			keyPair := strings.Split(data, "\n")
			if(len(data) < 10){
				fmt.Printf("Unexpected create key output %s\n",keyPair)
				panic(1)
			}
			mutex.Lock()
			keyPairs[name] = util.KeyPair{PrivateKey: keyPair[0], PublicKey: keyPair[1]}
			mutex.Unlock()
		}(name)
	}
	sem.Acquire(ctx,conf.ThreadLimit)
	sem.Release(conf.ThreadLimit)
	if !state.ErrorFree(){
		return keyPairs,state.GetError()
	}
	return keyPairs,nil
}
*/

func eos_createWallet(client *util.SshClient, node int) (string,error) {
	data,err := client.DockerExec(node,"cleos wallet create --to-console | tail -n 1")
	if err != nil{
		return "",err
	}
	//fmt.Printf("CREATE WALLET DATA %s\n",data)
	offset := 0
	for data[len(data) - (offset + 1)] != '"' {
		offset++
	}
	offset++
	data = data[1 : len(data) - offset]
	fmt.Printf("CREATE WALLET DATA %s\n",data)
	return data,nil
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
