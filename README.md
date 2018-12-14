
##REST API
__Warning: Do not use any path marked PRIVATE, they will begin to require credentials in the near future__

###GET /servers/
Get the current registered servers
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
{
	"server_name":{
		"addr":(string),
		"iaddr":{
			"ip":(string),
			"gateway":(string),
			"subnet":(int)
		},
		"nodes":(int),
		"max":(int),
		"id":(int),
		"serverID":(int),
		"iface":(string),
		"switches":[
			{
				"addr":(string),
				"iface":(string),
				"brand":(int),
				"id":(int)
			}
		]
	},
	"server2_name":{...}...
}
```
#####EXAMPLE
```
curl -XGET http://localhost:8000/servers/
```

###PUT /servers/{name}
####PRIVATE
Register and add a new server to be 
controlled by the instance
#####BODY
```
{
	"addr":(string),
	"iaddr":{
		"ip":(string),
		"gateway":(string),
		"subnet":(int)
	},
	"nodes":(int),
	"max":(int),
	"id":-1,
	"serverID":(int),
	"iface":(string),
	"switches":[
		{
			"addr":(string),
			"iface":(string),
			"brand":(int),
			"id":(int)
		}
	]
}
```
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
<server id>
```

#####EXAMPLE
```bash
curl -X PUT http://localhost:8000/servers/foxtrot -d '{"addr":"172.16.6.5","iaddr":{"ip":"10.254.6.100","gateway":"10.254.6.1","subnet":24},"nodes":0,"max":10,"serverID":6,"id":-1,"iface":"eth0","switches":[{"addr":"172.16.1.1","iface":"eno3","brand":1,"id":5}],"ips":null}}'
```

###GET /servers/{id}
Get a server by id
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
{
	"addr":(string),
	"iaddr":{
		"ip":(string),
		"gateway":(string),
		"subnet":(int)
	},
	"nodes":(int),
	"max":(int),
	"id":(int),
	"serverID":(int),
	"iface":(string),
	"switches":[
		{
			"addr":(string),
			"iface":(string),
			"brand":(int),
			"id":(int)
		}
	]
}
```

###DELETE /servers/{id}
####PRIVATE
Delete a server
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
Success
```

#####EXAMPLE
```bash
curl -X DELETE http://localhost:8000/servers/5
```

###UPDATE /servers/{id}
####PRIVATE
Update server information
#####BODY
```
{
	"addr":(string),
	"iaddr":{
		"ip":(string),
		"gateway":(string),
		"subnet":(int)
	},
	"nodes":(int),
	"max":(int),
	"id":(int),
	"serverID":(int),
	"iface":(string),
	"switches":[
		{
			"addr":(string),
			"iface":(string),
			"brand":(int),
			"id":(int)
		}
	]
}
```
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
Success
```

#####EXAMPLE
```bash
curl -X UPDATE http://localhost:8000/servers/5 -d '{"addr":"172.16.4.5","iaddr":{"ip":"10.254.4.100","gateway":"10.254.4.1","subnet":24},"nodes":0,"max":30,"id":5,"serverID":4,"iface":"eno3","switches":[{"addr":"172.16.1.1","iface":"eth4","brand":1,"id":3}],"ips":null}'
```

###GET /testnets/
Get all testnets which are currently running
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
[
	{
		"id":(int),
		"blockchain":(string),
		"nodes":(int),
		"image":(string)
	},...

]
```

###POST /testnets/
Add and deploy a new testnet
#####BODY
```
{
	"servers":[(int),(int)...],
	"blockchain":(string),
	"nodes":(int),
	"image":(string),
	"resources":{
		"cpus":(string),
		"memory":(string)
	},
	"params":(Object containing params specific to the chain/client being built)
}
```
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
Success
```
#####EXAMPLE
```bash
curl -X POST http://localhost:8000/testnets/ -d '{"servers":[3],"blockchain":"ethereum","nodes":3,"image":"ethereum:latest",
"resources":{"cpus":"2.0","memory":"10gb"},"params":null}'
```

###GET /testnets/{id}
Get data on a single testnet
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
{
	"id":(int),
	"blockchain":(string),
	"nodes":(int),
	"image":(string)
}
```

