package util

import (
	"golang.org/x/crypto/ssh"
	"github.com/tmc/scp"
	"fmt"
	"strings"
)

var VERBOSE = true

var clients = make(map[string]*ssh.Client)

/**
 * Easy shorthand for multiple calls to sshExec
 * @param  string		host		The IP address of the host to execute the commands on
 * @param  ...string	commands	The commands to execute
 * @return []string					The results of the execution of each command 
 */
func SshMultiExec(host string, commands ...string) []string {
	
	out := make([]string,0);
	for _, command := range commands {
		
		res := SshExec(host,command)
		out = append(out,string(res))
	}
	return out
}

/**
 * Speeds up remote execution by chaining commands together
 * @param  string		host		The IP address of the host to execute the commands on
 * @param  ...string	commands	The commands to execute
 * @return string          			The result of the execution
 */
func SshFastMultiExec(host string, commands ...string) string {

	cmd := ""
	for i,command := range commands {
		
		if i != 0 {
			cmd += "&&"
		}
		cmd += command
	}
	return SshExec(host,cmd)
}

/**
 * Execute a command on a remote machine
 * @param  string	host	The host to execute the command on
 * @param  string	command	The command to execute
 * @return string			The result of execution
 */
func SshExec(host string,command string) string {
	if VERBOSE {
		fmt.Printf("Running command on %s : %s\n",host,command)
	}
	
	session, err := sshConnect(host)
	CheckAndPrint(err,fmt.Sprintf("Error with command %s:%s\n",host,command))

	out, err := session.CombinedOutput(command)
	CheckAndPrint(err,fmt.Sprintf("Error with command %s:%s\nResponse:%s\n\n",host,command,string(out)))
	session.Close()
	return string(out)
}

/**
 * Execute a command on a remote machine, and ignore a failed 
 * execution
 * @param  string	host	The host to execute the command on
 * @param  string	command	The command to execute
 * @return string			The result of execution
 */
func SshExecIgnore(host string,command string) string {
	if VERBOSE {
		fmt.Printf("Running command on %s : %s\n",host,command)
	}
	
	session,_ := sshConnect(host)
	out, _ := session.CombinedOutput(command)

	session.Close()
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
func sshConnect(host string) (*ssh.Session, error) {
	USER := "appo"
	PASS := "w@ntest"
	var client *ssh.Client
	var err error
	var isOpen bool
	client, isOpen = clients[host]

	if isOpen == false {
		sshConfig := &ssh.ClientConfig{
			User: USER,
			Auth: []ssh.AuthMethod{ssh.Password(PASS)},
		}
		sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

		client, err = ssh.Dial("tcp", fmt.Sprintf("%s:22",host), sshConfig)
		if err != nil {
			return  nil, err
		}
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return  nil, err
	}

	return session, nil
}

/**
 * DEPRECATED
 */
func initSCPR(host string,dir string){
	directories := LsDir(dir)
	dirStr := ""
	for _,dir := range directories {
		dirStr += " "+dir
	}
	SshExec(host,fmt.Sprintf("mkdir -p %s",dirStr))
}

/**
 * Copy a file over to a remote machine
 * @param  string	host	The IP address of the remote host
 * @param  string	src		The source path of the file
 * @param  string	dest	The destination path of the file
 */
func Scp(host string,src string,dest string){
	if VERBOSE {
		fmt.Printf("Copying %s to %s:%s...",src,host,dest)
	}
	
	session, err := sshConnect(host)
	CheckFatal(err)
	err = scp.CopyPath(src,dest,session)
	CheckAndPrint(err,"SCP failed")
	
	if VERBOSE {
		fmt.Printf("done\n")
	}
	session.Close()
	
}

/**
 * The base path of the given path
 * @param  string	path	The absolute path
 * @return string			The path up to the last dir/file
 */
func getPath(path string) string {
	index := strings.LastIndex(path,"/")
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
func scpr(host string,dir string){

	path := getPath(dir)
	SshExec(host,"mkdir -p "+path)

	file := fmt.Sprintf("%s.tar.gz",dir)
	BashExec(fmt.Sprintf("tar cfz %s %s",file,dir))
	Scp(host,file,file)
	SshExec(host,fmt.Sprintf("tar xfz %s && rm %s",file,file))
}

/**
 * Copy over a directory to a specified path
 * @param  string   host	The host to copy the directory to
 * @param  string   src		The source of the file/directory
 * @param  string   dest	The destination of the file/directory on the remote machine
 */
func scprd(host string,src string,dest string){
	initSCPR(host,dest+src)
	files := Lsr(src)
	//fmt.Printf("Files: %+v\n",files)
	for _,f := range files{
		Scp(host,f,dest+f)
	}
}