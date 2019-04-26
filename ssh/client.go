package ssh

import (
	"../state"
	"../util"
	"context"
	"fmt"
	"github.com/Whiteblock/scp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/semaphore"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"
)

var conf = util.GetConfig()

const maxRunAttempts int = 20

const maxConnections int = 50

/*
   Client maintains a persistent connect with a server,
   allowing commands to be run on that server. This object is thread safe.
*/
type Client struct {
	clients  []*ssh.Client
	host     string
	serverId int
	mux      *sync.RWMutex
	sem      *semaphore.Weighted
}

/*
   NewClient creates an instance of Client, with a connection to the
   host server given.
*/
func NewClient(host string, serverId int) (*Client, error) {
	out := new(Client)
	for i := maxConnections; i > 0; i -= 5 {
		client, err := sshConnect(host)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		out.clients = append(out.clients, client)
	}
	out.host = host
	out.serverId = serverId
	out.mux = &sync.RWMutex{}
	out.sem = semaphore.NewWeighted(int64(maxConnections))
	return out, nil
}

func (this *Client) getSession() (*Session, error) {
	this.mux.RLock()
	ctx := context.TODO()
	this.sem.Acquire(ctx, 1)
	for _, client := range this.clients {
		session, err := client.NewSession()
		if err != nil {
			continue
		}
		this.mux.RUnlock()
		return NewSession(session, this.sem), nil
	}
	this.mux.RUnlock()

	client, err := sshConnect(this.host)
	for err != nil && (strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "EOF")) {
		log.Println(err)
		time.Sleep(50 * time.Millisecond)
		client, err = sshConnect(this.host)
	}
	if client == nil {
		this.sem.Release(1)
		return nil, fmt.Errorf("Client is nil! Error(\"%s\")", err.Error())
	}
	if err != nil {
		this.sem.Release(1)
		log.Println(err)
		return nil, err
	}
	session, err := client.NewSession()
	if err != nil {
		this.sem.Release(1)
		log.Println(err)
		return nil, err
	}
	this.mux.Lock()
	this.clients = append(this.clients, client)
	this.mux.Unlock()
	return NewSession(session, this.sem), nil
}

/*
   MultiRun provides an easy shorthand for multiple calls to sshExec
*/
func (this *Client) MultiRun(commands ...string) ([]string, error) {

	out := []string{}
	for _, command := range commands {

		res, err := this.Run(command)
		if err != nil {
			return nil, err
		}
		out = append(out, string(res))
	}
	return out, nil
}

