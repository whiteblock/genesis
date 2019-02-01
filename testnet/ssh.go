package testnet


import(
    "log"
    util "../util"
    db "../db"
)



var _clients = map[int]*util.SshClient{}


func GetClient(id int)(*util.SshClient,error) {
    cli,ok := _clients[id]
    if !ok || cli == nil{
        server,_,err := db.GetServer(id)
        if err != nil {
            log.Println(err)
            return nil,err
        }
        cli,err = util.NewSshClient(server.Addr)
        if err != nil {
            log.Println(err)
            return nil,err
        }
        _clients[id] = cli
    }
    return cli,nil
}

func GetClients(servers []int) ([]*util.SshClient,error) {

    out := make([]*util.SshClient,len(servers))
    var err error
    for i := 0; i < len(servers); i++ {
        out[i],err = GetClient(servers[i])
        if err != nil {
            log.Println(err)
            return nil,err
        }
    }
    return out,nil
}
    