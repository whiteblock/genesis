package main

import (
	//"context"
	"fmt"
	//"golang.org/x/sync/semaphore"
	"strings"
)

var (
	//sem4 = semaphore.NewWeighted(THREAD_LIMIT)
	GENESIS_KEY = "5KQwrPbwdL6PhXujxW37FSSQZ1JiwsST4cqQzDeyXtP79zkvFD3"
)

/**
 * Setup the EOS test net
 * @param  int		nodes		The number of producers to make
 * @param  []Server servers		The list of relevant servers
 */
func eos(nodes int,servers []Server){
	//ctx := context.TODO()
	fmt.Println("-------------Setting Up EOS-------------")
	fmt.Println("\n*** Create Config File ***")
	eos_createConf(servers)

	//copy over the config files
	fmt.Println("\n*** Copy Over Config File ***")
	eos_initNodes(servers)
	//start keos on all of the nodes
	fmt.Println("\n*** Start Keos ***")
	eos_startKeos(servers)
	//get all of the wallets for the nodes 
	wallets := eos_getWallets(servers)
	//get all of the keypairs for the nodes
	fmt.Println("\n*** Get Key Pairs ***")
	keyPairs := eos_getKeyPairs(servers)
	//get PTP flags 
	fmt.Println("\n*** Create Peering Flags ***")
	ptpFlags := eos_getPTPFlags(servers)

	node := 0

	var masterNodeIP string

	fmt.Println("\n******* Starting Up Nodes *******")
	for _,server := range servers {
		for _, ip := range server.ips {
			fmt.Printf("\n*** Starting Node %d ***\n",node)

			if node == 0 {
				eos_setupProducer(server.addr,ip,keyPairs[ip],node,ptpFlags)
				eos_finishFirstProducer(server.addr,ip,keyPairs[ip])
				masterNodeIP = ip
			}else{
				eos_createAccount(server.addr,ip,keyPairs[ip],node,masterNodeIP)
				eos_setupProducer(server.addr,ip,keyPairs[ip],node,ptpFlags)
				eos_startScheduling(server.addr,keyPairs[ip],node,masterNodeIP,keyPairs[masterNodeIP].publicKey)
			}
			node++;
		}
	}
	fmt.Println("Finished setting up eos")
	write("treys_large_zedong.txt",fmt.Sprintf("%+v\n\n\n\n\n%+v\n\n\n\n",keyPairs,wallets))
}

/**
 * Create the Config file for the EOS Nodes
 * @param  []Server servers The list of servers
 */
func eos_createConf(servers []Server){
	//signature-provider = EOS6MRyAjQq8ud7hVNYcfnVPJqcVpscN5So8BhtHuGYqET5GDW5CV=KEY:5KQwrPbwdL6PhXujxW37FSSQZ1JiwsST4cqQzDeyXtP79zkvFD3
	constantEntries := []string {
		"bnet-endpoint = 0.0.0.0:4321",			"bnet-no-trx = false",						"blocks-dir = \"blocks\"",
		"chain-state-db-size-mb = 8192",		"reversible-blocks-db-size-mb = 340",		"contracts-console = false",
		"https-client-validate-peers = 1",		"access-control-allow-credentials = false",	"p2p-max-nodes-per-host = 1",
		"allowed-connection = any",				"max-clients = 0",							"connection-cleanup-period = 30",
		"network-version-match = 0",			"sync-fetch-span = 100",					"max-implicit-request = 1500",
		"enable-stale-production = false",		"pause-on-startup = false",					"max-transaction-time = 30",
		"max-irreversible-block-age = -1",		"producer-name = eosio",					"keosd-provider-timeout = 5",
		"txn-reference-block-lag = 0",			"wallet-dir = \".\"",						"unlock-timeout = 900",
		"plugin = eosio::chain_api_plugin",		"plugin = eosio::history_api_plugin", 		"http-server-address = localhost:8889",
		"p2p-listen-endpoint = localhost:9877",	"agent-name = \"EOS Test Agent\"", 			"plugin = eosio::net_api_plugin",
		"plugin = eosio::net_plugin",			"plugin = eosio::wallet_plugin",			"plugin = eosio::wallet_api_plugin",
		"plugin = eosio::history_plugin", 		"plugin = eosio::http_plugin",}

	confData := combineConfig(constantEntries)
	for _,server := range servers {
		for _,ip := range server.ips {
			confData += fmt.Sprintf("p2p-peer-address = %s:9876\n",ip)
		}
	}
	write("config.init",confData)
}

