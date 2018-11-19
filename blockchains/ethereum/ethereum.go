package eth

import (
	"encoding/json"
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"regexp"
	"github.com/satori/go.uuid"
	//"sync"
	util "../../util"
	db "../../db"
	state "../../state"
)

/**CONSTANTS**/
const THREAD_LIMIT int64		=   10
const MAX_PEERS	int 			= 	1000
const INIT_WALLET_VALUE	string	=	"100000000000000000000"



/**
 * Build the Ethereum Test Network
 * @param  uint64	gas			The gas limit
 * @param  uint64	chainId		The chain id 
 * @param  uint64	networkId	The test net network id
 * @param  int		nodes		The number of nodes in the network
 * @param  []Server	servers		The list of servers passed from build
 */
func Ethereum(gas uint64,chainId uint64,networkId uint64,nodes int,servers []db.Server){
	//var mutex = &sync.Mutex{}
	var sem = semaphore.NewWeighted(THREAD_LIMIT)
	ctx := context.TODO()
	util.Rm("tmp/node*","tmp/all_wallet","tmp/static-nodes.json","tmp/keystore","tmp/CustomGenesis.json")
	state.SetBuildSteps(8)
	defer func(){
		fmt.Printf("Cleaning up...")
		util.Rm("tmp/node*","tmp/all_wallet","tmp/static-nodes.json","tmp/keystore","tmp/CustomGenesis.json")
		fmt.Printf("done\n")
	}()
	
	for i := 1; i <= nodes; i++ {
		util.Mkdir(fmt.Sprintf("./tmp/node%d",i))
		//fmt.Printf("---------------------  CREATING pre-allocated accounts for NODE-%d  ---------------------\n",i)

	}
	state.IncrementBuildProgress() 

	/**Create the Password files**/
	{
		var data string
		for i := 1; i <= nodes; i++{
			data += "second\n"
		}

		for i := 1; i <= nodes; i++{
			util.Write(fmt.Sprintf("tmp/node%d/passwd.file",i),data)
		}
	}
	state.IncrementBuildProgress()


	/**Create the wallets**/
	wallets := []string{}

	for i := 1; i <= nodes; i++{

		node := i
		//sem.Acquire(ctx,1)
		gethResults := util.BashExec(
			fmt.Sprintf("geth --datadir tmp/node%d/ --password tmp/node%d/passwd.file account new",
				node,node))
		//fmt.Printf("RAW:%s\n",gethResults)
		addressPattern := regexp.MustCompile(`\{[A-z|0-9]+\}`)
		addresses := addressPattern.FindAllString(gethResults,-1)
		if len(addresses) < 1 {
			return
		}
		address := addresses[0]
		address = address[1:len(address)-1]
	 	//sem.Release(1)
	 	//fmt.Printf("Created wallet with address: %s\n",address)
	 	//mutex.Lock()
	 	wallets = append(wallets,address)
	 	//mutex.Unlock()
		
		
	}
	state.IncrementBuildProgress()
	unlock := ""

	for i,wallet := range wallets {
		if i != 0 {
			unlock += ","
		}
		unlock += wallet
	}
	fmt.Printf("unlock = %s\n%+v\n\n",wallets,unlock)

	state.IncrementBuildProgress()

	createGenesisfile(gas,chainId,wallets)
	state.IncrementBuildProgress()
	initNodeDirectories(nodes,networkId,servers)
	state.IncrementBuildProgress()
	util.Mkdir("tmp/keystore")
	distributeUTCKeystore(nodes)

	state.IncrementBuildProgress()

	for i := 1; i <= nodes; i++ {
		util.Cp("tmp/static_nodes.json",fmt.Sprintf("tmp/node%d/",i))
	}
	node := 0
	for _, server := range servers {
		for j, ip := range server.Ips{
			sem.Acquire(ctx,1)
			fmt.Printf("-----------------------------  Starting NODE-%d  -----------------------------\n",node)

			go func(networkId uint64,node int,server string,num int,unlock string,nodeIP string){
				name := fmt.Sprintf("whiteblock-node%d",num)
				util.SshExec(server,fmt.Sprintf("rm -rf tmp/node%d",node))
				util.Scpr(server,fmt.Sprintf("tmp/node%d",node))

				gethCmd := fmt.Sprintf(`geth --datadir /whiteblock/node%d --nodiscover --maxpeers %d --networkid %d --rpc --rpcaddr %s --rpcapi "web3,db,eth,net,personal,miner" --rpccorsdomain "0.0.0.0" --mine --unlock="%s" --password /whiteblock/node%d/passwd.file console`,
						node,
						MAX_PEERS,
						networkId,
						nodeIP,
						unlock,
						node)

				util.SshMultiExec(server,
					fmt.Sprintf("docker exec %s mkdir -p /whiteblock/node%d/",name,node),
					fmt.Sprintf("docker cp ~/tmp/node%d %s:/whiteblock",node,name),
					fmt.Sprintf("docker exec -d %s tmux new -s whiteblock -d",name),
					fmt.Sprintf("docker exec -d %s tmux send-keys -t whiteblock '%s' C-m",name,gethCmd),
				)

				sem.Release(1)
			}(networkId,node+1,server.Addr,j,unlock,ip)
			node ++
		}
	}
	err := sem.Acquire(ctx,THREAD_LIMIT)
	util.CheckFatal(err)
	state.IncrementBuildProgress()
	sem.Release(THREAD_LIMIT)
	/*
	setupEthNetStats(servers[0].Addr)
	node = 0
	for _,server := range servers {
		for j,ip := range server.Ips{
			sem.Acquire(ctx,1)
			go func(serverIP string,nodeIP string,ethnetIP string,absNum int,relNum int){
				relName := fmt.Sprintf("whiteblock-node%d",relNum)
				absName := fmt.Sprintf("whiteblock-node%d",absNum)
				sedCmd := fmt.Sprintf(`docker exec %s sed -i -r 's/"INSTANCE_NAME"(\s)*:(\s)*"(\S)*"/"INSTANCE_NAME"\t: "%s"/g' /eth-net-intelligence-api/app.json`,relName,absName)
				sedCmd2 := fmt.Sprintf(`docker exec %s sed -i -r 's/"WS_SERVER"(\s)*:(\s)*"(\S)*"/"WS_SERVER"\t: "http:\/\/%s:3000"/g' /eth-net-intelligence-api/app.json`,relName,ethnetIP)
				sedCmd3 := fmt.Sprintf(`docker exec %s sed -i -r 's/"RPC_HOST"(\s)*:(\s)*"(\S)*"/"RPC_HOST"\t: "%s"/g' /eth-net-intelligence-api/app.json`,relName,nodeIP)

				//sedCmd3 := fmt.Sprintf("docker exec -it %s sed -i 's/\"WS_SECRET\"(\\s)*:(\\s)*\"[A-Z|a-z|0-9| ]*\"/\"WS_SECRET\"\\t: \"second\"/g' /eth-net-intelligence-api/app.json",container)
				util.SshMultiExec(serverIP,
					fmt.Sprintf("docker exec -d %s tmux new -s ethnet -d",relName),
					sedCmd,
					sedCmd2,
					sedCmd3,
					fmt.Sprintf("docker exec -d %s tmux send-keys -t ethnet 'cd /eth-net-intelligence-api && pm2 start app.json' C-m",relName),
				)
	
				//util.SshExec(server.addr,
				//fmt.Sprintf("%s&&%s&&%s&&%s",sedCmd,sedCmd2,sedCmd3,startEthNetStatsCmd))
	
				sem.Release(1)
			}(server.Addr,ip,servers[0].Iaddr.Ip,node,j)
			node++
		}
	}
	setupBlockExplorer(servers[0].Ips[0],servers[0].Addr)

	err = sem.Acquire(ctx,THREAD_LIMIT)
	util.CheckFatal(err)

	sem.Release(THREAD_LIMIT)
	*/
	//fmt.Printf("To view Eth Net Stat type:\t\t\ttmux attach-session -t netstats\n")
	
}
/***************************************************************************************************************************/




