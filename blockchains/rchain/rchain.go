package rchain

import(
    //"context"
    "fmt"
    "log"
    "time"
    "regexp"
    util "../../util"
    db "../../db"
)

var conf *util.Config

func init(){
    conf = util.GetConfig()
}


func Build(data map[string]interface{},nodes int,servers []db.Server,clients []*util.SshClient) ([]string,error) {
    util.Rm("./rchain.conf")
    rchainConf,err := NewRChainConf(data)
    if err != nil{
        log.Println(err)
        return nil,err
    }

    services,err := util.GetServiceIps(GetServices())
    if err != nil{
        log.Println(err)
        return nil,err
    }

    defer func(){
        util.Rm("./rchain.conf")
    }()

    /**Make the data directories**/
    for i,server := range servers {
        for j,_ := range server.Ips{
            clients[i].DockerExec(j,"mkdir /datadir")
        }
    }
    /**Setup the first node**/
    err = createFirstConfigFile(servers[0],0,rchainConf,services["wb_influx_proxy"])
    if err != nil{
        log.Println(err)
        return nil,err
    }
    
    km,err := NewKeyMaster()
    keyPairs := make([]util.KeyPair,nodes)

    for i,_ := range keyPairs {
        keyPairs[i],err = km.GetKeyPair()
        if err != nil{
            log.Println(err)
            return nil,err
        }
    }
    /**Setup bonds**/
    {
        bonds := make([]string,nodes)
        for i,keyPair := range keyPairs{
            bonds[i] = fmt.Sprintf("%s 1000000",keyPair.PublicKey)
        }
        err = util.Write("./bonds.txt",util.CombineConfig(bonds))
        if err != nil{
            log.Println(err)
            return nil,err
        }
        defer util.Rm("./bonds.txt")

        err = clients[0].Scp("./bonds.txt","/home/appo/bonds.txt")
        if err != nil{
            log.Println(err)
            return nil,err
        }
        defer clients[0].Run("rm -f /home/appo/bonds.txt")
        
        _,err = clients[0].Run("docker cp /home/appo/bonds.txt whiteblock-node0:/bonds.txt")
        if err != nil{
            log.Println(err)
            return nil,err
        }
        
    }

    var enode string
    {
        err = clients[0].DockerExecdLog(0,
            fmt.Sprintf("%s run --standalone --data-dir \"/datadir\" --host %s --bonds-file /bonds.txt --has-faucet",
                rchainConf.Command,servers[0].Ips[0]))
        if err != nil{
            log.Println(err)
            return nil,err
        }
        fmt.Println("Attempting to get the enode address")
        
        for i := 0; i < 1000; i++ {
            fmt.Println("Checking if the boot node is ready...")
            time.Sleep(time.Duration(1 * time.Second))
            output,err := clients[0].DockerExec(0,fmt.Sprintf("cat %s",conf.DockerOutputFile))
            if err != nil{
                log.Println(err)
                return nil,err
            }
            re := regexp.MustCompile(`(?m)rnode:\/\/[a-z|0-9]*\@([0-9]{1,3}\.){3}[0-9]{1,3}\?protocol=[0-9]*\&discovery=[0-9]*`)
            
            if !re.MatchString(output){
                fmt.Println("Not ready")
                continue
            }
            enode = re.FindAllString(output,1)[0]
            fmt.Println("Ready")
            break
        }

        
        log.Println("Got the address for the bootnode: "+enode)
        err = createConfigFile(enode,rchainConf,services["wb_influx_proxy"])
        if err != nil{
            log.Println(err)
            return nil,err
        }
    }
    defer util.Rm("./rnode.toml")
    /**Copy config files to the rest of the nodes**/
    for i,server := range servers{
        err = clients[i].Scp("./rnode.toml","/home/appo/rnode.toml")
        if err != nil{
            log.Println(err)
            return nil,err
        }
        for node,_ := range server.Ips{
            if node == 0 && i == 0 {
                continue
            }
            _,err = clients[i].Run(fmt.Sprintf("docker cp /home/appo/rnode.toml whiteblock-node%d:/rnode.toml",node))
            if err != nil{
                log.Println(err)
                return nil,err
            }
            
        }
        _,err = clients[i].Run("rm -f ~/rnode.toml")
        if err != nil{
            log.Println(err)
            return nil,err
        }
    }
    
    if err != nil{
        log.Println(err)
        return nil,err
    }
    fmt.Println("Starting the rest of the nodes...");
    /**Start up the rest of the nodes**/
    node := 0
    var validators int64 = 0
    for i,server := range servers {
        for j,ip := range server.Ips{
            if node == 0 {
                node++;
                continue
            }
            if validators < rchainConf.ValidatorCount {
                err = clients[i].DockerExecdLog(j,
                    fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --validator-private-key %s --host %s",
                                rchainConf.Command,enode,keyPairs[node].PrivateKey,ip))
                validators++
            }else{
                err = clients[i].DockerExecdLog(j,
                    fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --host %s",
                                rchainConf.Command,enode,ip))
            }
            
            if err != nil{
                log.Println(err)
                return nil,err
            }
            node++;
        }
    }
    /*err = SetupPrometheus(servers,clients)
    if err != nil{
        log.Println(err)
        return nil,err
    }*/
    return nil,nil
}


