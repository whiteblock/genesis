package main

import (
	"encoding/json"
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"regexp"
	"github.com/satori/go.uuid"
)

/**CONSTANTS**/
const ETH_MAX_PEERS	int 			= 	1000
const ETH_INIT_WALLET_VALUE	string	=	"100000000000000000000"

var eth_sem = semaphore.NewWeighted(THREAD_LIMIT)


/**
 * Build the Ethereum Test Network
 * @param  uint64	gas			The gas limit
 * @param  uint64	chainId		The chain id 
 * @param  uint64	networkId	The test net network id
 * @param  int		nodes		The number of nodes in the network
 * @param  []Server	servers		The list of servers passed from build
 */
func ethereum(gas uint64,chainId uint64,networkId uint64,nodes int,servers []Server){
	ctx := context.TODO()
	rm("tmp/node*","tmp/all_wallet","tmp/static-nodes.json","tmp/keystore","tmp/CustomGenesis.json")
	for i := 1; i <= nodes; i++ {
		mkdir(fmt.Sprintf("./tmp/node%d",i))
		//fmt.Printf("---------------------  CREATING pre-allocated accounts for NODE-%d  ---------------------\n",i)

	}
	eth_createPassFiles(nodes)
	wallets := []string{}

	for i := 1; i <= nodes; i++{
		eth_sem.Acquire(ctx,1)
		wallets = append(wallets,eth_getWallet(i));
	}
	unlock := ""

	for i,wallet := range wallets {
		if i != 0 {
			unlock += ","
		}
		unlock += wallet
	}

	eth_createGenesisfile(gas,chainId,wallets)

	eth_initNodeDirectories(nodes,networkId,servers)
	mkdir("tmp/keystore")
	eth_distributeUTCKeystore(nodes)
	for i := 1; i <= nodes; i++ {
		cp("tmp/static_nodes.json",fmt.Sprintf("tmp/node%d/",i))
	}
	for i := 0; i < nodes; i++ {
		eth_sem.Acquire(ctx,1)
		fmt.Printf("-----------------------------  Starting NODE-%d  -----------------------------\n",i + 1)
		serverip, nodeIP , num := getInfo(servers,i)
		go eth_startNode(networkId,i+1,serverip,num,unlock,nodeIP)
	}
	err := eth_sem.Acquire(ctx,THREAD_LIMIT)
	check_fatal(err)

	eth_sem.Release(THREAD_LIMIT)

	eth_setupEthNetStats(servers[0].addr)
	node := 0
	for _,server := range servers {
		for j,ip := range server.ips{
			eth_sem.Acquire(ctx,1)
			go eth_setupEthNetIntel(server.addr,ip,servers[0].iaddr.ip,node,j)
			node++
		}
	}
	eth_setupBlockExplorer(servers[0].ips[0],servers[0].addr)

	err = eth_sem.Acquire(ctx,THREAD_LIMIT)
	check_fatal(err)

	eth_sem.Release(THREAD_LIMIT)
	fmt.Printf("Cleaning up...")
	rm("tmp/node*","tmp/all_wallet","tmp/static-nodes.json","tmp/keystore","tmp/CustomGenesis.json")
	fmt.Printf("done\n")
	//fmt.Printf("To view Eth Net Stat type:\t\t\ttmux attach-session -t netstats\n")
	
}
/***************************************************************************************************************************/

/**
 * Create the password files
 * @param  int	nodes	The number of nodes
 */
func eth_createPassFiles(nodes int){
	var data string
	for i := 1; i <= nodes; i++{
		data += "second\n"
	}

	for i := 1; i <= nodes; i++{
		write(fmt.Sprintf("tmp/node%d/passwd.file",i),data)
	}
}

/**
 * Creates a wallet for a node
 * @param  uint64  node        The node number
 */
func eth_getWallet(node int) string{
	gethResults := bashExec(
		fmt.Sprintf("geth --datadir tmp/node%d/ --password tmp/node%d/passwd.file account new",
			node,node))
	//fmt.Printf("RAW:%s\n",gethResults)
	addressPattern := regexp.MustCompile(`\{[A-z|0-9]+\}`)
	addresses := addressPattern.FindAllString(gethResults,-1)
	if len(addresses) < 1 {
		return ""
	}
	address := addresses[0]
	address = address[1:len(address)-1]
 	eth_sem.Release(1)
	return address
}

/**
 * Create the custom genesis file for Ethereum
 * @param  uint64	gas			The target gas limit
 * @param  uint64	chainId		The chain id
 * @param  []string wallets		The wallets to be allocated a balance
 */
