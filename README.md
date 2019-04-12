#Genesis

[![Maintainability](https://api.codeclimate.com/v1/badges/4aae8bc608f619f0990e/maintainability)](https://codeclimate.com/repos/5cb0ac61f5eb734088003e71/maintainability)

##Installation
* clone repository
* `cd router`
* `go get`
* `go build`

##Configuration
Configuration options are located in `config.json` in the same directory as the binary

* __ssh-user__: The default username for ssh
* __ssh-password__: The default password for ssh
* __listen__: The socket to listen on 
* __rsa-key__: The location of the ssh private key
* __rsa-user__: The corresponding username for that private key
* __verbose__: Enable or disable verbose mode
* __server-bits__: The bits given to each server's number
* __cluster-bits__: The bits given to each clusters's number
* __node-bits__: The bits given to each nodes's number
* __thread-limit__: The maximum number of threads that can be used for building
* __ip-prefix__: Used for the IP Scheme
* __docker-output-file__: The location instead the docker containers where the clients stdout and stderr will be captured
* __influx__: The influxdb endpoint
* __influx-user__: The influx auth username
* __influx-password__: The influx auth password
* __service-network__: CIDR of the network for the services
* __service-network-name__: The name for the service network
* __node-prefix__: The prefix for each node name
* __node-network-prefix__: The prefix for each cluster network
* __service-prefix__: The prefix for each service

* __nodes-public-key__: Location of the public key for the nodes
* __nodes-private-key__: Location of the private key for the nodes
* __handle-node-ssh-keys__: Should genesis handle the nodes ssh keys?
* __max-nodes__: Set a maximum number of nodes that a client can build
* __max-node-memory__: Set the max memory per node that a client can use
* __max-node-cpu__: Set the max cpus per node that a client can use
      

###Config Environment Overrides
These will override what is set in the config.json file, and allow configuration via
only ENV variables

* `SSH_USER`
* `SSH_PASSWORD`
* `LISTEN`
* `RSA_KEY`
* `RSA_USER`
* `VERBOSE` (only need to set it)
* `SERVER_BITS`
* `CLUSTER_BITS`
* `NODE_BITS`
* `THREAD_LIMIT`
* `IP_PREFIX`
* `DOCKER_OUTPUT_FILE`
* `INFLUX`
* `INFLUX_USER`
* `INFLUX_PASSWORD`
* `SERVICE_NETWORK`
* `SERVICE_NETWORK_NAME`
* `NODE_PREFIX`
* `NODE_NETWORK_PREFIX`
* `SERVICE_PREFIX`
* `NODES_PUBLIC_KEY`
* `NODES_PRIVATE_KEY`
* `HANDLE_NODE_SSH_KEYS` (only need to set it)
* `MAX_NODES`
* `MAX_NODE_MEMORY`
* `MAX_NODE_CPU`

###Additional Information
* Config order of priority ENV -> config file -> defaults
* `ssh-user`,`ssh-password` and `rsa-user`, `rsa-key` are both used, starting with pass auth then falling back to key auth


##IP Scheme
We are using ipv4 so each address will have 32 bits.

The following assumptions will be made

* Each server will have a relatively unique `serverId`
* This uniqueness need only apply to servers which will contain nodes which communicate with each other
* There are going to be 3 ip addresses reserved from each subnet
* Nodes in the same docker network are able to route between each other by default

For simplicity, the following variables will be used

* __A__ = `ip-prefix`
* __B__ = `server-bits`
* __C__ = `cluster-bits`
* __D__ = `node-bits`

Note the following rules

* __A__,__B__,__C__, and __D__ must be greater than 0
* ceil(log2(__A__)) + __B__ + __C__ + __D__ <= 32
* __D__ must be atleast 2
* (2^__B__) = The maximum number of servers
* (2^__C__) = The number of cluster in a given server
* (2^__D__ - 3) = How many nodes are groups together in each cluster
* (2^__D__ - 3) * (2^__C__) = The max number of nodes on a server
* (2^__D__ - 3) * (2^__C__) * (2^__B__) = The maximum number of nodes that could be on the platform

###What is a cluster?

Each cluster corresponds to a subnet,docker network,and vlan. 
Containers in the same cluster will have minimal latency applied to them. In the majority of cases
it is best to just have one node per cluster, allowing for latency control between all of the nodes.

###How is it all calculated?

Given a node number __X__ and a `serverId` of __Y__,
Let __Z__ be the cluster number,
__I__ be the generated IP in big-endian,
and the earlier mentioned variables applied

__Z__ = floor(__X__ / (2^__D__ - 3)))
__I__ = (__A__ * 2^(__B__+__C__+__D__) ) + ( __Y__ * 2^(__B__+__C__) ) + (__Z__ * 2^__C__) + (__X__ % (2^__D__ - 3) + 2)

if __Z__ == (2^__C__ - 1) then __I__ = __I__ - 2

####Explaination
First get the cluster the node is in
Then construct the IP one segment at a time through addition
Due to the restrictions, each piece will fit neatly into place without overlap
Finally, check if it is not the last cluster on the server,
add 1 to the ip address if it is not the last cluster. 

####Example
Given a node number(__X__) of 2 and a `serverId`(__Y__) of 3
Given the IP Scheme of __A__ = 10, __B__ = 8, __C__ = 14, __D__ = 2
__Z__ = floor(2/(2^2 - 3))
__Z__ = 2
It is going to be in cluster 2
Now, for the construction of the IP
Visually, it can be represented as 
IP = AAAAAAAA BBBBBBBB CCCCCCCC CCCCCCDD
The values are simply placed inside the bit space of the IP address as represented,
with the exception of the __D__ bits, which needs to be calculated
calculate this number as (2 % (2^2 - 3) + 2) or 2
Then since (__Z__ != (2^__C__ - 1)) 2 != 16383, the value remains 2
Finally, construct IP
Part A = 00001010
Part B = 00000011
Part C = 00000000000010
Part D = 10
IP = 00001010 00000011 00000000000010 10
   = 00001010 00000011 00000000 00001010
   = 10       3        0        10

The gateway is calculated in a similar way, except take Part D to always equal 1

Gateway IP = 00001010 00000011 00000000000010 01
           = 00001010 00000011 00000000 00001001
           = 10       3        0        9

Finally the subnet is 32 - __D__

Resulting in
IP = 10.3.0.10
Gateway IP = 10.3.0.9
Subnet = 10.3.0.8/30



##REST API

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
curl -X PUT http://localhost:8000/servers/foxtrot -d \
'{"addr":"172.16.6.5","iaddr":{"ip":"10.254.6.100","gateway":"10.254.6.1","subnet":24},
"nodes":0,"max":10,"serverID":6,"id":-1,"iface":"eth0","switches":[{"addr":"172.16.1.1","iface":"eno3","brand":1,"id":5}],"ips":null}}'
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
curl -X UPDATE http://localhost:8000/servers/5 -d \
 '{"addr":"172.16.4.5","iaddr":{"ip":"10.254.4.100","gateway":"10.254.4.1","subnet":24}, 
 "nodes":0,"max":30,"id":5,"serverID":4,"iface":"eno3","switches":[{"addr":"172.16.1.1","iface":"eth4","brand":1,"id":3}],"ips":null}'
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
"resources":[{"cpus":"2.0","memory":"10gb"}],"environments"{"BLAH":"blah"},"files":{"config.json":"base64 of file"},"params":null}'
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
        "id":(string),
        "testNetId":(int),
        "server":(int),
        "localId":(int),
        "ip":(string)
    },...
]
```


###GET /status/nodes/{testnetid}
Get the nodes that are running in the given testnet
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
[
    {
        "name":"whiteblock-node0",
        "server":4,
        "cpu":1.4,
        "up":true
    },...
]
```
#####EXAMPLE
```bash
curl -XGET http://localhost:8000/status/nodes/
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

###GET /log/{server}/{node}
Get both stdout and stderr from the blockchain process

#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
<The contents>
```

