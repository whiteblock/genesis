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
    Iaddr       Iface       `json:"iaddr"`//Internal IP of the server that the nodes can reach
    Nodes       int         `json:"nodes"`
    Max         int         `json:"max"`
    Id          int         `json:"id"`
    ServerID    int         `json:"serverID"`
    Iface       string      `json:"iface"`
    Ips         []string    `json:"ips"`
}

type Iface struct {
    Ip          string      `json:"ip"`
    Gateway     string      `json:"gateway"`
    Subnet      int         `json:"subnet"`
}

/*
    Ensure that a server object contains valid data
 */
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
    return nil
}

/*
    Get all of the servers, indexed by name
 */
func GetAllServers() (map[string]Server,error) {

    rows, err :=  db.Query(fmt.Sprintf("SELECT id,server_id,addr, iaddr_ip,iaddr_gateway,iaddr_subnet,nodes,max,iface,name FROM %s",ServerTable ))
    if err != nil {
        return nil,err 
    }
    defer rows.Close()
    allServers := make(map[string]Server)
    for rows.Next() {
        var name string
        var server Server
        //var subnet string
        err := rows.Scan(&server.Id,&server.ServerID,&server.Addr,
                                  &server.Iaddr.Ip, &server.Iaddr.Gateway, &server.Iaddr.Subnet,
                                  &server.Nodes,&server.Max,&server.Iface,&name)
        if err != nil {
            return nil,err
        }
        
        allServers[name] = server
    }
    return allServers,nil
}

/*
    Get servers from their ids
 */
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

/*
    Get a server by id
 */
func GetServer(id int) (Server, string, error) {
    var name string
    var server Server

    rows, err :=  db.Query(fmt.Sprintf("SELECT id,server_id,addr,iaddr_ip,iaddr_gateway,iaddr_subnet,nodes,max,iface,name FROM %s WHERE id = %d",
        ServerTable,id ))
    if err != nil {
        return server,name,err
    }

    
    
    if !rows.Next() {
        return server, name, errors.New("Not found")
    }
    defer rows.Close()
    err = rows.Scan(&server.Id,&server.ServerID,&server.Addr,
                              &server.Iaddr.Ip, &server.Iaddr.Gateway, &server.Iaddr.Subnet,
                              &server.Nodes,&server.Max,&server.Iface,&name)
    if err != nil{
        log.Println(err)
        return server,name,err
    }    

    return server,name,nil
}

/*
    Insert a new server into the database
 */
func InsertServer(name string,server Server) (int,error) {

    tx,err := db.Begin()
    if err != nil {
        return -1,err
    }

    stmt,err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (addr,server_id,iaddr_ip,iaddr_gateway,iaddr_subnet,nodes,max,iface,name) VALUES (?,?,?,?,?,?,?,?,?)",ServerTable))
    if err != nil {
        return -1,err
    }

    defer stmt.Close()

    res,err := stmt.Exec(server.Addr,server.ServerID,server.Iaddr.Ip, server.Iaddr.Gateway, server.Iaddr.Subnet,
                       server.Nodes,server.Max,server.Iface,name)
    if err != nil {
        return -1,err
    }
    tx.Commit()
    id, err := res.LastInsertId()
    return int(id),err
}

/*
    Delete a server by id
 */
func DeleteServer(id int) error {

    _,err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = %d",ServerTable,id))
    return err
}

/*
    Update a server by id
 */
func UpdateServer(id int,server Server) error {

    tx,err := db.Begin()
    if err != nil {
        return err
    }

    stmt,err := tx.Prepare(fmt.Sprintf("UPDATE %s SET server_id = ?,addr = ?, iaddr_ip = ?, iaddr_gateway = ?, iaddr_subnet = ?, nodes = ?, max = ?, iface = ? WHERE id = ? ",ServerTable))
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
                      server.Id)
    if err != nil {
        return err
    }
    return tx.Commit()
}

/*
    Update the number of nodes a server has
 */
func UpdateServerNodes(id int,nodes int) error {

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

/*
    Get the ips of the hosts for a testnet
 */
func GetHostIPsByTestNet(id int) ([]string,error) {

    rows, err :=  db.Query(fmt.Sprintf("SELECT id,server_id,addr,iaddr_ip,iaddr_gateway,iaddr_subnet,nodes,max,iface,name FROM %s INNER JOIN %s ON %s.id == %s.server WHERE %s.id == %d GROUP BY %s.id",
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