/**
 * Create the custom genesis file for Ethereum
 * @param  uint64	gas			The target gas limit
 * @param  uint64	chainId		The chain id
 * @param  []string wallets		The wallets to be allocated a balance
 */
func createGenesisfile(gas uint64,chainId uint64,wallets []string){
	alloc := "\n"
	for i,wallet := range wallets {
		alloc += fmt.Sprintf("\t\t\"%s\":{\"balance\":\"%s\"}",wallet,INIT_WALLET_VALUE)
		if len(wallets) - 1 != i {
			alloc += ","
		}
		alloc += "\n"
	}

	genesis := fmt.Sprintf(
`{
	"config": {
		"chainId": %d,
		"homesteadBlock": 0,
		"eip155Block": 0,
		"eip158Block": 0
	},
	"difficulty": "0x0100000",
	"gasLimit": "0x0%X",
	"alloc": {%s 	}
}`,chainId,gas,alloc)

	util.Write("tmp/CustomGenesis.json",genesis)
}

/**
 * Creates the datadir for a node and returns the enode address
 * @param  int		node		The nodes number
 * @param  uint64	networkId	The test net network id
 * @param  string	ip			The node's IP address
 * @return string				The node's enode address
 */
func initNode(node int, networkId uint64,ip string) string {
	fmt.Printf("---------------------  CREATING block directory for NODE-%d ---------------------\n",node)
	gethResults := util.BashExec(fmt.Sprintf("echo -e \"admin.nodeInfo.enode\\nexit\\n\" |  geth --rpc --datadir tmp/node%d/ --networkid %d console",node,networkId))
	//fmt.Printf("RAWWWWWWWWWWWW%s\n\n\n",gethResults)
	enodePattern := regexp.MustCompile(`enode:\/\/[A-z|0-9]+@(\[\:\:\]|([0-9]|\.)+)\:[0-9]+`)
	enode := enodePattern.FindAllString(gethResults,1)[0]
	fmt.Printf("ENODE fetched is: %s\n",enode)
	enodeAddressPattern := regexp.MustCompile(`\[\:\:\]|([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})`)
	enode = enodeAddressPattern.ReplaceAllString(enode,ip);

	util.Write(fmt.Sprintf("./tmp/node%d/enode",node),fmt.Sprintf("%s\n",enode))
	return enode
}

