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

package ssh

import (
	"fmt"
)

type fakeClient struct {
	responses []string //The responses which will be returned
}

// GetTestClient gets an ssh client which is only suitable for testing, and
// will return the given respones in order. Once they are exhausted, it will then
// give errors
func GetTestClient(responses []string) Client {
	out := new(fakeClient)
	return out
}

// MultiRun provides an easy shorthand for multiple calls to sshExec
func (fc *fakeClient) MultiRun(commands ...string) ([]string, error) {
	out := []string{}
	for _, command := range commands {

		res, err := fc.Run(command)
		if err != nil {
			return nil, err
		}
		out = append(out, res)
	}
	return out, nil
}

// FastMultiRun speeds up remote execution by chaining commands together
func (fc *fakeClient) FastMultiRun(commands ...string) (string, error) {
	return fc.Run(commands[0])
}

// Run executes a given command on the connected remote machine.
func (fc *fakeClient) Run(command string) (string, error) {
	if len(fc.responses) == 0 {
		return "", fmt.Errorf("no more responses")
	}
	out := fc.responses[0]
	fc.responses = fc.responses[1:]
	return out, nil
}

// KeepTryRun attempts to run a command successfully multiple times. It will
// keep trying until it reaches the max amount of tries or it is successful once.
func (fc *fakeClient) KeepTryRun(command string) (string, error) {
	return fc.Run(command)
}

// DockerExec executes a command inside of a node
func (fc *fakeClient) DockerExec(node Node, command string) (string, error) {
	return fc.Run(command)
}

// DockerCp copies a file on a remote machine from source to the dest in the node
func (fc *fakeClient) DockerCp(node Node, source string, dest string) error {
	return nil
}

// KeepTryDockerExec is like KeepTryRun for nodes
func (fc *fakeClient) KeepTryDockerExec(node Node, command string) (string, error) {
	return fc.Run(command)
}

// KeepTryDockerExecAll is like KeepTryRun for nodes, but can handle more than one command.
// Executes the given commands in order.
func (fc *fakeClient) KeepTryDockerExecAll(node Node, commands ...string) ([]string, error) {
	return fc.MultiRun(commands...)
}

// DockerExecd runs the given command, and then returns immediately.
// This function will not return the output of the command.
// This is useful if you are starting a persistent process inside a container
func (fc *fakeClient) DockerExecd(node Node, command string) (string, error) {
	return fc.Run(command)
}

// DockerExecdit runs the given command, and then returns immediately.
// This function will not return the output of the command.
// This is useful if you are starting a persistent process inside a container.
// Also flags the session as interactive and sets up a virtual tty.
func (fc *fakeClient) DockerExecdit(node Node, command string) (string, error) {
	return fc.Run(command)
}

// DockerExecdLog will cause the stdout and stderr of the command to be stored in the logs.
// Should only be used for the blockchain process.
func (fc *fakeClient) DockerExecdLog(node Node, command string) error {
	return nil
}

// DockerExecdLogAppend will cause the stdout and stderr of the command to be stored in the logs.
// Should only be used for the blockchain process. Will append to existing logs.
func (fc *fakeClient) DockerExecdLogAppend(node Node, command string) error {
	return nil
}

// DockerRead will read a file on a node, if lines > -1 then
// it will return the last `lines` lines of the file
func (fc *fakeClient) DockerRead(node Node, file string, lines int) (string, error) {
	return fc.Run(file)
}

// DockerMultiExec will run all of the given commands strung together with && on
// the given node.
func (fc *fakeClient) DockerMultiExec(node Node, commands []string) (string, error) {
	return fc.Run(commands[0])
}

// KTDockerMultiExec is like DockerMultiExec, except it keeps attempting the command after
// failure
func (fc *fakeClient) KTDockerMultiExec(node Node, commands []string) (string, error) {
	return fc.Run(commands[0])
}

// Scp is a wrapper for the scp command. Can be used to copy
// a file over to a remote machine.
func (fc *fakeClient) Scp(src string, dest string) error {
	return nil
}

// Close cleans up the resources used by sshClient object
func (fc *fakeClient) Close() {

}