func eth_createGenesisfile(gas uint64,chainId uint64,wallets []string){
	alloc := "\n"
	for i,wallet := range wallets {
		alloc += fmt.Sprintf("\t\t\"%s\":{\"balance\":\"%s\"}",wallet,ETH_INIT_WALLET_VALUE)
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

	write("tmp/CustomGenesis.json",genesis)
}

/**
 * Creates the datadir for a node and returns the enode address
 * @param  int		node		The nodes number
 * @param  uint64	networkId	The test net network id
 * @param  string	ip			The node's IP address
 * @return string				The node's enode address
 */
func eth_initNode(node int, networkId uint64,ip string) string {
	fmt.Printf("---------------------  CREATING block directory for NODE-%d ---------------------\n",node)
	gethResults := bashExec(fmt.Sprintf("echo -e \"admin.nodeInfo.enode\\nexit\\n\" |  geth --rpc --datadir tmp/node%d/ --networkid %d console",node,networkId))
	//fmt.Printf("RAWWWWWWWWWWWW%s\n\n\n",gethResults)
	enodePattern := regexp.MustCompile(`enode:\/\/[A-z|0-9]+@\[\:\:\]\:[0-9]+`)
	enode := enodePattern.FindAllString(gethResults,1)[0]
	enodeAddressPattern := regexp.MustCompile(`\[\:\:\]`)
	enode = enodeAddressPattern.ReplaceAllString(enode,ip);

	write(fmt.Sprintf("./tmp/node%d/enode",node),fmt.Sprintf("%s\n",enode))
	return enode
}

/**
 * Initialize the chain from the custom genesis file
 * @param  int		nodes		The number of nodes
 * @param  uint64	networkId	The test net network id
 * @param  []Server	servers		The list of servers
 */
func eth_initNodeDirectories(nodes int,networkId uint64,servers []Server){
	static_nodes := []string{};
	for i := 1; i <= nodes; i++ {
		//fmt.Printf("---------------------  CREATING block directory for NODE-%d ---------------------\n",i)
		//Load the CustomGenesis file
		bashExec(
			fmt.Sprintf("geth --datadir tmp/node%d --networkid %d init tmp/CustomGenesis.json",i,networkId))
		

		_,ip,_ := getInfo(servers,i - 1)
		static_nodes = append(static_nodes,eth_initNode(i,networkId,ip))
	}
	out, err := json.Marshal(static_nodes)
	//fmt.Printf("-----Static Nodes.json------\n%+v\n\n",static_nodes)
	check_fatal(err)
	for i := 1; i <= nodes; i++ {
		write(fmt.Sprintf("tmp/node%d/static-nodes.json",i),string(out))
	}
	
		
}

/**
 * Distribute the UTC keystore files amongst the nodes
 * @param  int	nodes	The number of nodes
 */
func eth_distributeUTCKeystore(nodes int){
	//Copy all UTC keystore files to every Node directory
	for i := 1; i <= nodes; i++ {
		cpr(fmt.Sprintf("tmp/node%d/keystore/",i),"tmp/")
	}
	for i := 1; i <= nodes; i++ {
		cpr("tmp/keystore/",fmt.Sprintf("tmp/node%d/",i))
	}
}

/**
 * Starts geth on a node
 * @param  uint64	networkId	The test net network id
 * @param  int		node		The absolute node number
 * @param  string	server		The IP address of the server
 * @param  int		num			The relative node number
 * @param  string	unlock		The unlock argument string
 * @param  string	nodeIP		The IP address of the node
 */
func eth_startNode(networkId uint64,node int,server string,num int,unlock string,nodeIP string){
	
	name := fmt.Sprintf("whiteblock-node%d",num)
	sshExec(server,fmt.Sprintf("rm -rf tmp/node%d",node))
	scpr(server,fmt.Sprintf("tmp/node%d",node))

	gethCmd := fmt.Sprintf(`geth --datadir /whiteblock/node%d --nodiscover --maxpeers %d --networkid %d --rpc --rpcaddr %s --rpcapi "web3,db,eth,net,personal" --rpccorsdomain "0.0.0.0" --mine --unlock="%s" --password /whiteblock/node%d/passwd.file console`,
			node,
			ETH_MAX_PEERS,
			networkId,
			nodeIP,
			unlock,
			node)

	sshMultiExec(server,
		
		fmt.Sprintf("docker exec %s mkdir -p /whiteblock/node%d/",name,node),
		fmt.Sprintf("docker cp ~/tmp/node%d %s:/whiteblock",node,name),
		fmt.Sprintf("docker exec -d %s tmux new -s whiteblock -d",name),
		fmt.Sprintf("docker exec -d %s tmux send-keys -t whiteblock '%s' C-m",name,gethCmd),
	)

	eth_sem.Release(1)
}

/**
 * Setups Eth Net Intelligence API on a node
 * @param  string	serverIP				The IP address of the host server
 * @param  string	nodeIP					The IP address of the node
 * @param  string	ethnetIP				The IP of the node reachable interface for ETH Net Stats
 * @param  int		absNum					The absolute number of the node
 * @param  int		relNum					The relative number of the node on the host server
 */
func eth_setupEthNetIntel(serverIP string,nodeIP string,ethnetIP string,absNum int,relNum int){
	relName := fmt.Sprintf("whiteblock-node%d",relNum)
	absName := fmt.Sprintf("whiteblock-node%d",absNum)
	sedCmd := fmt.Sprintf(`docker exec %s sed -i -r 's/"INSTANCE_NAME"(\s)*:(\s)*"(\S)*"/"INSTANCE_NAME"\t: "%s"/g' /eth-net-intelligence-api/app.json`,relName,absName)
	sedCmd2 := fmt.Sprintf(`docker exec %s sed -i -r 's/"WS_SERVER"(\s)*:(\s)*"(\S)*"/"WS_SERVER"\t: "http:\/\/%s:3000"/g' /eth-net-intelligence-api/app.json`,relName,ethnetIP)
	sedCmd3 := fmt.Sprintf(`docker exec %s sed -i -r 's/"RPC_HOST"(\s)*:(\s)*"(\S)*"/"RPC_HOST"\t: "%s"/g' /eth-net-intelligence-api/app.json`,relName,nodeIP)

	//sedCmd3 := fmt.Sprintf("docker exec -it %s sed -i 's/\"WS_SECRET\"(\\s)*:(\\s)*\"[A-Z|a-z|0-9| ]*\"/\"WS_SECRET\"\\t: \"second\"/g' /eth-net-intelligence-api/app.json",container)
	sshFastMultiExec(serverIP,
		fmt.Sprintf("docker exec -d %s tmux new -s ethnet -d",relName),
		sedCmd,
		sedCmd2,
		sedCmd3,
		fmt.Sprintf("docker exec -d %s tmux send-keys -t ethnet 'cd /eth-net-intelligence-api && pm2 start app.json' C-m",relName),
		)
	
	//sshExec(server.addr,
		//fmt.Sprintf("%s&&%s&&%s&&%s",sedCmd,sedCmd2,sedCmd3,startEthNetStatsCmd))
	
	eth_sem.Release(1)
}

/**
 * Setup Eth Net Stats on a server
 * @param  string 	 ip 	The servers config
 */
func eth_setupEthNetStats(ip string){
	sshExecIgnore(ip,"rm -rf eth-netstats")
	sshExec(ip,"wget http://172.16.0.8/eth-netstats.tar.gz && tar xf eth-netstats.tar.gz && rm eth-netstats.tar.gz")

	sshExecIgnore(ip,"tmux kill-session -t netstats")
	sshExec(ip,"tmux new -s netstats -d")
	sshExec(ip,"tmux send-keys -t netstats 'cd /home/appo/eth-netstats && npm install && grunt && WS_SECRET=second npm start' C-m")

}

/**
 * Set up the block explorer
 * @param  string	nodeIP        The IP address of the node
 * @param  string	serverIP      The IP address of the host server
 */
func eth_setupBlockExplorer(nodeIP string,serverIP string){
	sshExecIgnore(serverIP,"rm -rf ~/explorer")
	sshExecIgnore(serverIP,"tmux kill-session -t blockExplorer")

	id, _ := uuid.NewV4()
	sshFastMultiExec(serverIP,
		"git clone https://github.com/ethereumproject/explorer.git",
		"cp ~/explorer/config.example.json ~/explorer/config.json",
		fmt.Sprintf(`sed -i -r 's/"nodeAddr"(\s)*:(\s)*"(\S)*"/"nodeAddr":\t"%s"/g' ~/explorer/config.json`,nodeIP),
		`sed -i -r 's/"symbol"(\s)*:(\s)*"(\S| )*"/"symbol":\t"ETH"/g' ~/explorer/config.json`,
		`sed -i -r 's/"name"(\s)*:(\s)*"(\S| )*"/"name":\t"Whiteblock"/g' ~/explorer/config.json`,
		`sed -i -r 's/"title"(\s)*:(\s)*"(\S| )*"/"title":\t"Whiteblock Block Explorer"/g' ~/explorer/config.json`,
		`sed -i -r 's/"author"(\s)*:(\s)*"(\S| )*"/"author":\t"WB"/g' ~/explorer/config.json`,
		"tmux new -s blockExplorer -d",
		
	)
	sshMultiExec(serverIP,
		
		"tmux send-keys -t blockExplorer 'cd ~/explorer && npm install' C-m",
		fmt.Sprintf("tmux send-keys -t blockExplorer 'PORT=8000 MONGO_URI=mongodb://localhost/blockDB%s npm start' C-m",id),)
	
}
