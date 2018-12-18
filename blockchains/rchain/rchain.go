package rchain

import(
	//"context"
	"fmt"
	//"errors"
	"log"
	"time"
	"regexp"
	util "../../util"
	db "../../db"
)



func Build(data map[string]interface{},nodes int,servers []db.Server) ([]string,error) {
	util.Rm("./rchain.conf")
	cmd := "/rchain/node/target/rnode-0.8.1/usr/share/rnode/bin/rnode"
	defer func(){
		util.Rm("./rchain.conf")
	}()

	/**Make the data directories**/
	for _,server := range servers {
		for i,_ := range server.Ips{
			util.DockerExec(server.Addr,i,"bash -c 'mkdir /datadir'")
		}
	}
	/**Setup the first node**/
	err := createFirstConfigFile(servers[0],0)
	if err != nil{
		log.Println(err)
		return nil,err
	}
	var enode string
	{
		_,err = util.DockerExecd(servers[0].Addr,0,fmt.Sprintf("bash -c '%s --config-file /rchain.toml run --standalone --data-dir \"/datadir\" --host 0.0.0.0>> /datadir/rchain.stdout'",cmd))
		if err != nil{
			log.Println(err)
			return nil,err
		}
		println("Attempting to get the enode address")
		
		for i := 0; i < 200; i++ {
			println("Checking if the boot node is ready...")
			time.Sleep(time.Duration(1 * time.Second))
			output,err := util.DockerExec(servers[0].Addr,0,fmt.Sprintf("cat /datadir/rchain.stdout"))
			if err != nil{
				log.Println(err)
				return nil,err
			}
			re := regexp.MustCompile(`(?m)rnode:\/\/[a-z|0-9]*\@([0-9]{1,3}\.){3}[0-9]{1,3}\?protocol=[0-9]*\&discovery=[0-9]*`)
			
			if !re.MatchString(output){
				println("Not ready")
				continue
			}
			enode = re.FindAllString(output,1)[0]
			println("Ready")
			break
		}
		

		
		log.Println("Got the address for the bootnode: "+enode)
		err = createConfigFile(enode)
		if err != nil{
			log.Println(err)
			return nil,err
		}
	}
	/**Copy config files to the rest of the nodes**/
	for i,server := range servers{
		err = util.Scp(server.Addr,"./rnode.toml","/home/appo/rnode.toml")
		if err != nil{
			log.Println(err)
			return nil,err
		}
		for node,_ := range server.Ips{
			if node == 0 && i == 0 {
				continue
			}
			_,err = util.SshExec(server.Addr,fmt.Sprintf("docker cp /home/appo/rnode.toml whiteblock-node%d:/rnode.toml",node))
			if err != nil{
				log.Println(err)
				return nil,err
			}
			
		}
		_,err = util.SshExec(server.Addr,"rm -f ~/rnode.toml")
		if err != nil{
			log.Println(err)
			return nil,err
		}
	}
	err = util.Rm("./rnode.toml")
	if err != nil{
		log.Println(err)
		return nil,err
	}

	/**Start up the rest of the nodes**/
	for i,server := range servers{
		for node,_ := range server.Ips{
			if node == 0 && i == 0 {
				continue
			}
			_,err = util.DockerExecd(server.Addr,node,fmt.Sprintf("%s --config-file /rchain.toml run --data-dir \"/datadir\" --bootstrap \"%s\" --host 0.0.0.0",cmd,enode))
			if err != nil{
				log.Println(err)
				return nil,err
			}
		}
	}
	return nil,nil
}


func createFirstConfigFile(server db.Server,node int) error {
data := fmt.Sprintf(`[server]
host = "%s"
port = 40500
http-port = 40502
metrics-port = 40503
no-upnp = true
default-timeout = 2000
standalone = true
#data-dir = "/datadir"
map-size = 1073741824
casper-block-store-size = 1073741824
in-memory-store = false
max-num-of-connections = 500
[grpc-server]
host = "0.0.0.0"
port = 40501
port-internal = 40504
[tls]
#certificate = "/var/lib/rnode/certificate.pem"
#key = "/var/lib/rnode/key.pem"
[validators]
count = 5
shard-id = "wbtest"
sig-algorithm = "ed25519"
bonds-file = "/root/.rnode/genesis"
private-key = "7fa626af8e4b96797888e6fc6884ce7c278c360170b13e4ce4000090c6f2bab9"`,server.Ips[node])
	err := util.Write("./rnode.toml",data)
	if err != nil{
		log.Println(err)
		return err
	}
	err = util.Scp(server.Addr,"./rnode.toml","/home/appo/rnode.toml")
	if err != nil{
		log.Println(err)
		return err
	}
	_,err = util.SshExec(server.Addr,fmt.Sprintf("docker cp /home/appo/rnode.toml whiteblock-node%d:/rnode.toml",node))
	if err != nil{
		log.Println(err)
		return err
	}
	_,err = util.SshExec(server.Addr,"rm -f ~/rnode.toml")
	if err != nil{
		log.Println(err)
		return err
	}
	return util.Rm("./rnode.toml")
}

func createConfigFile(bootnodeAddr string) error {
data := fmt.Sprintf(`[server]
host = "0.0.0.0"
port = 40500
http-port = 40502
metrics-port = 40503
no-upnp = true
default-timeout = 2000
bootstrap = "%s"
standalone = false
#data-dir = "/datadir"
map-size = 1073741824
casper-block-store-size = 1073741824
in-memory-store = false
max-num-of-connections = 500
[grpc-server]
host = "0.0.0.0"
port = 40501
port-internal = 40504
[tls]
#certificate = "/var/lib/rnode/certificate.pem"
#key = "/var/lib/rnode/key.pem"
[validators]
count = 5
shard-id = "wbtest"
sig-algorithm = "ed25519"
bonds-file = "/root/.rnode/genesis"
private-key = "7fa626af8e4b96797888e6fc6884ce7c278c360170b13e4ce4000090c6f2bab9"`,bootnodeAddr)
	return util.Write("./rnode.toml",data)	
}
