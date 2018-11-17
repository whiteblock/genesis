package util

import (
	"fmt"
	"strings"
	"io/ioutil"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
)

/**
 * Easy shorthand for multiple calls to sshExec
 * @param  string		host		The IP address of the host to execute the commands on
 * @param  ...string	commands	The commands to execute
 * @return []string					The results of the execution of each command
 */
func SshMultiExec(host string, commands ...string) []string {

	out := make([]string, 0)
	for _, command := range commands {

		res := SshExec(host, command)
		out = append(out, string(res))
	}
	return out
}

/**
 * Speeds up remote execution by chaining commands together
 * @param  string		host		The IP address of the host to execute the commands on
 * @param  ...string	commands	The commands to execute
 * @return string          			The result of the execution
 */
func sshFastMultiExec(host string, commands ...string) string {

	cmd := ""
	for i, command := range commands {

		if i != 0 {
			cmd += "&&"
		}
		cmd += command
	}
	return SshExec(host, cmd)
}

/**
 * Execute a command on a remote machine
 * @param  string	host	The host to execute the command on
 * @param  string	command	The command to execute
 * @return string			The result of execution
 */
func SshExec(host string, command string) string {
	if conf.Verbose {
		fmt.Printf("Running command on %s : %s\n", host, command)
	}

	session,client,err := sshConnect(host)
	
	CheckAndPrint(err, fmt.Sprintf("Error with command %s:%s\n", host, command))
	defer session.Close()
	defer client.Close()
	out, err := session.CombinedOutput(command)
	CheckAndPrint(err, fmt.Sprintf("Error with command %s:%s\nResponse:%s\n\n", host, command, string(out)))

	return string(out)
}

func SshExecCheck(host string, command string) (string, error) {
	if conf.Verbose {
		fmt.Printf("Running command on %s : %s\n", host, command)
	}

	session,client, err := sshConnect(host)
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
	session.Close()
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

	session, client,err := sshConnect(host)
	if err != nil{
		panic(err)
	}
	defer client.Close()
	defer session.Close()
	out, _ := session.CombinedOutput(command)

	
	return string(out)
}

/**
 * Checks to see if a connection to the remote host is established,
 * if not then establish it, and then create a new Session on that
 * connection
 * @param  string		host	The host to get a connection
 * @return *ssh.Session			The new SSH connection
 * @return error				The error, if any occured
 */
func sshConnect(host string) (*ssh.Session,*ssh.Client, error) {
	var client *ssh.Client
	var err error
	

	sshConfig := &ssh.ClientConfig{
		User: conf.SshUser,
		Auth: []ssh.AuthMethod{ssh.Password(conf.SshPassword)},
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err = ssh.Dial("tcp", fmt.Sprintf("%s:22", host), sshConfig)
	if err != nil {
		fmt.Println("First ssh attempt failed: " + err.Error())
	}
	if err != nil {//Try to connect using the id_rsa file
		key, err := ioutil.ReadFile(conf.RsaKey)
		if err != nil {
			return nil,nil,err
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, nil, err
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
			return nil,nil,err
		}
	}
		


	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
		
	}

	return session,client, nil
}

/**
 * DEPRECATED
 */
func InitSCPR(host string, dir string) {
	directories := LsDir(dir)
	dirStr := ""
	for _, dir := range directories {
		dirStr += " " + dir
	}
	SshExec(host, fmt.Sprintf("mkdir -p %s", dirStr))
}

/**
 * Copy a file over to a remote machine
 * @param  string	host	The IP address of the remote host
 * @param  string	src		The source path of the file
 * @param  string	dest	The destination path of the file
 */
func Scp(host string, src string, dest string) {
	if conf.Verbose {
		fmt.Printf("Copying %s to %s:%s...", src, host, dest)
	}

	session,client, err := sshConnect(host)
	defer session.Close()
	defer client.Close()
	CheckFatal(err)
	err = scp.CopyPath(src, dest, session)
	CheckAndPrint(err, "SCP failed")

	if conf.Verbose {
		fmt.Printf("done\n")
	}
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
func Scpr(host string, dir string) {

	path := GetPath(dir)
	SshExec(host, "mkdir -p "+path)

	file := fmt.Sprintf("%s.tar.gz", dir)
	BashExec(fmt.Sprintf("tar cfz %s %s", file, dir))
	Scp(host, file, file)
	SshExec(host, fmt.Sprintf("tar xfz %s && rm %s", file, file))
}

/**
 * Copy over a directory to a specified path
 * @param  string   host	The host to copy the directory to
 * @param  string   src		The source of the file/directory
 * @param  string   dest	The destination of the file/directory on the remote machine
 */
func Scprd(host string, src string, dest string) {
	InitSCPR(host, dest+src)
	files := Lsr(src)
	//fmt.Printf("Files: %+v\n",files)
	for _, f := range files {
		Scp(host, f, dest+f)
	}
}
