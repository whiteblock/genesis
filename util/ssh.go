package util

import (
    "fmt"
    "strings"
    "io/ioutil"
    "golang.org/x/crypto/ssh"
)

func sshConnect(host string) (*ssh.Client, error) {
    sshConfig := &ssh.ClientConfig{
        User: conf.SshUser,
        Auth: []ssh.AuthMethod{ssh.Password(conf.SshPassword)},
    }
    sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

    client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", host), sshConfig)

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

/*
    GetPath extracts the base path of the given path
 */
func GetPath(path string) string {
    index := strings.LastIndex(path, "/")
    if index != -1 {
        return path
    }
    return path[:index]
}