/**
 * Copies over the config.init file to each node
 * @param  []Server	servers	The list of servers
 */
func eos_initNodes(servers []Server){
	node := 0
	for _,server := range servers {
		_scp(server.addr,"./config.init","/home/appo/config.init")
		for i := 0; i < len(server.ips); i++ {
			sshExec(server.addr,fmt.Sprintf("docker cp /home/appo/config.init whiteblock-node%d:/config.init",node))
			node++
		}
	}	
}

/**
 * Start the keos daemon on each node
 * @param  []Server	servers	The list of servers
 */
func eos_startKeos(servers []Server){
	node := 0
	for _,server := range servers {
		for _,ip := range server.ips {
			sshExec(server.addr,fmt.Sprintf("docker exec -d whiteblock-node%d keosd --http-server-address %s:8899",node,ip))
			node++
		}
	}		
}

/**
 * Get the wallets for all nodes
 * @param  []Server	servers	The list of servers
 */
func eos_getWallets(servers []Server) map[string]string {
	wallets := make(map[string]string)
	node := 0
	for _,server := range servers {
		for _,ip := range server.ips {
			data := sshExec(server.addr,fmt.Sprintf("docker exec whiteblock-node%d cleos --wallet-url http://%s:8899 wallet create --to-console | tail -n 1",node,ip))
			wallets[ip] = data[1:len(data)-1]
			node++
		}
	}
	return wallets
}

/**
 * Get the key pairs for all nodes
 * @param  []Server	servers	The list of servers
 */
func eos_getKeyPairs(servers []Server) map[string]KeyPair {
	keyPairs := make(map[string]KeyPair)
	node := 0
	for _,server := range servers {
		for _,ip := range server.ips {
			data := sshExec(server.addr,fmt.Sprintf("docker exec whiteblock-node%d cleos --wallet-url http://%s:8899 create key --to-console | awk '{print $3}'",node,ip))
			//fmt.Printf("RAW KEY DATA%s\n",data)
			keyPair := strings.Split(data,"\n")
			keyPairs[ip] = KeyPair{privateKey:keyPair[0],publicKey:keyPair[1]}
			//fmt.Printf("Extracted key pair: %+v\n",keyPairs[ip])
			node++
		}
	}
	return keyPairs
}

/**
 * Sets up a given node as an EOS Producer
 * @param  string	serverIP	The IP address of the host server
 * @param  string	ip			The IP of the node
 * @param  KeyPair 	keyPair		The node's key pair
 * @param  int		num			The node's number
 * @param  string	ptpFlags	The peering flags for nodeos
 */
func eos_setupProducer(serverIP string,ip string,keyPair KeyPair,num int, ptpFlags string) {
	staticFlags := "--plugin eosio::chain_api_plugin --plugin eosio::net_api_plugin "
	if num == 0 {
		staticFlags = "--enable-stale-production " + staticFlags
	}

	setupCmd := fmt.Sprintf("nodeos %s --producer-name %s --http-server-address %s:8889 --config-dir node%d --data-dir node%d %s %s",
				staticFlags,
				eos_getProducerName(num),
				ip,
				num,
				num,
				ptpFlags,
				eos_getKeyPairFlag(keyPair))

	sshExec(serverIP,fmt.Sprintf("docker exec -d whiteblock-node%d %s",num,setupCmd))

}

