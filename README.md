Genesis
======

[![Maintainability](https://api.codeclimate.com/v1/badges/4aae8bc608f619f0990e/maintainability)](https://codeclimate.com/repos/5cb0ac61f5eb734088003e71/maintainability)

# Installation
* clone repository
* `cd genesis`
* `go get`
* `go build`

# Configuration
Configuration options are located in `config.json` in the same directory as the binary

* __ssh-user__: The default username for ssh
* __ssh-key__: The location of the ssh private key
* __listen__: The socket to listen on 
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
      

## Config Environment Overrides
These will override what is set in the config.json file, and allow configuration via
only ENV variables

* `RSA_USER`
* `RSA_KEY`
* `LISTEN`
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

## Additional Information
* Config order of priority ENV -> config file -> defaults
* `ssh-user`,`ssh-password` and `rsa-user`, `rsa-key` are both used, starting with pass auth then falling back to key auth


# IP Scheme
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

### What is a cluster?

Each cluster corresponds to a subnet,docker network,and vlan. 

Containers in the same cluster will have minimal latency applied to them. In the majority of cases
it is best to just have one node per cluster, allowing for latency control between all of the nodes.

### How is it all calculated?

Given a node number __X__ and a `serverId` of __Y__,
Let __Z__ be the cluster number,
__I__ be the generated IP in big-endian,
and the earlier mentioned variables applied

__Z__ = floor(__X__ / (2^__D__ - 3)))
__I__ = (__A__ * 2^(__B__+__C__+__D__) ) + ( __Y__ * 2^(__B__+__C__) ) + (__Z__ * 2^__C__) + (__X__ % (2^__D__ - 3) + 2)

if __Z__ == (2^__C__ - 1) then __I__ = __I__ - 2

#### Explaination
First get the cluster the node is in

Then construct the IP one segment at a time through addition

Due to the restrictions, each piece will fit neatly into place without overlap

Finally, check if it is not the last cluster on the server,

add 1 to the ip address if it is not the last cluster. 

#### Example
Given a node number(__X__) of 2 and a `serverId`(__Y__) of 3

Given the IP Scheme of `__A__ = 10, __B__ = 8, __C__ = 14, __D__ = 2`
```
__Z__ = floor(2/(2^2 - 3))
__Z__ = 2
```
It is going to be in cluster 2

Now, for the construction of the IP

Visually, it can be represented as
```
IP = AAAAAAAA BBBBBBBB CCCCCCCC CCCCCCDD
```
The values are simply placed inside the bit space of the IP address as represented,
with the exception of the __D__ bits, which needs to be calculated

calculate this number as `(2 % (2^2 - 3) + 2)` or 2

Then since `(__Z__ != (2^__C__ - 1)) 2 != 16383`, the value remains 2

Finally, construct IP
```
Part A = 00001010
Part B = 00000011
Part C = 00000000000010
Part D = 10
IP = 00001010 00000011 00000000000010 10
   = 00001010 00000011 00000000 00001010
   = 10       3        0        10
```

The gateway is calculated in a similar way, except take Part D to always equal 1

```
Gateway IP = 00001010 00000011 00000000000010 01
           = 00001010 00000011 00000000 00001001
           = 10       3        0        9
```

Finally the subnet is 32 - __D__

Resulting in
```
IP = 10.3.0.10
Gateway IP = 10.3.0.9
Subnet = 10.3.0.8/30
```

# REST API

## GET /servers/
Get the current registered servers

### RESPONSE
```
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

### EXAMPLE
```bash
curl -XGET http://localhost:8000/servers/
```


## PUT /servers/{name}
Register and add a new server to be 
controlled by the instance

### BODY
```
{
    "addr":(string),
    "nodes":(int),
    "max":(int),
    "id":-1,
    "subnetID":(int)
}
```

### RESPONSE
```
<server id>
```

### EXAMPLE
```bash
curl -X PUT http://localhost:8000/servers/foxtrot -d \
'{"addr":"172.16.6.5","nodes":0,"max":10,"subnetID":6,"id":-1}'
```


## GET /servers/{id}
Get a server by id

### RESPONSE
```
{
    "addr":(string),
    "nodes":(int),
    "max":(int),
    "id":(int),
    "subnetID":(int)
}
```

## DELETE /servers/{id}
Remove a server

### RESPONSE
```
Success
```

### EXAMPLE
```bash
curl -X DELETE http://localhost:8000/servers/5
```

## UPDATE /servers/{id}
Update server information

### BODY
```
{
    "addr":(string),
    "nodes":(int),
    "max":(int),
    "id":(int),
    "subnetID":(int)
}
```
### RESPONSE
```
Success
```

### EXAMPLE
```bash
curl -X UPDATE http://localhost:8000/servers/5 -d \
 '{"addr":"172.16.4.5","nodes":0,"max":30,"id":5,"subnetID":4}'
```

## POST /testnets/
Add and deploy a new testnet

### BODY
```
<json object representing the build, see example and details after it>
```

### RESPONSE
```
Success
```

### EXAMPLE
```bash
curl -X POST http://localhost:8000/testnets/ -d '{
    "servers":[1],
    "blockchain":"ethereum",
    "nodes":3,
    "image":["ethereum:latest"],
    "resources":[{
        "cpus":"2.5",
        "memory":"12gb"
    }],
    "params":{
        "networkId":15468,
        "difficulty":100000,
        "initBalance":"100000000000000000000",
        "maxPeers":1000,
        "gasLimit":4000000,
        "homesteadBlock":0,
        "eip155Block":10,
        "eip158Block":10
    },
    "environments":[
        {
            "NODE":"0"
        },
        {
            "NODE":"1"
        }
    ],
    "files":[
        {
            "config.toml":"TG9yZW0gabnNlY3RldHVyIGFkaXBpc2NpbmcgZdyxX=="
        },
        {
            "config.toml":"UHJvaW4gZmluaWJ1cyBsYW9yZWV0IHRpbmNpZHVudA=="
        }
    ],
    "logs":[],
    "extras":{
        "defaults":{
            "files":{
                "config.toml":"IEludGVnZXIgYXVjdG9yIHVybmEgbGFvcmVldCBjb252YWxsaXMgdmVzdGlidWx1bS4="
            }
        },
        "postbuild":{
            "ssh":{
                "pubKeys":[]
            }
        },
        "prebuild":{
            "auth":{
                "username":"bill",
                "password":"big_strong_bill"
            },
            "build":false,
            "dockerfile":null,
            "freezeAfterInfrastructure":false,
            "pull":false
        }
    }
}'
```

### DETAILS
* servers : The servers on which to build
* blockchain: The blockchain to build out
* nodes: The total number of nodes to build
* images: The docker images to use in building the nodes, the first image in the list will be used as the default.
* resources: The first resource object is the default.
  * cpus: The max number of cpus which can be used by the node.
  * memory: The maximum amount of RAM that a node can use.
* params: Blockchain specific parameters to supplement the build
* environments: The environmental variables for the nodes.
* files: The file templates to replace the internal files, key is the file name, value is the file data base64 encoded.
* logs: The log files for each node. 
* extras: Extra build information which doesn't fit into any category. Most trivial expansions are done here
* defaults: Contains the default values for certain fields. Used for cases where you might want to differentiate between
 all nodes and just the first node.
* postbuild: Contains details for after infrastructure deployment functionality
  * ssh: Information on addition ssh credentials to allow access to the nodes. 
* prebuild:
  * auth: Docker login authorization credentials (if needed)
  * build: Whether or not it should build from a dockerfile
  * dockerfile: The dockerfile encoded in base64, which will be built if build is true
  * freezeAfterInfrastructure: Freeze after the context switch from building infrastructure to blockchain genesis ceremony
  * pull: Force an update of all of the used images. 


## DELETE /testnets/{id}
Tears down a testnet

### RESPONSE
```
Success
```

## GET /testnets/{id}/nodes/
Get the nodes in a testnet

### RESPONSE
```
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


