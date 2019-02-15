package deploy

import (
    "fmt"
    "log"
    "io/ioutil"
    "strings"
    db "../db"
    util "../util"
    state "../state"
)

/**
 * Finalization methods for the docker build process. Will be run immediately following their deployment
 */
func finalize(servers []db.Server,clients []*util.SshClient,buildState *state.BuildState) error {
    if(conf.HandleNodeSshKeys){
        err := copyOverSshKeys(servers,clients,buildState)
        if err != nil {
            log.Println(err)
            return err
        }
    }
    

    return nil
}

/**
 * Finalization methods for the docker build process. Will be run immediately following their deployment
 */
func finalizeNewNodes(servers []db.Server,clients []*util.SshClient,newNodes map[int][]string,buildState *state.BuildState) error {
    if(conf.HandleNodeSshKeys){
        err := copyOverSshKeysToNewNodes(servers,clients,newNodes,buildState)
        if err != nil {
            log.Println(err)
            return err
        }
    }
    return nil
}


func copyOverSshKeys(servers []db.Server,clients []*util.SshClient,buildState *state.BuildState) error {
    tmp, err := ioutil.ReadFile(conf.NodesPublicKey)
    pubKey := string(tmp)
    pubKey = strings.Trim(pubKey,"\t\n\v\r")
    if err != nil{
        log.Println(err)
        return err
    }

    for i,server := range servers {
        clients[i].Run("rm /home/appo/node_key")
        err = clients[i].Scp(conf.NodesPrivateKey,"/home/appo/node_key")
        if err != nil{
            log.Println(err)
            return err
        }
        defer clients[i].Run("rm /home/appo/node_key")

        for j,_ := range server.Ips{
            res,err := clients[i].DockerExec(j,"mkdir -p /root/.ssh/")
            if err != nil{
                log.Println(res)
                log.Println(err)
                return err
            }

            res,err = clients[i].Run(fmt.Sprintf("docker cp /home/appo/node_key %s%d:/root/.ssh/id_rsa",
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
            buildState.IncrementDeployProgress()
        }
    }
    return nil
}



func copyOverSshKeysToNewNodes(servers []db.Server,clients []*util.SshClient,newNodes map[int][]string,buildState *state.BuildState) error {
    tmp, err := ioutil.ReadFile(conf.NodesPublicKey)
    pubKey := string(tmp)
    pubKey = strings.Trim(pubKey,"\t\n\v\r")
    if err != nil{
        log.Println(err)
        return err
    }

    for i,server := range servers {
        nodes := len(newNodes[server.Id])
        clients[i].Run("rm /home/appo/node_key")
        err = clients[i].Scp(conf.NodesPrivateKey,"/home/appo/node_key")
        if err != nil{
            log.Println(err)
            return err
        }
        defer clients[i].Run("rm /home/appo/node_key")

        for j := server.Nodes; j < server.Nodes + nodes; j++ {
            res,err := clients[i].DockerExec(j,"mkdir -p /root/.ssh/")
            if err != nil{
                log.Println(res)
                log.Println(err)
                return err
            }

            res,err = clients[i].Run(fmt.Sprintf("docker cp /home/appo/node_key %s%d:/root/.ssh/id_rsa",
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
            buildState.IncrementDeployProgress()
        }
    }
    return nil
}