/*
   FastMultiRun speeds up remote execution by chaining commands together

*/
func (this *Client) FastMultiRun(commands ...string) (string, error) {

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
func (this *Client) Run(command string) (string, error) {
	session, err := this.getSession()
	if conf.Verbose {
		fmt.Printf("Running command: %s\n", command)
	}
	bs := state.GetBuildStateByServerId(this.serverId)
	defer session.Close()
	if bs.Stop() {
		return "", bs.GetError()
	}

	if err != nil {
		log.Println(err)
		return "", err
	}

	out, err := session.Get().CombinedOutput(command)
	if conf.Verbose {
		fmt.Println(string(out))
	}
	if err != nil {
		return string(out), util.FormatError(string(out), err)
	}
	return string(out), nil
}

/*
   KeepTryRun attempts to run a command successfully multiple times. It will
   keep trying until it reaches the max amount of tries or it is successful once.
*/
func (this *Client) KeepTryRun(command string) (string, error) {
	var res string
	var err error
	bs := state.GetBuildStateByServerId(this.serverId)
	if bs.Stop() {
		return "", bs.GetError()
	}
	for i := 0; i < maxRunAttempts; i++ {
		res, err = this.Run(command)
		if err == nil {
			break
		}
	}
	return res, err
}

/*
   DockerExec executes a command inside of a node
*/
func (this *Client) DockerExec(node int, command string) (string, error) {
	return this.Run(fmt.Sprintf("docker exec %s%d %s", conf.NodePrefix, node, command))
}

/*
   Run docker cp on a remote machine, coping a file from source to dest in the node
*/
func (this *Client) DockerCp(node int, source string, dest string) error {
	_, err := this.Run(fmt.Sprintf("docker cp %s %s%d:%s", source, conf.NodePrefix, node, dest))
	return err
}

/*
   KeepTryDockerExec is like KeepTryRun for nodes
*/
func (this *Client) KeepTryDockerExec(node int, command string) (string, error) {
	return this.KeepTryRun(fmt.Sprintf("docker exec %s%d %s", conf.NodePrefix, node, command))
}

/*
   KeepTryDockerExecAll is like KeepTryRun for nodes
*/
func (this *Client) KeepTryDockerExecAll(node int, commands ...string) ([]string, error) {
	out := []string{}
	for _, command := range commands {
		res, err := this.KeepTryRun(fmt.Sprintf("docker exec %s%d %s", conf.NodePrefix, node, command))
		if err != nil {
			return nil, err
		}
		out = append(out, res)
	}
	return out, nil
}

/*
   DockerExecd runs the given command, and then returns immediately.
   This function will not return the output of the command.
   This is useful if you are starting a persistent process inside a container
*/
func (this *Client) DockerExecd(node int, command string) (string, error) {
	return this.Run(fmt.Sprintf("docker exec -d %s%d %s", conf.NodePrefix, node, command))
}

/*
   DockerExecd runs the given command, and then returns immediately.
   This function will not return the output of the command.
   This is useful if you are starting a persistent process inside a container
*/
func (this *Client) DockerExecdit(node int, command string) (string, error) {
	return this.Run(fmt.Sprintf("docker exec -itd %s%d %s", conf.NodePrefix, node, command))
}

func (this *Client) logSanitizeAndStore(node int, command string) {
	if strings.Count(command, "'") != strings.Count(command, "\\'") {
		panic("DockerExecdLog commands cannot contain unescaped ' characters")
	}
	bs := state.GetBuildStateByServerId(this.serverId)
	bs.Set(fmt.Sprintf("%d", node), util.Command{Cmdline: command, ServerId: this.serverId, Node: node})
}

/*
   DockerExecdLog will cause the stdout and stderr of the command to be stored in the logs.
   Should only be used for the blockchain process.
*/
func (this *Client) DockerExecdLog(node int, command string) error {
	this.logSanitizeAndStore(node, command)

	_, err := this.Run(fmt.Sprintf("docker exec -d %s%d bash -c '%s 2>&1 > %s'", conf.NodePrefix,
		node, command, conf.DockerOutputFile))
	return err
}

/*
   DockerExecdLogAppend will cause the stdout and stderr of the command to be stored in the logs.
   Should only be used for the blockchain process. Will append to existing logs.
*/
func (this *Client) DockerExecdLogAppend(node int, command string) error {
	this.logSanitizeAndStore(node, command)
	_, err := this.Run(fmt.Sprintf("docker exec -d %s%d bash -c '%s 2>&1 >> %s'", conf.NodePrefix,
		node, command, conf.DockerOutputFile))
	return err
}

/*
   DockerRead will read the current output of the command ran with DockerExecdLog. Must be called
   after DockerExecdLog
*/
func (this *Client) DockerRead(node int, file string, lines int) (string, error) {
	if lines > -1 {
		return this.Run(fmt.Sprintf("docker exec %s%d tail -n %d %s", conf.NodePrefix, node, lines, file))
	}
	return this.Run(fmt.Sprintf("docker exec %s%d cat %s", conf.NodePrefix, node, file))
}

/*
   DockerMultiExec will run all of the given commands strung together with && on
   the given node.
*/
func (this *Client) DockerMultiExec(node int, commands []string) (string, error) {
	merged_command := ""

	for _, command := range commands {
		if len(merged_command) != 0 {
			merged_command += "&&"
		}
		merged_command += fmt.Sprintf("docker exec -d %s%d %s", conf.NodePrefix, node, command)
	}

	return this.Run(merged_command)
}

/*
   KTDockerMultiExec is like DockerMultiExec, except it keeps attempting the command after
   failure
*/
func (this *Client) KTDockerMultiExec(node int, commands []string) (string, error) {
	merged_command := ""

	for _, command := range commands {
		if len(merged_command) != 0 {
			merged_command += "&&"
		}
		merged_command += fmt.Sprintf("docker exec -d %s%d %s", conf.NodePrefix, node, command)
	}

	return this.KeepTryRun(merged_command)
}

/*
   Scp is a wrapper for the scp command. Can be used to copy
   a file over to a remote machine.
*/
func (this *Client) Scp(src string, dest string) error {
	if conf.Verbose {
		fmt.Printf("Remote copying %s to %s...", src, dest)
	}
	if !strings.HasPrefix(src, "./") && src[0] != '/' {
		bs := state.GetBuildStateByServerId(this.serverId)
		src = "/tmp/" + bs.BuildId + "/" + src
	}

	session, err := this.getSession()
	if err != nil {
		return err
	}
	defer session.Close()

	err = scp.CopyPath(src, dest, session.Get())
	if err != nil {
		return err
	}

	if conf.Verbose {
		fmt.Printf("done\n")
	}

	return nil
}

/*
   Scp is a wrapper for the scp command. Can be used to copy
   a file over to a remote machine.
*/
func (this *Client) InternalScp(src string, dest string) error {
	if conf.Verbose {
		fmt.Printf("Remote copying %s to %s...", src, dest)
	}

	session, err := this.getSession()
	if err != nil {
		return err
	}
	defer session.Close()

	err = scp.CopyPath(src, dest, session.Get())
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

func (this Client) Scpr(dir string) error {

	path := GetPath(dir)
	_, err := this.Run("mkdir -p " + path)
	if err != nil {
		log.Println(err)
		return err
	}

	file := fmt.Sprintf("%s.tar.gz", dir)
	_, err = BashExec(fmt.Sprintf("tar cfz %s %s", file, dir))
	if err != nil {
		log.Println(err)
		return err
	}
	err = this.Scp(file, file)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = this.Run(fmt.Sprintf("tar xfz %s && rm %s", file, file))
	return err
}*/

/*
   Clean up the resources used by this object
*/
func (this *Client) Close() {
	this.mux.Lock()
	defer this.mux.Unlock()
	for _, client := range this.clients {
		if client == nil {
			continue
		}
		client.Close()
	}
}

func sshConnect(host string) (*ssh.Client, error) {

	key, err := ioutil.ReadFile(conf.SshKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	sshConfig := &ssh.ClientConfig{
		User: conf.SshUser,
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", host), sshConfig)
	i := 0
	for err != nil && i < 10 {
		client, err = ssh.Dial("tcp", fmt.Sprintf("%s:22", host), sshConfig)
		i++
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return client, nil
}
