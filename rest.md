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

### EXAMPLE
```bash
curl -X GET http://localhost:8000/servers/5
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
    "images":["ethereum:latest"],
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

### EXAMPLE
```bash
curl -X DELETE http://localhost:8000/testnets/2
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

### EXAMPLE
```bash
curl -X GET http://localhost:8000/testnets/2/nodes/
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

### EXAMPLE
```bash
curl -X POST http://localhost:8000/nodes/9e09efe8_d7a3_4429_832c_447d876194c8
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

### RESPONSE
```
Success
```

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

### RESPONSE
```
Success
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

### RESPONSE
```
Success
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

### RESPONSE
<!-- 
TODO: add dummy values
-->

### EXAMPLE
```bash
curl -X GET http://localhost:8000/resources/geth/genesis.json
```


## GET /build
Gets the details of the latest build
<!-- 
TODO: add dummy values
-->
### RESPONSE
```json
{
  "servers": [],
  "blockchain": "",
  "nodes": 0,
  "images": [], 
  "params": [],
  "resources": [],
  "environments": [],
  "files": [],
  "logs": [],
  "extras": [] 
}
```

### EXAMPLE
```bash
curl -X GET http://localhost:8000/build
```

## GET /build/{id}
Gets the details of the given build

<!-- 
TODO: add dummy values
-->

### RESPONSE
```json
{
  "servers": [],
  "blockchain": "",
  "nodes": 0,
  "images": [], 
  "params": [],
  "resources": [],
  "environments": [],
  "files": [],
  "logs": [],
  "extras": [] 
}
```

### EXAMPLE
```bash
curl -X GET http://localhost:8000/build/5
```

## POST /build/freeze/{id}
Pause the given build

### RESPONSE
```
Build has been frozen
```

### EXAMPLE
```bash
curl -X POST http://localhost:8000/build/freeze/5
```

## POST /build/thaw/{id}
Unpause the given build

### RESPONSE
```
Build has been resumed
```

### EXAMPLE
```bash
curl -X POST http://localhost:8000/build/thaw/5
```

## DELETE /build/freeze/{id}
Unpause the given build

### RESPONSE
```
Build has been resumed
```

### EXAMPLE
```bash
curl -X DELETE http://localhost:8000/build/freeze/5
```

## GET /emulate/{testnetID}
Get the current network conditions for the testnet

### RESPONSE 
```json
[
  {
    "node": 0,
    "limit": 0,
    "loss": 0.0,
    "delay": 0, 
    "rate": "",
    "duplicate": 0.0,
    "corrupt": 0.0,
    "reorder": 0.0
  }
]
```

### EXAMPLE
```bash
curl -X GET http://localhost:8000/emulate/5
```

## POST /nodes/restart/{testnetID}/{num}
Restart a node on a testnet

### RESPONSE
```
Success
```

### EXAMPLE
```bash
curl -X POST http://localhost:8000/nodes/restart/8c80891a-2046-4e4a-a3ca-652a38cb8093/5
```

## POST /nodes/raise/{testnetID}/{node}/{signal}
Send a signal to the main process of the given node 

### RESPONSE
```
Sent signal SIGINT to node 1
```

### EXAMPLE
```bash
curl -X POST http://localhost:8000/nodes/raise/8c80891a-2046-4e4a-a3ca-652a38cb8093/1/SIGINT
```

## POST /nodes/kill/{testnetID}/{node}
Attempt to kill the given node

### RESPONSE
```
Killed node 1
```

### EXAMPLE
```bash
curl -X POST http://localhost:8000/nodes/kill/8c80891a-2046-4e4a-a3ca-652a38cb8093/1
```

## POST /outage/{testnetID}/{node1}/{node2}
Prevent the given node1 and node2 from establishing a connection with each other

### RESPONSE
```
Success
```

### EXAMPLE
```bash
curl -X POST http://localhost:8000/outage/8c80891a-2046-4e4a-a3ca-652a38cb8093/1/2
```

## DELETE /outage/{testnetID}/{node1}/{node2}
Allow the given node1 and node2 to establish a connection with each other

### RESPONSE
```
Success
```

### EXAMPLE
```bash
curl -X DELETE http://localhost:8000/outage/8c80891a-2046-4e4a-a3ca-652a38cb8093/1/2
```

## DELETE /outage/{testnetID}
Remove all blocked connections from a testnet

### RESPONSE
```
Success
```

### EXAMPLE
```bash
curl -X DELETE http://localhost:8000/outage/8c80891a-2046-4e4a-a3ca-652a38cb8093
```

## GET /outage/{testnetID}
Get the currently blocked connections

### RESPONSE
```json
[
  {
    "to": 0,
    "from": 0
  }
]
```

### EXAMPLE
```bash
curl -X GET http://localhost:8000/outage/8c80891a-2046-4e4a-a3ca-652a38cb8093
```

## GET /outage/{testnetID}/{node}
Get the blocked connections for the given node

### RESPONSE
```json
[
  {
    "to": 0,
    "from": 0
  }
]
```

### EXAMPLE
```bash
curl -X GET http://localhost:8000/outage/8c80891a-2046-4e4a-a3ca-652a38cb8093/1
```

## POST /partition/{testnetID}
Create a network partition on a testnet

### RESPONSE
```
success
```

### EXAMPLE
```bash
curl -X POST http://localhost:8000/partition/8c80891a-2046-4e4a-a3ca-652a38cb8093
```

## GET /partition/{testnetID}
Get the partitions on the testnet

### RESPONSE
```
[
    [1, 2, 3],
    [4, 5, 6]
]
```

### EXAMPLE 
```bash
curl -X GET http://localhost:8000/partition/8c80891a-2046-4e4a-a3ca-652a38cb8093
```

## GET /blockchains
Get the currently supported blockchains by genesis

### RESPONSE
```
[
    "EOS", "Geth", "RChain"
]
```

### EXAMPLE
```bash
curl -X GET http://localhost:8000/blockchains
```

