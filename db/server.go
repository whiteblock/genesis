package db

import(
    _ "github.com/mattn/go-sqlite3"
    "fmt"
    "regexp"
    "errors"
    "log"
)

type Server struct{
    Addr        string      `json:"addr"`//IP to access the server
    Iaddr       Iface       `json:"iaddr"`//Internal IP of the server for NIC attached to the vyos
    Nodes       int         `json:"nodes"`
    Max         int         `json:"max"`
    Id          int         `json:"id"`
    ServerID    int         `json:"serverID"`
    Iface       string      `json:"iface"`
    Switches    []Switch    `json:"switches"`
    Ips         []string    `json:"ips"`
}

type Iface struct {
    Ip          string      `json:"ip"`
    Gateway     string      `json:"gateway"`
    Subnet      int         `json:"subnet"`
}

func (s Server) Validate() error {
    var re = regexp.MustCompile(`(?m)[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`)
    if !re.Match([]byte(s.Addr)) {
        return errors.New("Addr is invalid")
    }
    if s.Nodes < 0 {
        return errors.New("Nodes is invalid")
    }
    if s.Nodes > s.Max {
        return errors.New("Max is invalid")
    }
    if s.ServerID < 1 {
        return errors.New("ServerID is invalid")
    }
    if len(s.Switches) != 0 {
        err := s.Switches[0].Validate()
        if err != nil {
            return err
        }
    }
    
    return nil
}

func GetAllServers() (map[string]Server,error) {

    rows, err :=  db.Query(fmt.Sprintf("SELECT id,server_id,addr, iaddr_ip,iaddr_gateway,iaddr_subnet,nodes,max,iface,switch,name FROM %s",ServerTable ))
    if err != nil {
        return nil,err 
    }
    defer rows.Close()
    allServers := make(map[string]Server)
    for rows.Next() {
        var name string
        var server Server
        var switchId int
        //var subnet string
        err := rows.Scan(&server.Id,&server.ServerID,&server.Addr,
                                  &server.Iaddr.Ip, &server.Iaddr.Gateway, &server.Iaddr.Subnet,
                                  &server.Nodes,&server.Max,&server.Iface,&switchId,&name)
        if err != nil {
            return nil,err
        }
        if switchId != -1 {
            swtch, err := GetSwitchById(switchId)
            if err != nil {
                return nil,err
            }
            server.Switches = []Switch{ swtch }
        }else{
            server.Switches = nil
        }
        

        
        allServers[name] = server
    }
    return allServers,nil
}

func GetServers(ids []int) ([]Server,error) {
    var servers []Server
    for _, id := range ids {
        server,_,err := GetServer(id)
        if err != nil {
            return servers, err
        }
        servers = append(servers,server)
    }
    return servers,nil
}

func GetServer(id int) (Server, string, error) {
    var name string
    var server Server

    rows, err :=  db.Query(fmt.Sprintf("SELECT id,server_id,addr,iaddr_ip,iaddr_gateway,iaddr_subnet,nodes,max,iface,switch,name FROM %s WHERE id = %d",
        ServerTable,id ))
    if err != nil {
        return server,name,err
    }

    
    
    if !rows.Next() {
        return server, name, errors.New("Not found")
    }
    defer rows.Close()
    var switchId int
    //var subnet string
    err = rows.Scan(&server.Id,&server.ServerID,&server.Addr,
                              &server.Iaddr.Ip, &server.Iaddr.Gateway, &server.Iaddr.Subnet,
                              &server.Nodes,&server.Max,&server.Iface,&switchId,&name)
    if err != nil{
        log.Println(err)
        return server,name,err
    }
    //fmt.Println(subnet)
    //fmt.Printf("Switch id is %d\n",switchId)
    if switchId != -1 {
        swtch, err := GetSwitchById(switchId)
        if err != nil {
            return server,name,err
        }
        server.Switches = []Switch{ swtch }
    }else{
        server.Switches = nil
    }
    

    return server,name,nil
}

func _insertServer(name string,server Server,switchId int) (int,error) {

    tx,err := db.Begin()
    if err != nil {
        return -1,err
    }

    stmt,err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (addr,server_id,iaddr_ip,iaddr_gateway,iaddr_subnet,nodes,max,iface,switch,name) VALUES (?,?,?,?,?,?,?,?,?,?)",ServerTable))
    if err != nil {
        return -1,err
    }

    defer stmt.Close()

    res,err := stmt.Exec(server.Addr,server.ServerID,server.Iaddr.Ip, server.Iaddr.Gateway, server.Iaddr.Subnet,
                       server.Nodes,server.Max,server.Iface,switchId,name)
    if err != nil {
        return -1,err
    }
    tx.Commit()
    id, err := res.LastInsertId()
    return int(id),err
}