func createFirstConfigFile(server db.Server,node int,rchainConf *RChainConf,influxIP string) error {
    data := util.CombineConfig([]string{
        "[server]",
        "host = \"0.0.0.0\"",
        "port = 40500",
        "http-port = 40502",
        "metrics-port = 40503",
        fmt.Sprintf("no-upnp = %v",rchainConf.NoUpnp),
        fmt.Sprintf("default-timeout = %d",rchainConf.DefaultTimeout),
        "standalone = true",
        //"data-dir = \"/datadir\"",
        fmt.Sprintf("map-size = %d",rchainConf.MapSize),
        fmt.Sprintf("casper-block-store-size = %d",rchainConf.CasperBlockStoreSize),
        fmt.Sprintf("in-memory-store = %v",rchainConf.InMemoryStore),
        fmt.Sprintf("max-num-of-connections = %d",rchainConf.MaxNumOfConnections),
        "\n[grpc-server]",
        "host = \"0.0.0.0\"",
        "port = 40501",
        "port-internal = 40504",
        "\n[tls]",
        "#certificate = \"/var/lib/rnode/certificate.pem\"",
        "#key = \"/var/lib/rnode/key.pem\"",
        "\n[validators]",
        fmt.Sprintf("count = %d",rchainConf.ValidatorCount),
        "shard-id = \"wbtest\"",
        fmt.Sprintf("sig-algorithm = \"%s\"",rchainConf.SigAlgorithm),
        "bonds-file = \"/root/.rnode/genesis\"",
        "private-key = \"7fa626af8e4b96797888e6fc6884ce7c278c360170b13e4ce4000090c6f2bab\"",
        "\n[kamon]",
        "prometheus = false",
        "influx-db = true",
        "\n[influx-db]",
        fmt.Sprintf("hostname = \"%s\"",influxIP),
        "port = 8086",
        "database = \"rnode\"",

    });

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

func createConfigFile(bootnodeAddr string,rchainConf *RChainConf,influxIP string) error {
    data := util.CombineConfig([]string{
        "\n[server]",
        "host = \"0.0.0.0\"",
        "port = 40500",
        "http-port = 40502",
        "metrics-port = 40503",
        fmt.Sprintf("no-upnp = %v",rchainConf.NoUpnp),
        fmt.Sprintf("default-timeout = %d",rchainConf.DefaultTimeout),
        fmt.Sprintf("bootstrap = \"%s\"",bootnodeAddr),
        "standalone = false",
        //"data-dir = \"/datadir\"",
        fmt.Sprintf("map-size = %d",rchainConf.MapSize),
        fmt.Sprintf("casper-block-store-size = %d",rchainConf.CasperBlockStoreSize),
        fmt.Sprintf("in-memory-store = %v",rchainConf.InMemoryStore),
        fmt.Sprintf("max-num-of-connections = %d",rchainConf.MaxNumOfConnections),
        "\n[grpc-server]",
        "host = \"0.0.0.0\"",
        "port = 40501",
        "port-internal = 40504",
        "\n[tls]",
        "#certificate = \"/var/lib/rnode/certificate.pem\"",
        "#key = \"/var/lib/rnode/key.pem\"",
        "\n[validators]",
        fmt.Sprintf("count = %d",rchainConf.ValidatorCount),
        "shard-id = \"wbtest\"",
        fmt.Sprintf("sig-algorithm = \"%s\"",rchainConf.SigAlgorithm),
        "bonds-file = \"/root/.rnode/genesis\"",
        "private-key = \"7fa626af8e4b96797888e6fc6884ce7c278c360170b13e4ce4000090c6f2bab\"",
        "\n[kamon]",
        "prometheus = false",
        "influx-db = true",
        "\n[influx-db]",
        fmt.Sprintf("hostname = \"%s\"",influxIP),
        "port = 8086",
        "database = \"rnode\"",
    });

    return util.Write("./rnode.toml",data)  
}
