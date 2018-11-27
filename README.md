
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
	"image":(string)
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
curl -X POST http://localhost:8000/testnets/ -d '{"servers":[3],"blockchain":"ethereum","nodes":3,"image":"ethereum:latest"}'
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


##Configuration

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