func InsertServer(name string,server Server) (int,error) {
    if len(server.Switches) == 0 {
        return _insertServer(name,server,-1)
    }
    switchId,err := InsertSwitch(server.Switches[0])
    if err != nil {
        return -1,err
    }
    return _insertServer(name,server,switchId)
}

func DeleteServer(id int) error {

    _,err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = %d",ServerTable,id))
    return err
}

func UpdateServer(id int,server Server) error {
    //Handle Updating of Switch
    //fmt.Printf("UPDATED server is %+v\n",server)
    swtch,err := GetSwitchById(server.Switches[0].Id)

    var switchId int
    if server.Switches == nil || len(server.Switches) == 0 {
        switchId = -1
    }else if err != nil {
        switchId,err = InsertSwitch(server.Switches[0])
        if err != nil {
            return err
        }
    }else{
        switchId = swtch.Id
    }

    db := getDB()
    defer db.Close()

    tx,err := db.Begin()
    if err != nil {
        return err
    }

    stmt,err := tx.Prepare(fmt.Sprintf("UPDATE %s SET server_id = ?,addr = ?, iaddr_ip = ?, iaddr_gateway = ?, iaddr_subnet = ?, nodes = ?, max = ?, iface = ?, switch = ? WHERE id = ? ",ServerTable))
    if err != nil {
        return err
    }
    defer stmt.Close()

    _,err = stmt.Exec(server.ServerID,
                      server.Addr,
                      server.Iaddr.Ip,
                      server.Iaddr.Gateway,
                      server.Iaddr.Subnet,
                      server.Nodes,
                      server.Max,
                      server.Iface,
                      switchId,
                      server.Id)
    if err != nil {
        return err
    }
    return tx.Commit()
}

func UpdateServerNodes(id int,nodes int) error {

    db := getDB()
    defer db.Close()

    tx,err := db.Begin()
    if err != nil {
        return err
    }

    stmt,err := tx.Prepare(fmt.Sprintf("UPDATE %s SET nodes = ? WHERE id = ?",ServerTable))

    if err != nil {
        return err
    }
    defer stmt.Close()

    _,err = stmt.Exec(nodes,id)
    if err != nil {
        return err
    }
    return tx.Commit()

}

func GetHostIPsByTestNet(id int) ([]string,error) {

    rows, err :=  db.Query(fmt.Sprintf("SELECT id,server_id,addr,iaddr_ip,iaddr_gateway,iaddr_subnet,nodes,max,iface,switch,name FROM %s INNER JOIN %s ON %s.id == %s.server WHERE %s.id == %d GROUP BY %s.id",
        ServerTable,
        NodesTable,
        ServerTable,
        NodesTable,
        ServerTable,
        id,
        ServerTable))

    ips := []string{}

    if err != nil {
        return ips, err
    }

    defer rows.Close()

    for rows.Next() {
        var ip string
        err = rows.Scan(&ip)
        if err != nil {
            return ips, err
        }

        ips = append(ips,ip)
    }
    return ips, nil
}

func GetServersByTestNet(id int) ([]Server,error) {

    rows, err :=  db.Query(fmt.Sprintf("SELECT ip FROM %s INNER JOIN %s ON %s.id == %s.server WHERE %s.id == %d GROUP BY %s.id",
        ServerTable,
        NodesTable,
        ServerTable,
        NodesTable,
        ServerTable,
        id,
        ServerTable))
    
    servers := []Server{}

    if err != nil {
        return nil, err
    }

    defer rows.Close()

    for rows.Next() {
        var name string
        var server Server
        var switchId int

        err = rows.Scan(&server.Id,&server.Addr,&server.Iaddr.Ip,&server.Iaddr.Gateway,&server.Iaddr.Subnet,
                             &server.Nodes,&server.Max,&switchId,&name)
        if err != nil {
            return nil, err
        }
        swtch, err := GetSwitchById(switchId)
        
        if err != nil {
            return servers, err
        }

        server.Switches = []Switch{ swtch }

        servers = append(servers,server)
    }
    return servers, nil
}