## GET /status/nodes/{testnetid}
Get the nodes that are running in the given testnet

### RESPONSE
```json
[
    {
    "ip": "10.1.0.2",
    "name": "whiteblock-node0",
    "resourceUse": {
      "cpu": 1.5,
      "residentSetSize": 629700,
      "virtualMemorySize": 40105576
    },
    "server": 1,
    "up": true
  }
]
```

### EXAMPLE
```bash
curl -XGET http://localhost:8000/status/nodes/
```

## GET /params/{blockchain}/
Get the build params for a blockchain

### RESPONSE
```json
[
    ["chainId","int"]
    ["homesteadBlock","int"],
    ["eip155Block","int"],
    ["eip158Block","int"]
]
```

### EXAMPLE
```bash
curl -X GET http://localhost:8000/params/ethereum
```

## GET /defaults/{blockchain}
Get the default parameters for a blockchain

### RESPONSE
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

### EXAMPLE
```bash
curl -X GET http://localhost:8000/defaults/ethereum
```

## GET /log/{server}/{node}
Get both stdout and stderr from the blockchain process

### RESPONSE
```
<The contents>
```

### EXAMPLE
```bash
curl -X POST http://localhost:8000/log/4/0
```

## GET /nodes/{testnetid}
Get the nodes for the latest testnet