###GET /testnets/{id}/nodes/
####PRIVATE
Get the nodes in a testnet
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
[
	{
		"id":(int),
		"testNetId":(int),
		"server":(int),
		"localId":(int),
		"ip":(string)
	},...
]
```


###GET /status/nodes/
Get the nodes that are running in the latest testnet
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
[
	{
		"name":"whiteblock-node0",
		"server":4
	},...
]
```
#####EXAMPLE
```bash
curl -XGET http://localhost:8000/status/nodes/
```



###POST /exec/{server}/{node}
###Temporarily disabled
Execute a command on a given node
#####BODY
```
<bash command>
```
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
<command results>
```

#####EXAMPLE
```bash
curl -X POST http://localhost:8000/exec/4/0 -d 'ls'
```


###GET /params/{blockchain}/
Get the build params for a blockchain
#####RESPONSE
```json
[
	{"chainId":"int"},
	{"networkId":"int"},
	{"difficulty":"int"},
	{"initBalance":"string"},
	{"maxPeers":"int"},
	{"gasLimit":"int"},
	{"homesteadBlock":"int"},
	{"eip155Block":"int"},
	{"eip158Block":"int"}
]
```
#####EXAMPLE
```bash
curl -X GET http://localhost:8000/params/ethereum
```

###GET /defaults/{blockchain}
Get the default parameters for a blockchain
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
{
	"chainId":15468,
	"networkId":15468,
	"difficulty":100000,
	"initBalance":100000000000000000000,
	"maxPeers":1000,
	"gasLimit":4000000,
	"homesteadBlock":0,
	"eip155Block":0,
	"eip158Block":0
}
```
#####EXAMPLE
```bash
curl -X GET http://localhost:8000/defaults/ethereum
```


##Configuration
Configuration options are located in `config.json` in the same directory as the binary

* __builder__: The application to use to build the nodes
* __ssh-user__: The default username for ssh
* __ssh-password__: The default password for ssh
* __vyos-home-dir__: The location to put the vyos script
* __rsa-key__: The location of the ssh private key
* __rsa-user__: The corresponding username for that private key
* __verbose__: Enable or disable verbose mode
* __server-bits__: The bits given to each server's number
* __cluster-bits__: The bits given to each clusters's number
* __node-bits__: The bits given to each nodes's number
* __thread-limit__: The maximum number of threads that can be used for building


##Blockchain Specific Parameters

###Geth (Go-Ethereum)
__Note:__ Any configuration option can be left out, and this entire section can even be null,
the example contains all of the defaults

####Options
* `chainId`: The chain id set in the genesis.conf
* `networkId`: The network id
* `difficulty`: The initial difficulty set in the genesis.conf file
* `initBalance`: The initial balance for the accounts
* `maxPeers`: The maximum number of peers for each node
* `gasLimit`: The initial gas limit
* `homesteadBlock`: Set in genesis.conf
* `eip155Block`: Set in genesis.conf
* `eip158Block`: Set in genesis.conf

####Example (using defaults)
```json
{
	"chainId":15468,
	"networkId":15468,
	"difficulty":100000,
	"initBalance":100000000000000000000,
	"maxPeers":1000,
	"gasLimit":4000000,
	"homesteadBlock":0,
	"eip155Block":0,
	"eip158Block":0
}
```
###Syscoin (RegTest)

####Options
* `rpcUser`: The username credential
* `rpcPass`: The password credential
* `masterNodeConns`: The number of connections to set up for the master nodes
* `nodeConns`:  The number of connections to set up for the normal nodes
* `percentMasternodes`: The percentage of the network consisting of master nodes

* `options`: Options to set enabled for all nodes
* `senderOptions`: Options to set enabled for senders
* `receiverOptions`: Options to set enabled for receivers
* `mnOptions`: Options to set enabled for master nodes

* `extras`: Extra options to add to the config file for all nodes
* `senderExtras`: Extra options to add to the config file for senders
* `receiverExtras`: Extra options to add to the config file for receivers
* `mnExtras`: Extra options to add to the config file for master nodes

####Example (using defaults)
```json
{
	"rpcUser":"username",
	"rpcPass":"password",
	"masterNodeConns":25,
	"nodeConns":8,
	"percentMasternodes":90,
	"options":[
		"server",
		"regtest",
		"listen",
		"rest"
	],
	"senderOptions":[
		"tpstest",
		"addressindex"
	],
	"mnOptions":[],
	"receiverOptions":[
		"tpstest"
	],
	"extras":[],
	"senderExtras":[],
	"receiverExtras":[],
	"mnExtras":[]
}
```