/**
 * Initialize the chain from the custom genesis file
 * @param  int		nodes		The number of nodes
 * @param  uint64	networkId	The test net network id
 * @param  []Server	servers		The list of servers
 */
func initNodeDirectories(nodes int,networkId uint64,servers []db.Server){
	static_nodes := []string{};
	node := 1
	for _,server := range servers{
		for _,ip := range server.Ips{
			//fmt.Printf("---------------------  CREATING block directory for NODE-%d ---------------------\n",i)
			//Load the CustomGenesis file
			util.BashExec(
				fmt.Sprintf("geth --datadir tmp/node%d --networkid %d init tmp/CustomGenesis.json",node,networkId))
			

			static_nodes = append(static_nodes,initNode(node,networkId,ip))
			node++;
		}
	}
	out, err := json.Marshal(static_nodes)
	//fmt.Printf("-----Static Nodes.json------\n%+v\n\n",static_nodes)
	util.CheckFatal(err)
	for i := 1; i <= nodes; i++ {
		util.Write(fmt.Sprintf("tmp/node%d/static-nodes.json",i),string(out))
	}
	
		
}

/**
 * Distribute the UTC keystore files amongst the nodes
 * @param  int	nodes	The number of nodes
 */
func distributeUTCKeystore(nodes int){
	//Copy all UTC keystore files to every Node directory
	for i := 1; i <= nodes; i++ {
		util.Cpr(fmt.Sprintf("tmp/node%d/keystore/",i),"tmp/")
	}
	for i := 1; i <= nodes; i++ {
		util.Cpr("tmp/keystore/",fmt.Sprintf("tmp/node%d/",i))
	}
}

/**
 * Setup Eth Net Stats on a server
 * @param  string 	 ip 	The servers config
 */
func setupEthNetStats(ip string){
	util.SshExecIgnore(ip,"rm -rf eth-netstats")
	util.SshExec(ip,"wget http://172.16.0.8/eth-netstats.tar.gz && tar xf eth-netstats.tar.gz && rm eth-netstats.tar.gz")

	util.SshExecIgnore(ip,"tmux kill-session -t netstats")
	util.SshExec(ip,"tmux new -s netstats -d")
	util.SshExec(ip,"tmux send-keys -t netstats 'cd /home/appo/eth-netstats && npm install && grunt && WS_SECRET=second npm start' C-m")

}

/**
 * Set up the block explorer
 * @param  string	nodeIP        The IP address of the node
 * @param  string	serverIP      The IP address of the host server
 */
func setupBlockExplorer(nodeIP string,serverIP string){
	util.SshExecIgnore(serverIP,"tmux kill-session -t blockExplorer")
	util.SshExecIgnore(serverIP,"rm -rf ~/explorer")
	

	id, _ := uuid.NewV4()
	util.SshMultiExec(serverIP,
		"git clone https://github.com/ethereumproject/explorer.git",
		"cp ~/explorer/config.example.json ~/explorer/config.json",
		fmt.Sprintf(`sed -i -r 's/"nodeAddr"(\s)*:(\s)*"(\S)*"/"nodeAddr":\t"%s"/g' ~/explorer/config.json`,nodeIP),
		`sed -i -r 's/"symbol"(\s)*:(\s)*"(\S| )*"/"symbol":\t"ETH"/g' ~/explorer/config.json`,
		`sed -i -r 's/"name"(\s)*:(\s)*"(\S| )*"/"name":\t"Whiteblock"/g' ~/explorer/config.json`,
		`sed -i -r 's/"title"(\s)*:(\s)*"(\S| )*"/"title":\t"Whiteblock Block Explorer"/g' ~/explorer/config.json`,
		`sed -i -r 's/"author"(\s)*:(\s)*"(\S| )*"/"author":\t"WB"/g' ~/explorer/config.json`,
		"tmux new -s blockExplorer -d",
		
	)
	util.SshMultiExec(serverIP,
		
		"tmux send-keys -t blockExplorer 'cd ~/explorer && npm install' C-m",
		fmt.Sprintf("tmux send-keys -t blockExplorer 'PORT=8000 MONGO_URI=mongodb://localhost/blockDB%s npm start' C-m",id),)
	
}
