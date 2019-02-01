package util

import (
    "strings"
    "log"
    "golang.org/x/crypto/ssh"
    "fmt"
    "github.com/tmc/scp"
    "errors"
    state "../state"
)

const maxRunAttempts int = 5

/*
    SshClient maintains a persistent connect with a server,
    allowing commands to be run on that server. This object is thread safe.
 */
type SshClient struct {
    clients     []*ssh.Client
}
/*
    NewSshClient creates an instance of SshClient, with a connection to the 
    host server given. 
 */
func NewSshClient(host string) (*SshClient,error){
    out := new(SshClient)
    for i := conf.ThreadLimit; i > 0; i -= 5 {
        client,err := sshConnect(host)
        if err != nil{
            log.Println(err)
            return nil,err
        }
        out.clients = append(out.clients,client)
    }
    return out,nil
}

func (this SshClient) getSession() (*ssh.Session,error) {
    for _, client := range this.clients {
        session, err := client.NewSession()
        if err != nil {
            continue
        }
        return session,nil
    }
    return nil,errors.New("Unable to get a session")
}

/*
    MultiRun provides an easy shorthand for multiple calls to sshExec
 */
func (this SshClient) MultiRun(commands ...string) ([]string,error) {

    out := []string{}
    for _, command := range commands {

        res,err := this.Run(command)
        if err != nil{
            return nil,err
        }
        out = append(out, string(res))
    }
    return out,nil
}

/*
    FastMultiRun speeds up remote execution by chaining commands together

 */
func (this SshClient) FastMultiRun(commands ...string) (string,error) {

    cmd := ""
    for i, command := range commands {

        if i != 0 {
            cmd += "&&"
        }
        cmd += command
    }
    return this.Run(cmd)
}

/*
    Run executes a given command on the connected remote machine. 
 */
func (this SshClient) Run(command string) (string,error) {
    session, err := this.getSession()
    if conf.Verbose {
        fmt.Printf("Running command: %s\n", command)
    }
    if state.Stop() {
        return "",state.GetError()
    }
    
    if err != nil {
        log.Println(err)
        return "",err
    }
    defer session.Close()
    out, err := session.CombinedOutput(command)
    if conf.Verbose {
        fmt.Println(string(out))
    }
    return string(out),err
}

/*
    KeepTryRun attempts to run a command successfully multiple times. It will 
    keep trying until it reaches the max amount of tries or it is successful once.
 */
func (this SshClient) KeepTryRun(command string) (string,error) {
    var res string
    var err error
    if state.Stop() {
        return "",state.GetError()
    }
    for i := 0; i < maxRunAttempts; i++ {
        res,err = this.Run(command)
        if err == nil{
            break
        }
    }
    return res,err
}

/*
    DockerExec executes a command inside of a node
 */
func (this SshClient) DockerExec(node int,command string) (string,error) {
    return this.Run(fmt.Sprintf("docker exec %s%d %s",conf.NodePrefix,node,command))
}

/*
    KeepTryDockerExec is like KeepTryRun for nodes
 */
func (this SshClient) KeepTryDockerExec(node int,command string) (string,error) {
    return this.KeepTryRun(fmt.Sprintf("docker exec %s%d %s",conf.NodePrefix,node,command))
}

/*
    DockerExecd runs the given command, and then returns immediately. 
    This function will not return the output of the command. 
    This is useful if you are starting a persistent process inside a container
 */
func (this SshClient) DockerExecd(node int,command string) (string,error) {
    return this.Run(fmt.Sprintf("docker exec -d %s%d %s",conf.NodePrefix,node,command))
}

/*
    DockerExecdLog will cause the stdout and stderr of the command to be stored in the logs. 
    Should only be used for the blockchain process.
 */
func (this SshClient) DockerExecdLog(node int,command string) error {
    if strings.Count(command,"'") != strings.Count(command,"\\'"){
        panic("DockerExecdLog commands cannot contain unescaped ' characters")
    }
    _,err := this.Run(fmt.Sprintf("docker exec -d %s%d bash -c '%s 2>&1 >> %s'",conf.NodePrefix,
                                        node,command,conf.DockerOutputFile))
    return err
}

/*
    DockerRead will read the current output of the command ran with DockerExecdLog. Must be called 
    after DockerExecdLog
 */
func (this SshClient) DockerRead(node int,file string) (string,error) {
    return this.Run(fmt.Sprintf("docker exec %s%d cat %s",conf.NodePrefix,node,file))
}

/*
    DockerMultiExec will run all of the given commands strung together with && on
    the given node. 
 */
func (this SshClient) DockerMultiExec(node int,commands []string) (string,error){
    merged_command := ""

    for _,command := range commands {
        if len(merged_command) != 0 {
            merged_command += "&&"
        }
        merged_command += fmt.Sprintf("docker exec -d %s%d %s",conf.NodePrefix,node,command)
    }

    return this.Run(merged_command)
}

/*
    KTDockerMultiExec is like DockerMultiExec, except it keeps attempting the command after
    failure
 */
func (this SshClient) KTDockerMultiExec(node int,commands []string) (string,error){
    merged_command := ""

    for _,command := range commands {
        if len(merged_command) != 0 {
            merged_command += "&&"
        }
        merged_command += fmt.Sprintf("docker exec -d %s%d %s",conf.NodePrefix,node,command)
    }

    return this.KeepTryRun(merged_command)
}



/*
    Scp is a wrapper for the scp command. Can be used to copy
    a file over to a remote machine.
 */
func (this SshClient) Scp(src string, dest string) error {
    if conf.Verbose {
        fmt.Printf("Remote copying %s to %s...", src, dest)
    }   
    session, err := this.getSession()
    if err != nil {
        return err
    }
    defer session.Close()

    err = scp.CopyPath(src, dest, session)
    if err != nil {
        return err
    }

    if conf.Verbose {
        fmt.Printf("done\n")
    }

    return nil
}

/*
    Scpr copies over a directory to a specified path on a remote host
*/
func (this SshClient) Scpr(dir string) error {

    path := GetPath(dir)
    _,err := this.Run("mkdir -p "+path)
    if err != nil {
        log.Println(err)
        return err
    }

    file := fmt.Sprintf("%s.tar.gz", dir)
    _,err = BashExec(fmt.Sprintf("tar cfz %s %s", file, dir))
    if err != nil {
        log.Println(err)
        return err
    }
    err = this.Scp(file, file)
    if err != nil{
        log.Println(err)
        return err
    }
    _,err = this.Run(fmt.Sprintf("tar xfz %s && rm %s", file, file))
    return err
}

func (this SshClient) Close(){
    for _,client := range this.clients {
        if client == nil {
            continue
        }
        client.Close()
    }
}
