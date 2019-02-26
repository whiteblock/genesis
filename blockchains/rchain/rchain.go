package rchain

import(
    "fmt"
    "log"
    "time"
    "regexp"
    "io/ioutil"
    "github.com/Whiteblock/mustache"
    util "../../util"
    db "../../db"
    state "../../state"
)

var conf *util.Config

func init(){
    conf = util.GetConfig()
}


func Build(data map[string]interface{},nodes int,servers []db.Server,clients []*util.SshClient,buildState *state.BuildState) ([]string,error) {

    util.Rm("./rchain.conf")
    rchainConf,err := NewRChainConf(data)
    if err != nil {
        log.Println(err)
        return nil,err
    }
    buildState.SetBuildSteps(9+(len(servers)*2)+(nodes*3))
    buildState.SetBuildStage("Setting up data collection")

    services,err := util.GetServiceIps(GetServices())
    buildState.IncrementBuildProgress()
    if err != nil {
        log.Println(err)
        return nil,err
    }

    defer func(){
        util.Rm("./rchain.conf")
    }()

    /**Make the data directories**/
    for i,server := range servers {
        for j,_ := range server.Ips {
            buildState.IncrementBuildProgress()
            clients[i].DockerExec(j,"mkdir /datadir")
        }
    }
    /**Setup the first node**/
    err = createFirstConfigFile(clients[0],0,rchainConf,services["wb_influx_proxy"])
    if err != nil{
        log.Println(err)
        return nil,err
    }
    buildState.IncrementBuildProgress()
    km,err := NewKeyMaster()
    keyPairs := make([]util.KeyPair,nodes)

    for i,_ := range keyPairs {
        keyPairs[i],err = km.GetKeyPair()
        if err != nil{
            log.Println(err)
            return nil,err
        }
    }
    buildState.IncrementBuildProgress()

    buildState.SetBuildStage("Setting up bonds")
    /**Setup bonds**/
    {
        bonds := make([]string,nodes)
        for i,keyPair := range keyPairs{
            bonds[i] = fmt.Sprintf("%s 1000000",keyPair.PublicKey)
        }
        buildState.IncrementBuildProgress()
        err = util.Write("./bonds.txt",util.CombineConfig(bonds))
        if err != nil{
            log.Println(err)
            return nil,err
        }
        buildState.IncrementBuildProgress()
        defer util.Rm("./bonds.txt")

        err = clients[0].Scp("./bonds.txt","/home/appo/bonds.txt")
        if err != nil{
            log.Println(err)
            return nil,err
        }
        buildState.IncrementBuildProgress()
        defer clients[0].Run("rm -f /home/appo/bonds.txt")
        
        _,err = clients[0].Run("docker cp /home/appo/bonds.txt whiteblock-node0:/bonds.txt")
        if err != nil{
            log.Println(err)
            return nil,err
        }
        buildState.IncrementBuildProgress()
        
    }

    buildState.SetBuildStage("Starting the boot node")
    var enode string
    {
        err = clients[0].DockerExecdLog(0,
            fmt.Sprintf("%s run --standalone --data-dir \"/datadir\" --host %s --bonds-file /bonds.txt --has-faucet",
                rchainConf.Command,servers[0].Ips[0]))
        buildState.IncrementBuildProgress()
        if err != nil{
            log.Println(err)
            return nil,err
        }
        //fmt.Println("Attempting to get the enode address")
        buildState.SetBuildStage("Waiting for the boot node's address")
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
        buildState.IncrementBuildProgress()
        /*
            influxIp
            validators
         */
        log.Println("Got the address for the bootnode: "+enode)
        err = createConfigFile(enode,rchainConf,services["wb_influx_proxy"])
        if err != nil{
            log.Println(err)
            return nil,err
        }
    }

    buildState.SetBuildStage("Configuring the other rchain nodes")
    defer util.Rm("./rnode.conf")
    /**Copy config files to the rest of the nodes**/
    for i,server := range servers{
        err = clients[i].Scp("./rnode.conf","/home/appo/rnode.conf")
        if err != nil{
            log.Println(err)
            return nil,err
        }
        buildState.IncrementBuildProgress()
        for node,_ := range server.Ips {
            if node == 0 && i == 0 {
                continue
            }
            _,err = clients[i].Run(fmt.Sprintf("docker cp /home/appo/rnode.conf whiteblock-node%d:/datadir/rnode.conf",node))
            if err != nil{
                log.Println(err)
                return nil,err
            }
            buildState.IncrementBuildProgress()
            
        }
        _,err = clients[i].Run("rm -f ~/rnode.conf")
        if err != nil{
            log.Println(err)
            return nil,err
        }
        buildState.IncrementBuildProgress()
    }
    
    if err != nil{
        log.Println(err)
        return nil,err
    }
    buildState.SetBuildStage("Starting the rest of the nodes")
    /**Start up the rest of the nodes**/
    node := 0
    var validators int64 = 0
    for i,server := range servers {
        for j,ip := range server.Ips {
            if node == 0 {
                node++;
                continue
            }
            if validators < rchainConf.Validators {
                err = clients[i].DockerExecdLog(j,
                    fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --validator-private-key %s --host %s",
                                rchainConf.Command,enode,keyPairs[node].PrivateKey,ip))
                validators++
            }else{
                err = clients[i].DockerExecdLog(j,
                    fmt.Sprintf("%s run --data-dir \"/datadir\" --bootstrap \"%s\" --host %s",
                                rchainConf.Command,enode,ip))
            }
            buildState.IncrementBuildProgress()
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


func createFirstConfigFile(client *util.SshClient,node int,rchainConf *RChainConf,influxIP string) error {
    filler := util.ConvertToStringMap(map[string]interface{}{
        "influxIp":influxIP,
        "validatorCount":rchainConf.ValidatorCount,
        "standalone":false,
    })
    dat, err := ioutil.ReadFile("./resources/rchain/rchain.conf.mustache")
    if err != nil {
        return err
    }
    data, err := mustache.Render(string(dat), filler)

    err = util.Write("./rnode.conf",data)
    if err != nil{
        log.Println(err)
        return err
    }
    err = client.Scp("./rnode.conf","/home/appo/rnode.conf")
    if err != nil{
        log.Println(err)
        return err
    }
    _,err = client.Run(fmt.Sprintf("docker cp /home/appo/rnode.conf whiteblock-node%d:/datadir/rnode.conf",node))
    if err != nil{
        log.Println(err)
        return err
    }
    _,err = client.Run("rm -f ~/rnode.conf")
    if err != nil{
        log.Println(err)
        return err
    }
    return util.Rm("./rnode.conf")
}

func Add(data map[string]interface{},nodes int,servers []db.Server,clients []*util.SshClient,
         newNodes map[int][]string,buildState *state.BuildState) ([]string,error) {
    return nil,nil
}

func createConfigFile(bootnodeAddr string,rchainConf *RChainConf,influxIP string) error {
    filler := util.ConvertToStringMap(map[string]interface{}{
        "influxIp":influxIP,
        "validatorCount":rchainConf.ValidatorCount,
        "standalone":false,
        "bootstrap":fmt.Sprintf("bootstrap = \"%s\"",bootnodeAddr),
    })
    dat, err := ioutil.ReadFile("./resources/rchain/rchain.conf.mustache")
    if err != nil {
        return err
    }
    data, err := mustache.Render(string(dat), filler)
    return util.Write("./rnode.conf",data)  
}
