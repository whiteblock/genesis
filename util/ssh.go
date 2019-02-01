package util

import (
    "fmt"
    "strings"
    "io/ioutil"
    "golang.org/x/crypto/ssh"
    "log"
)


func _sshConnect(host string) (*ssh.Session,*ssh.Client, error) {
    client,err := sshConnect(host)
    if err != nil {
        log.Println(err)
        return nil, nil, err
    }
    session, err := client.NewSession()
    if err != nil {
        client.Close()
        log.Println(err)
        return nil, nil, err
    }

    return session,client, nil
}

func sshConnect(host string) (*ssh.Client, error) {
    sshConfig := &ssh.ClientConfig{
        User: conf.SshUser,
        Auth: []ssh.AuthMethod{ssh.Password(conf.SshPassword)},
    }
    sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

    client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", host), sshConfig)
    /*if err != nil {
        fmt.Println("First ssh attempt failed: " + err.Error())
    }*/
    if err != nil {//Try to connect using the id_rsa file
        key, err := ioutil.ReadFile(conf.RsaKey)
        if err != nil {
            return nil,err
        }
        signer, err := ssh.ParsePrivateKey(key)
        if err != nil {
            return nil, err
        }
        sshConfig = &ssh.ClientConfig{
            User: conf.RsaUser,
            Auth: []ssh.AuthMethod{
                // Use the PublicKeys method for remote authentication.
                ssh.PublicKeys(signer),
            },
        }
        sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
        client, err = ssh.Dial("tcp", fmt.Sprintf("%s:22", host), sshConfig)
        if err != nil{
            return nil,err
        }
    }


    return client, nil
}


/**
 * The base path of the given path
 * @param  string   path    The absolute path
 * @return string           The path up to the last dir/file
 */
func GetPath(path string) string {
    index := strings.LastIndex(path, "/")
    if index != -1 {
        return path
    }
    return path[:index]
}