#####EXAMPLE
```bash
curl -X POST http://localhost:8000/exec/4/0 -d 'ls'
```

###GET /nodes/{testnetid}
Get the nodes for the latest testnet
#####RESPONSE
```json
[
    {
        "id": 1647,
        "testNetId": 134,
        "server": 3,
        "localId": 0,
        "ip": "10.6.0.2",
        "label": ""
    },
    {
        "id": 1648,
        "testNetId": 134,
        "server": 3,
        "localId": 1,
        "ip": "10.6.0.6",
        "label": ""
    }
]
```
#####EXAMPLE
```bash
curl -X GET http://localhost:8000/nodes
```

###DELETE /build/{buildid}
Stop the given build
#####RESPONSE
`Stop signal has been sent...`
#####EXAMPLE
```bash
curl -X DELETE http://localhost:8000/build
```

###POST /nodes/{num}
Add {num} nodes to the current test net (Body is optional)
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
curl -X POST http://localhost:8000/nodes/3 -d '{"servers":[3],"blockchain":"ethereum","nodes":3,"image":"ethereum:latest",
"resources":{"cpus":"2.0","memory":"10gb"},"params":null}'
```


###DELETE /nodes/{num}
Delete {num} nodes from the testnet
#####BODY
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
Success
```
#####EXAMPLE
```bash
curl -X DELETE http://localhost:8000/nodes/5 
```


###DELETE /emulate/{testnetId}
Turn off emulate for a whole testnet 
####BODY

####EXAMPLE
```bash
curl -X DELETE http://localhost:8000/emulate/9e09efe8_d7a3_4429_832c_447d876194c8
```

###POST /emulate/{testnetId}
Set emulation for a node or nodes
####BODY
TODO
####EXAMPLE
```bash
curl -X POST http://localhost:8000/emulate/9e09efe8_d7a3_4429_832c_447d876194c8
```

###POST /emulate/all/{testnetId}
Set emulation for a whole testnet
####BODY
TODO
####EXAMPLE
```bash
curl -X POST http://localhost:8000/emulate/all/9e09efe8_d7a3_4429_832c_447d876194c8
```

###GET /resources/{blockchain}

####BODY
TODO
####EXAMPLE
```bash
curl -X POST http://localhost:8000/resources/geth
```

###GET /resources/{blockchain}/{file}

####BODY
TODO
####EXAMPLE
```bash
curl -X GET http://localhost:8000/resources/geth/genesis.json
```

###POST /resources/{blockchain}/{file}

####BODY
TODO
####EXAMPLE
```bash
curl -X POST http://localhost:8000/resources/geth/genesis.json
```



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