### RESPONSE
```json
[
    {
        "id": 1647,
        "testnetId": 134,
        "server": 3,
        "localId": 0,
        "ip": "10.6.0.2",
        "label": "",
        "absNum":4,
        "image":"geth:latest",
        "blockchain":"geth"
    }
]
```

### EXAMPLE
```bash
curl -X GET http://localhost:8000/nodes
```

## DELETE /build/{buildid}
Stop the given build

### RESPONSE
`Stop signal has been sent...`

### EXAMPLE
```bash
curl -X DELETE http://localhost:8000/build
```

## POST /nodes/{testnetID}
Appends nodes to the given testnet otherwise, acts like `POST /testnet/`, main difference is the response is not
the testnet id

### RESPONSE
```
Success
```


## DELETE /nodes/{testnetId}/{num}
Delete {num} nodes from the testnet

### RESPONSE
```
Success
```

### EXAMPLE
```bash
curl -X DELETE http://localhost:8000/nodes/9e09efe8_d7a3_4429_832c_447d876194c8/5 
```


## DELETE /emulate/{testnetId}
Turn off emulate for a whole testnet

### EXAMPLE
```bash
curl -X DELETE http://localhost:8000/emulate/9e09efe8_d7a3_4429_832c_447d876194c8
```

## POST /emulate/{testnetId}
Set emulation for a node or nodes

### BODY
```json
[{"node":1,"limit":1000,"loss":0,"delay":5000,"rate":"","duplicate":0,"corrupt":0,"reorder":0},
 {"node":2,"limit":1000,"loss":0,"delay":5000,"rate":"","duplicate":0,"corrupt":0,"reorder":0},
 {"node":0,"limit":1000,"loss":0,"delay":5000,"rate":"","duplicate":0,"corrupt":0,"reorder":0}]
```

### EXAMPLE
```bash
curl -X POST http://localhost:8000/emulate/9e09efe8_d7a3_4429_832c_447d876194c8
```

## POST /emulate/all/{testnetId}
Set emulation for a whole testnet

### BODY
```json
{"limit":1000,"loss":0,"delay":5000,"rate":"","duplicate":0,"corrupt":0,"reorder":0}
```

### EXAMPLE
```bash
curl -X POST http://localhost:8000/emulate/all/9e09efe8_d7a3_4429_832c_447d876194c8 
```

## GET /resources/{blockchain}
Get the static file resources used by genesis for the given blockchain

### RESPONSE
```json
[
  "defaults.json",
  "genesis.json",
  "params.json"
]
```

### EXAMPLE
```bash
curl -X POST http://localhost:8000/resources/geth
```

## GET /resources/{blockchain}/{file}
Gets the contents of that file resource for the given blockchain

### EXAMPLE
```bash
curl -X GET http://localhost:8000/resources/geth/genesis.json
```


## GET /build

## GET /build/{id}

## POST /build/freeze/{id}
Pause the given build

## POST /build/thaw/{id}
Unpause the given build

## DELETE /build/freeze/{id}
Unpause the given build

## GET /emulate/{testnetID}
Get the current network conditions for the testnet

## POST /nodes/restart/{testnetID}/{num}
Restart a node on a testnet

## POST /nodes/raise/{testnetID}/{node}/{signal}
Send a signal to the main process of the given node 

## POST /nodes/kill/{testnetID}/{node}
Attempt to kill the given node

## POST /outage/{testnetID}/{node1}/{node2}
Prevent the given node1 and node2 from establishing a connection with each other

## DELETE /outage/{testnetID}/{node1}/{node2}
Allow the given node1 and node2 to establish a connection with each other

## DELETE /outage/{testnetID}
Remove all blocked connections from a testnet

## GET /outage/{testnetID}
Get the currently blocked connections

## GET /outage/{testnetID}/{node}
Get the blocked connections for the given node

## POST /outage/partition/{testnetID}
Create a network partition on a testnet

## GET /outage/partition/{testnetID}
Get the partitions on the testnet

## GET /blockchains
Get the currently supported blockchains by genesis



# Blockchain Specific Parameters

## Geth (Go-Ethereum)
__Note:__ Any configuration option can be left out, and this entire section can even be null,
the example contains all of the defaults

### Options
* `networkId`: The network id
* `difficulty`: The initial difficulty set in the genesis.conf file
* `initBalance`: The initial balance for the accounts
* `maxPeers`: The maximum number of peers for each node
* `gasLimit`: The initial gas limit
* `homesteadBlock`: Set in genesis.conf
* `eip155Block`: Set in genesis.conf
* `eip158Block`: Set in genesis.conf

### Example (using defaults)
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
## Syscoin (RegTest)

### Options
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

### Example (using defaults)
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
