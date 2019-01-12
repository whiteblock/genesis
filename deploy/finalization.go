package deploy

import (
    "fmt"
    "log"
    "io/ioutil"
    "strings"
    db "../db"
    util "../util"
    //state "../state"
)

/**
 * Finalization methods for the docker build process. Will be run immediately following their deployment
 */
func finalize(servers []db.Server,clients []*util.SshClient) error {
    if(conf.HandleNodeSshKeys){
        err := copyOverSshKeys(servers,clients)
        if err != nil {
            log.Println(err)
            return err
        }
    }
    

    return nil
}


func copyOverSshKeys(servers []db.Server,clients []*util.SshClient) error {
    tmp, err := ioutil.ReadFile(conf.NodesPublicKey)
    pubKey := string(tmp)
    pubKey = strings.Trim(pubKey,"\t\n\v\r")
    if err != nil{
        log.Println(err)
        return err
    }

    for i,server := range servers {
        err = clients[i].Scp(conf.NodesPrivateKey,"/home/appo/node_key")
        if err != nil{
            log.Println(err)
            return err
        }
        defer clients[i].Run("rm /home/appo/node_key")

        for j,_ := range server.Ips{
            _,err = clients[i].DockerExec(j,"mkdir -p /root/.ssh/")
            if err != nil{
                log.Println(err)
                return err
            }

            res,err := clients[i].Run(fmt.Sprintf("docker cp /home/appo/node_key %s%d:/root/.ssh/id_rsa",
                conf.NodePrefix,j))
            if err != nil {
                log.Println(res)
                log.Println(err)
                return err
            }

            res,err = clients[i].DockerExec(j,fmt.Sprintf(`bash -c 'echo "%s" >> /root/.ssh/authorized_keys'`,pubKey))
            if err != nil {
                log.Println(res)
                log.Println(err)
                return err
            }

            res,err = clients[i].DockerExecd(j,"service ssh start")
            if err != nil {
                log.Println(res)
                log.Println(err)
                return err
            }
        }
    }
    return nil
}