/*
	Copyright 2019 whiteblock Inc.
	This file is a part of the genesis.

	Genesis is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Genesis is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

// Package ssh contains abstractions which manage SSH connections for the user,
// allowing for faster and easier remote execution
package ssh

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/state"
	"github.com/whiteblock/genesis/util"
	"github.com/whiteblock/scp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/semaphore"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

var conf = util.GetConfig()

// Client maintains a persistent connect with a server,
// allowing commands to be run on that server. This object is thread safe.
type Client interface {

	// MultiRun provides an easy shorthand for multiple calls to sshExec
	MultiRun(commands ...string) ([]string, error)

	// FastMultiRun speeds up remote execution by chaining commands together
	FastMultiRun(commands ...string) (string, error)

	// Run executes a given command on the connected remote machine.
	Run(command string) (string, error)

	// KeepTryRun attempts to run a command successfully multiple times. It will
	// keep trying until it reaches the max amount of tries or it is successful once.
	KeepTryRun(command string) (string, error)

	// DockerExec executes a command inside of a node
	DockerExec(node Node, command string) (string, error)

	// DockerCp copies a file on a remote machine from source to the dest in the node
	DockerCp(node Node, source string, dest string) error

	// KeepTryDockerExec is like KeepTryRun for nodes
	KeepTryDockerExec(node Node, command string) (string, error)

	// KeepTryDockerExecAll is like KeepTryRun for nodes, but can handle more than one command.
	// Executes the given commands in order.
	KeepTryDockerExecAll(node Node, commands ...string) ([]string, error)

	// DockerExecd runs the given command, and then returns immediately.
	// This function will not return the output of the command.
	// This is useful if you are starting a persistent process inside a container
	DockerExecd(node Node, command string) (string, error)

	// DockerExecdit runs the given command, and then returns immediately.
	// This function will not return the output of the command.
	// This is useful if you are starting a persistent process inside a container.
	// Also flags the session as interactive and sets up a virtual tty.
	DockerExecdit(node Node, command string) (string, error)

	// DockerExecdLog will cause the stdout and stderr of the command to be stored in the logs.
	// Should only be used for the blockchain process.
	DockerExecdLog(node Node, command string) error

	// DockerExecdLogAppend will cause the stdout and stderr of the command to be stored in the logs.
	// Should only be used for the blockchain process. Will append to existing logs.
	DockerExecdLogAppend(node Node, command string) error

	// DockerRead will read a file on a node, if lines > -1 then
	// it will return the last `lines` lines of the file
	DockerRead(node Node, file string, lines int) (string, error)

	// DockerMultiExec will run all of the given commands strung together with && on
	// the given node.
	DockerMultiExec(node Node, commands []string) (string, error)

	// KTDockerMultiExec is like DockerMultiExec, except it keeps attempting the command after
	// failure
	KTDockerMultiExec(node Node, commands []string) (string, error)

	// Scp is a wrapper for the scp command. Can be used to copy
	// a file over to a remote machine.
	Scp(src string, dest string) error

	// Close cleans up the resources used by sshClient object
	Close()
}

type client struct {
	clients  []*ssh.Client
	host     string
	serverID int
	mux      *sync.RWMutex
	sem      *semaphore.Weighted
}

// NewClient creates an instance of Client, with a connection to the
// host server given.
func NewClient(host string, serverID int) (Client, error) {
	out := new(client)
	for i := conf.MaxConnections; i > 0; i -= 5 {
		c, err := sshConnect(host)
		if err != nil {
			return nil, util.LogError(err)
		}
		out.clients = append(out.clients, c)
	}
	out.host = host
	out.serverID = serverID
	out.mux = &sync.RWMutex{}
	out.sem = semaphore.NewWeighted(int64(conf.MaxConnections))
	return out, nil
}

func (sshClient *client) getSession() (*Session, error) {
	sshClient.mux.RLock()
	ctx := context.TODO()
	sshClient.sem.Acquire(ctx, 1)
	for _, client := range sshClient.clients {
		session, err := client.NewSession()
		if err != nil {
			continue
		}
		sshClient.mux.RUnlock()
		return NewSession(session, sshClient.sem), nil
	}
	sshClient.mux.RUnlock()

	client, err := sshConnect(sshClient.host)
	for err != nil && (strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "EOF")) {
		log.WithFields(log.Fields{"error": err}).Error("error connecting to remote host,retrying once")
		time.Sleep(50 * time.Millisecond)
		client, err = sshConnect(sshClient.host)
	}
	if client == nil {
		sshClient.sem.Release(1)
		return nil, fmt.Errorf("error(\"%s\"): client is nil", err.Error())
	}
	if err != nil {
		sshClient.sem.Release(1)
		return nil, util.LogError(err)
	}
	session, err := client.NewSession()
	if err != nil {
		sshClient.sem.Release(1)
		return nil, util.LogError(err)
	}
	sshClient.mux.Lock()
	sshClient.clients = append(sshClient.clients, client)
	sshClient.mux.Unlock()
	return NewSession(session, sshClient.sem), nil
}

// MultiRun provides an easy shorthand for multiple calls to sshExec
func (sshClient *client) MultiRun(commands ...string) ([]string, error) {

	out := []string{}
	for _, command := range commands {

		res, err := sshClient.Run(command)
		if err != nil {
			return nil, util.LogError(err)
		}
		out = append(out, res)
	}
	return out, nil
}

// FastMultiRun speeds up remote execution by chaining commands together
func (sshClient *client) FastMultiRun(commands ...string) (string, error) {

	cmd := ""
	for i, command := range commands {
		if i != 0 {
			cmd += "&&"
		}
		cmd += command
	}
	return sshClient.Run(cmd)
}

// Run executes a given command on the connected remote machine.
func (sshClient *client) Run(command string) (string, error) {
	session, err := sshClient.getSession()
	if err != nil {
		return "", util.LogError(err)
	}
	log.WithFields(log.Fields{"host": sshClient.host, "command": command}).Trace("executing command")

	bs := state.GetBuildStateByServerID(sshClient.serverID)
	defer session.Close()
	if bs.Stop() {
		return "", bs.GetError()
	}

	out, err := session.Get().CombinedOutput(command)
	if conf.MaxCommandOutputLogSize == -1 || len(out) <= conf.MaxCommandOutputLogSize {
		log.Infof("$ %s\n%s\n", command, out)
	} else {
		log.Infof("$ %s\n%s...\n", command, out[:conf.MaxCommandOutputLogSize])
	}

	if err != nil {
		return string(out), util.FormatError(string(out), err)
	}
	return string(out), nil
}

// KeepTryRun attempts to run a command successfully multiple times. It will
// keep trying until it reaches the max amount of tries or it is successful once.
func (sshClient *client) KeepTryRun(command string) (string, error) {
	var res string
	var err error
	bs := state.GetBuildStateByServerID(sshClient.serverID)
	if bs.Stop() {
		return "", bs.GetError()
	}
	for i := 0; i < conf.MaxRunAttempts; i++ {
		res, err = sshClient.Run(command)
		if err == nil {
			break
		}
	}
	return res, util.LogError(err)
}

// DockerExec executes a command inside of a node
func (sshClient *client) DockerExec(node Node, command string) (string, error) {
	return sshClient.Run(fmt.Sprintf("docker exec %s %s", node.GetNodeName(), command))
}

// DockerCp copies a file on a remote machine from source to the dest in the node
func (sshClient *client) DockerCp(node Node, source string, dest string) error {
	_, err := sshClient.Run(fmt.Sprintf("docker cp %s %s:%s", source, node.GetNodeName(), dest))
	return util.LogError(err)
}

// KeepTryDockerExec is like KeepTryRun for nodes
func (sshClient *client) KeepTryDockerExec(node Node, command string) (string, error) {
	return sshClient.KeepTryRun(fmt.Sprintf("docker exec %s %s", node.GetNodeName(), command))
}

// KeepTryDockerExecAll is like KeepTryRun for nodes, but can handle more than one command.
// Executes the given commands in order.
func (sshClient *client) KeepTryDockerExecAll(node Node, commands ...string) ([]string, error) {
	out := []string{}
	for _, command := range commands {
		res, err := sshClient.KeepTryRun(fmt.Sprintf("docker exec %s %s", node.GetNodeName(), command))
		if err != nil {
			return nil, util.LogError(err)
		}
		out = append(out, res)
	}
	return out, nil
}

// DockerExecd runs the given command, and then returns immediately.
// This function will not return the output of the command.
// This is useful if you are starting a persistent process inside a container
func (sshClient *client) DockerExecd(node Node, command string) (string, error) {
	return sshClient.Run(fmt.Sprintf("docker exec -d %s %s", node.GetNodeName(), command))
}

// DockerExecdit runs the given command, and then returns immediately.
// This function will not return the output of the command.
// This is useful if you are starting a persistent process inside a container.
// Also flags the session as interactive and sets up a virtual tty.
func (sshClient *client) DockerExecdit(node Node, command string) (string, error) {
	return sshClient.Run(fmt.Sprintf("docker exec -itd %s %s", node.GetNodeName(), command))
}

func (sshClient *client) logSanitizeAndStore(node Node, command string) {
	if strings.Count(command, "'") != strings.Count(command, "\\'") {
		log.Panic("DockerExecdLog commands cannot contain unescaped ' characters")
	}
	bs := state.GetBuildStateByServerID(sshClient.serverID)
	bs.Set(fmt.Sprintf("%d", node.GetAbsoluteNumber()), util.Command{Cmdline: command, ServerID: sshClient.serverID, Node: node.GetRelativeNumber()})
}

// DockerExecdLog will cause the stdout and stderr of the command to be stored in the logs.
// Should only be used for the blockchain process.
func (sshClient *client) DockerRunMainDaemon(node Node, command string) error {
	sshClient.logSanitizeAndStore(node, command)
	return sshClient.DockerExecdLog(node,command)
}

// DockerExecdLog will cause the stdout and stderr of the command to be stored in the logs.
// Should only be used for the blockchain process.
func (sshClient *client) DockerExecdLog(node Node, command string) error {
	sshClient.logSanitizeAndStore(node, command)

	_, err := sshClient.Run(fmt.Sprintf("docker exec -d %s bash -c '%s 2>&1 > %s'", node.GetNodeName(),
		command, conf.DockerOutputFile))
	return util.LogError(err)
}

// DockerExecdLogAppend will cause the stdout and stderr of the command to be stored in the logs.
// Should only be used for the blockchain process. Will append to existing logs.
func (sshClient *client) DockerExecdLogAppend(node Node, command string) error {
	sshClient.logSanitizeAndStore(node, command)
	_, err := sshClient.Run(fmt.Sprintf("docker exec -d %s bash -c '%s 2>&1 >> %s'", node.GetNodeName(),
		command, conf.DockerOutputFile))
	return util.LogError(err)
}

// DockerRead will read a file on a node, if lines > -1 then
// it will return the last `lines` lines of the file
func (sshClient *client) DockerRead(node Node, file string, lines int) (string, error) {
	if lines > -1 {
		return sshClient.DockerExec(node, fmt.Sprintf("tail -n %d %s", lines, file))
	}
	return sshClient.DockerExec(node, fmt.Sprintf("cat %s", file))
}

func (sshClient *client) dockerMultiExec(node Node, commands []string, kt bool) (string, error) {
	mergedCommand := ""

	for _, command := range commands {
		if len(mergedCommand) != 0 {
			mergedCommand += "&&"
		}
		mergedCommand += fmt.Sprintf("docker exec -d %s %s", node.GetNodeName(), command)
	}
	if kt {
		return sshClient.KeepTryRun(mergedCommand)
	}
	return sshClient.Run(mergedCommand)
}

// DockerMultiExec will run all of the given commands strung together with && on
// the given node.
func (sshClient *client) DockerMultiExec(node Node, commands []string) (string, error) {
	return sshClient.dockerMultiExec(node, commands, false)
}

// KTDockerMultiExec is like DockerMultiExec, except it keeps attempting the command after
// failure
func (sshClient *client) KTDockerMultiExec(node Node, commands []string) (string, error) {
	return sshClient.dockerMultiExec(node, commands, true)
}

// Scp is a wrapper for the scp command. Can be used to copy
// a file over to a remote machine.
func (sshClient *client) Scp(src string, dest string) error {
	log.WithFields(log.Fields{"src": src, "dst": dest}).Info("remote copying file")

	if !strings.HasPrefix(src, "./") && src[0] != '/' {
		bs := state.GetBuildStateByServerID(sshClient.serverID)
		src = "/tmp/" + bs.BuildID + "/" + src
	}

	session, err := sshClient.getSession()
	if err != nil {
		return util.LogError(err)
	}
	defer session.Close()

	return scp.CopyPath(src, dest, session.Get())
}

/*
   Scpr copies over a directory to a specified path on a remote host

func (sshClient Client) Scpr(dir string) error {

	path := GetPath(dir)
	_, err := sshClient.Run("mkdir -p " + path)
	if err != nil {
		return util.LogError(err)
	}

	file := fmt.Sprintf("%s.tar.gz", dir)
	_, err = BashExec(fmt.Sprintf("tar cfz %s %s", file, dir))
	if err != nil {
		return util.LogError(err)
	}
	err = sshClient.Scp(file, file)
	if err != nil {
		return util.LogError(err)
	}
	_, err = sshClient.Run(fmt.Sprintf("tar xfz %s && rm %s", file, file))
	return err
}*/

// Close cleans up the resources used by sshClient object
func (sshClient *client) Close() {
	sshClient.mux.Lock()
	defer sshClient.mux.Unlock()
	for _, client := range sshClient.clients {
		if client == nil {
			continue
		}
		client.Close()
	}
}

func sshConnect(host string) (*ssh.Client, error) {

	key, err := ioutil.ReadFile(conf.SSHKey)
	if err != nil {
		return nil, util.LogError(err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, util.LogError(err)
	}
	sshConfig := &ssh.ClientConfig{
		User: conf.SSHUser,
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
		log.WithFields(log.Fields{"host": host, "user": sshConfig.User,
			"keyLoc": conf.SSHKey}).Error("unable to establish an ssh connection")
		return nil, util.LogError(err)
	}

	return client, nil
}
