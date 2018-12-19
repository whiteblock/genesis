package util

import (
	"fmt"
	"strings"
	"io/ioutil"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
	"log"
)

/**
 * Easy shorthand for multiple calls to sshExec
 * @param  string		host		The IP address of the host to execute the commands on
 * @param  ...string	commands	The commands to execute
 * @return []string					The results of the execution of each command
 */
func SshMultiExec(host string, commands ...string) ([]string,error) {

	out := []string{}
	for _, command := range commands {

		res,err := SshExec(host, command)
		if err != nil{
			return nil,err
		}
		out = append(out, string(res))
	}
	return out,nil
}

/**
 * Speeds up remote execution by chaining commands together
 * @param  string		host		The IP address of the host to execute the commands on
 * @param  ...string	commands	The commands to execute
 * @return string          			The result of the execution
 */
func SshFastMultiExec(host string, commands ...string) (string,error) {

	cmd := ""
	for i, command := range commands {

		if i != 0 {
			cmd += "&&"
		}
		cmd += command
	}
	return SshExec(host, cmd)
}

func SshExec(host string, command string) (string, error) {
	if conf.Verbose {
		fmt.Printf("Running command on %s : %s\n", host, command)
	}

	session,client, err := _sshConnect(host)
	if err != nil {
		fmt.Printf("Error with command %s:%s\n", host, command)
		return "", err
	}
	defer session.Close()
	defer client.Close()

	out, err := session.CombinedOutput(command)
	if err != nil {
		fmt.Printf("Error with command %s:%s\nResponse:%s\n\n", host, command, string(out))
		return string(out), err
	}
	return string(out), nil
}



/**
 * Execute a command on a remote machine, and ignore a failed
 * execution
 * @param  string	host	The host to execute the command on
 * @param  string	command	The command to execute
 * @return string			The result of execution
 */
func SshExecIgnore(host string, command string) string {
	if conf.Verbose {
		fmt.Printf("Running command on %s : %s\n", host, command)
	}

	session, client,err := _sshConnect(host)
	if err != nil{
		panic(err)
	}
	defer client.Close()
	defer session.Close()
	out, _ := session.CombinedOutput(command)

	
	return string(out)
}

func DockerExec(host string,node int,command string) (string,error) {
	return SshExec(host,fmt.Sprintf("docker exec whiteblock-node%d %s",node,command))
}

func DockerExecd(host string,node int,command string) (string,error) {
	return SshExec(host,fmt.Sprintf("docker exec -d whiteblock-node%d %s",node,command))
}

func DockerExecdLog(host string,node int,command string) error {
	if strings.Count(command,"'") != strings.Count(command,"\\'"){
		panic("DockerExecdLog commands cannot contain unescaped ' characters")
	}
	_,err := SshExec(host,fmt.Sprintf("docker exec -d whiteblock-node%d bash -c '%s 2>&1 >> %s'",node,command,conf.DockerOutputFile))
	return err
}

func DockerRead(host string,node int,file string) (string,error) {
	return SshExec(host,fmt.Sprintf("docker exec whiteblock-node%d cat %s",node,file))
}

func DockerMultiExec(host string,node int,commands []string) (string,error){
	merged_command := ""

	for _,command := range commands {
		if len(merged_command) != 0 {
			merged_command += "&&"
		}
		merged_command += fmt.Sprintf("docker exec -d whiteblock-node%d %s",node,command)
	}

	return SshExec(host,merged_command)
}

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
	if err != nil {
		fmt.Println("First ssh attempt failed: " + err.Error())
	}
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
 * DEPRECATED
 */
func InitSCPR(host string, dir string) error {
	directories,err := LsDir(dir)
	if err != nil{
		return err
	}
	dirStr := ""
	for _, dir := range directories {
		dirStr += " " + dir
	}
	_,err = SshExec(host, fmt.Sprintf("mkdir -p %s", dirStr))
	return err
}

/**
 * Copy a file over to a remote machine
 * @param  string	host	The IP address of the remote host
 * @param  string	src		The source path of the file
 * @param  string	dest	The destination path of the file
 */
func Scp(host string, src string, dest string) error {
	if conf.Verbose {
		fmt.Printf("Copying %s to %s:%s...", src, host, dest)
	}

	session,client, err := _sshConnect(host)
	if err != nil {
		return err
	}
	defer session.Close()
	defer client.Close()
	
	err = scp.CopyPath(src, dest, session)
	if err != nil {
		return err
	}

	if conf.Verbose {
		fmt.Printf("done\n")
	}

	return nil
}

/**
 * The base path of the given path
 * @param  string	path	The absolute path
 * @return string			The path up to the last dir/file
 */
func GetPath(path string) string {
	index := strings.LastIndex(path, "/")
	if index != -1 {
		return path
	}
	return path[:index]
}

/**
 * Copy over a directory to a specified path on a remote host
 * @param  string	host	The host to copy the directory to
 * @param  string	dir		The directory to copy over
 */
func Scpr(host string, dir string) error {

	path := GetPath(dir)
	_,err := SshExec(host, "mkdir -p "+path)
	if err != nil {
		return err
	}

	file := fmt.Sprintf("%s.tar.gz", dir)
	_,err = BashExec(fmt.Sprintf("tar cfz %s %s", file, dir))
	if err != nil {
		return err
	}
	err = Scp(host, file, file)
	if err != nil{
		return err
	}
	_,err = SshExec(host, fmt.Sprintf("tar xfz %s && rm %s", file, file))
	return err
}

/**
 * Copy over a directory to a specified path
 * @param  string   host	The host to copy the directory to
 * @param  string   src		The source of the file/directory
 * @param  string   dest	The destination of the file/directory on the remote machine
 */
func Scprd(host string, src string, dest string) error {
	InitSCPR(host, dest+src)
	files,err := Lsr(src)
	if err != nil {
		return err
	}
	//fmt.Printf("Files: %+v\n",files)
	for _, f := range files {
		err = Scp(host, f, dest+f)
		if err != nil {
			return err
		}
	}
	return nil
}