/**
 * Finish the setup of the first producer for EOS
 * @param  string	serverIP	The IP address of the host server
 * @param  string	ip			The IP address of the node
 * @param  KeyPair	keyPair		The key pair of the node
 */
func eos_finishFirstProducer(serverIP string, ip string,keyPair KeyPair) {
	name := "whiteblock-node0"

	sshExec(serverIP,
		fmt.Sprintf("docker exec %s cleos --wallet-url http://%s:8899 -u http://%s:8889 wallet import --private-key %s ",name,ip,ip,keyPair.privateKey))

	sshExec(serverIP,
		fmt.Sprintf("docker exec %s cleos --wallet-url http://%s:8899 -u http://%s:8889 wallet import --private-key %s ",name,ip,ip,GENESIS_KEY))

	sshExec(serverIP,
		fmt.Sprintf("docker exec %s cleos --wallet-url http://%s:8899 -u http://%s:8889 set contract eosio /contracts/eosio.bios",name,ip,ip))
			
		
}

/**
 * Create an account on a node
 * @param  string	serverIP		The IP address of the host machine
 * @param  string	ip				The IP address of the node
 * @param  KeyPair	keyPair			The key pair of the node
 * @param  int		num				The relative number of the node
 * @param  string	masterNodeIP	The IP of the node named eosio, the first producer node
 */
func eos_createAccount(serverIP string,ip string, keyPair KeyPair,num int,masterNodeIP string) {
	name := fmt.Sprintf("whiteblock-node%d",num)

	sshMultiExec(serverIP,
		fmt.Sprintf("docker exec %s cleos --wallet-url http://%s:8899 wallet import --private-key %s",name,ip,GENESIS_KEY),
		fmt.Sprintf("docker exec %s cleos --wallet-url http://%s:8899 wallet import --private-key %s",name,ip,keyPair.privateKey),
		fmt.Sprintf("docker exec %s cleos --wallet-url http://%s:8899 -u http://%s:8889 create account eosio %s %s",name,masterNodeIP,masterNodeIP,eos_getProducerName(num),keyPair.publicKey),
		)
}

/**
 * Start producer scheduling on a node
 * @param  string	serverIP		The IP address of the host machine
 * @param  KeyPair	keyPair			The key pair of the node
 * @param  int		num				The relative number of the node
 * @param  string	masterNodeIP	The IP of the node named eosio, the first producer node
 * @param  string	masterPublicKey	The public key of the node named eosio, the first producer node
 */
func eos_startScheduling(serverIP string,keyPair KeyPair,num int,masterNodeIP string,masterPublicKey string){
	name := fmt.Sprintf("whiteblock-node%d",num)

	scheduleCmd := fmt.Sprintf(`cleos --wallet-url http://%s:8899 -u http://%s:8889 push action eosio setprods "{ \"schedule\": [{\"producer_name\": \"eosio\",\"block_signing_key\": \"%s\"}, {\"producer_name\": \"%s\",\"block_signing_key\": \"%s\"}]}" -p eosio@active`,
	masterNodeIP,masterNodeIP,masterPublicKey,eos_getProducerName(num),keyPair.publicKey)

	sshExec(serverIP,
		fmt.Sprintf("docker exec %s %s",name,scheduleCmd))
}


func eos_getProducerName(num int) string {
	if num == 0 {
		return "eosio"
	}
	return fmt.Sprintf("producer%d",num)
}


func eos_getPTPFlags(servers []Server) string {
	flags := ""
	for _,server := range servers {
		for _,ip := range server.ips {
			flags += fmt.Sprintf("--p2p-peer-address %s:9876 ",ip)
		}
	}
	return flags
}

func eos_getKeyPairFlag(keyPair KeyPair) string {
	return fmt.Sprintf("--private-key [\"%s\",\"%s\"]",keyPair.publicKey,keyPair.privateKey)
}