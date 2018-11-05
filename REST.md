

__Warning: Do not use any path marked PRIVATE, they will begin to require credentials in the near future__

###GET /servers/
Get the current registered servers
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
{
	"server_name":{
		"Addr":(string),
		"Iaddr":{
			"Ip":(string),
			"Gateway":(string),
			"Subnet":(int)
		},
		"Nodes":(int),
		"Max":(int),
		"Id":(int),
		"ServerID":(int),
		"Iface":(string),
		"Switches":[
			{
				"Addr":(string),
				"Iface":(string),
				"Brand":(int),
				"Id":(int)
			}
		]
	},
	"server2_name":{...}...
}
```


###PUT /servers/{name}
####PRIVATE
Register and add a new server to be 
controlled by the instance
#####BODY
```
{
	"Addr":(string),
	"Iaddr":{
		"Ip":(string),
		"Gateway":(string),
		"Subnet":(int)
	},
	"Nodes":(int),
	"Max":(int),
	"Id":-1,
	"ServerID":(int),
	"Iface":(string),
	"Switches":[
		{
			"Addr":(string),
			"Iface":(string),
			"Brand":(int),
			"Id":(int)
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

###GET /servers/{id}
Get a server by id
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
{
	"Addr":(string),
	"Iaddr":{
		"Ip":(string),
		"Gateway":(string),
		"Subnet":(int)
	},
	"Nodes":(int),
	"Max":(int),
	"Id":(int),
	"ServerID":(int),
	"Iface":(string),
	"Switches":[
		{
			"Addr":(string),
			"Iface":(string),
			"Brand":(int),
			"Id":(int)
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

###UPDATE /servers/{id}
####PRIVATE
Update server information
#####BODY
```
{
	"Addr":(string),
	"Iaddr":{
		"Ip":(string),
		"Gateway":(string),
		"Subnet":(int)
	},
	"Nodes":(int),
	"Max":(int),
	"Id":(int),
	"ServerID":(int),
	"Iface":(string),
	"Switches":[
		{
			"Addr":(string),
			"Iface":(string),
			"Brand":(int),
			"Id":(int)
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

###GET /testnets/
Get all testnets which are currently running
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
[
	{
		"Id":(int),
		"Blockchain":(string),
		"Nodes":(int),
		"Image":(string)
	},...

]
```

###POST /testnets/
Add and deploy a new testnet
#####BODY
```
{
	"Servers":[(int),(int)...],
	"Blockchain":(string),
	"Nodes":(int),
	"Image":(string)
}
```
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
Success
```

###GET /testnets/{id}
Get data on a single testnet
#####RESPONSE
```
HTTP/1.1 200 OK
Date: Mon, 22 Oct 2018 15:31:18 GMT
{
	"Id":(int),
	"Blockchain":(string),
	"Nodes":(int),
	"Image":(string)
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
	"Id":(int),
	"TestNetId":(int),
	"Server":(int),
	"LocalId":(int),
	"Ip":(string)
]
```

