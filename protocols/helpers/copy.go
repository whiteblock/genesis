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

package helpers

import (
	"fmt"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/state"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
	"sync"
)

type settings struct {
	useNew      bool
	sidecar     int
	reportError bool
}

// CopyAllToServers copies all of the src files to all of the servers within the given testnet.
// This can handle multiple pairs in form of ...,source,destination,source2,destination2
func CopyAllToServers(tn *testnet.TestNet, srcDst ...string) error {
	if len(srcDst)%2 != 0 {
		return fmt.Errorf("invalid number of variadic arguments, must be given an even number of them")
	}
	wg := sync.WaitGroup{}
	for _, client := range tn.Clients {
		for j := 0; j < len(srcDst)/2; j++ {
			wg.Add(1)
			go func(client ssh.Client, j int) {
				defer wg.Done()
				tn.BuildState.Defer(func() { client.Run(fmt.Sprintf("rm -rf %s", srcDst[2*j+1])) })
				err := client.Scp(srcDst[2*j], srcDst[2*j+1])
				if err != nil {
					tn.BuildState.ReportError(err)
					return
				}
			}(client, j)
		}
	}
	wg.Wait()
	return tn.BuildState.GetError()
}

func copyToAllNodes(tn *testnet.TestNet, s settings, srcDst ...string) error {
	if len(srcDst)%2 != 0 {
		return fmt.Errorf("invalid number of variadic arguments, must be given an even number of them")
	}
	wg := sync.WaitGroup{}
	preOrderedNodes := tn.PreOrderNodes(s.useNew, s.sidecar != -1, s.sidecar)

	for sid, nodes := range preOrderedNodes {
		for j := 0; j < len(srcDst)/2; j++ {
			rdy := make(chan bool, 1)
			wg.Add(1)
			intermediateDst := "/tmp/" + srcDst[2*j]

			go func(sid int, j int, rdy chan bool) {
				defer wg.Done()
				ScpAndDeferRemoval(tn.Clients[sid], tn.BuildState, srcDst[2*j], intermediateDst)
				rdy <- true
			}(sid, j, rdy)

			wg.Add(1)
			go func(nodes []ssh.Node, j int, intermediateDst string, rdy chan bool) {
				defer wg.Done()
				<-rdy
				for i := range nodes {
					wg.Add(1)
					go func(node ssh.Node, j int, intermediateDst string) {
						defer wg.Done()
						err := tn.Clients[node.GetServerID()].DockerCp(node, intermediateDst, srcDst[2*j+1])
						if err != nil {
							if s.reportError {
								tn.BuildState.ReportError(err)
							} else {
								tn.BuildState.Set("error", err)
							}

							return
						}
					}(nodes[i], j, intermediateDst)
				}
			}(nodes, j, intermediateDst, rdy)
		}
	}

	wg.Wait()
	return getError(tn, s)
}

// CopyToAllNodes copies files written with BuildState's write function over to all of the nodes.
// Can handle multiple files, in pairs of src and dst
func CopyToAllNodes(tn *testnet.TestNet, srcDst ...string) error {
	return copyToAllNodes(tn, settings{useNew: false, sidecar: -1, reportError: true}, srcDst...)
}

// CopyToAllNewNodes copies files written with BuildState's write function over to all of the newly built nodes.
// Can handle multiple files, in pairs of src and dst
func CopyToAllNewNodes(tn *testnet.TestNet, srcDst ...string) error {
	return copyToAllNodes(tn, settings{useNew: true, sidecar: -1, reportError: true}, srcDst...)
}

// CopyToAllNodesDR copies files written with BuildState's write function over to all of the nodes.
// Can handle multiple files, in pairs of src and dst. DR means it doesn't report the error to build state.
func CopyToAllNodesDR(tn *testnet.TestNet, srcDst ...string) error {
	return copyToAllNodes(tn, settings{useNew: false, sidecar: -1, reportError: false}, srcDst...)
}

// CopyToAllNewNodesDR copies files written with BuildState's write function over to all of the newly built nodes.
// Can handle multiple files, in pairs of src and dst. DR means it doesn't report the error to build state.
func CopyToAllNewNodesDR(tn *testnet.TestNet, srcDst ...string) error {
	return copyToAllNodes(tn, settings{useNew: true, sidecar: -1, reportError: false}, srcDst...)
}

// CopyToAllNodesSC is CopyToAllNodes for side cars
func CopyToAllNodesSC(ad *testnet.Adjunct, srcDst ...string) error {
	return copyToAllNodes(ad.Main, settings{useNew: false, sidecar: ad.Index, reportError: true}, srcDst...)
}

// CopyToAllNewNodesSC is CopyToAllNewNodes for side cars
func CopyToAllNewNodesSC(ad *testnet.Adjunct, srcDst ...string) error {
	return copyToAllNodes(ad.Main, settings{useNew: true, sidecar: ad.Index, reportError: true}, srcDst...)
}

func copyBytesToAllNodes(tn *testnet.TestNet, s settings, dataDst ...string) error {
	fmted := []string{}
	for i := 0; i < len(dataDst)/2; i++ {
		tmpFilename, err := util.GetUUIDString()
		if err != nil {
			return util.LogError(err)
		}

		err = tn.BuildState.Write(tmpFilename, dataDst[i*2])
		if err != nil {
			return util.LogError(err)
		}
		fmted = append(fmted, tmpFilename, dataDst[i*2+1])
	}
	return copyToAllNodes(tn, s, fmted...)
}

// CopyBytesToAllNodes functions similarly to CopyToAllNodes, except it operates on data and dst pairs instead of
// src and dest pairs, so you can just pass data directly to all of the nodes without having to call buildState.Write first.
func CopyBytesToAllNodes(tn *testnet.TestNet, dataDst ...string) error {
	return copyBytesToAllNodes(tn, settings{useNew: false, sidecar: -1, reportError: true}, dataDst...)
}

// CopyBytesToAllNewNodes is CopyBytesToAllNodes but only operates on newly built nodes
func CopyBytesToAllNewNodes(tn *testnet.TestNet, dataDst ...string) error {
	return copyBytesToAllNodes(tn, settings{useNew: true, sidecar: -1, reportError: true}, dataDst...)
}

// CopyBytesToAllNodesSC is CopyBytesToAllNodes but only operates on sidecar nodes
func CopyBytesToAllNodesSC(ad *testnet.Adjunct, dataDst ...string) error {
	return copyBytesToAllNodes(ad.Main, settings{useNew: false, sidecar: ad.Index, reportError: true}, dataDst...)
}

// CopyBytesToAllNewNodesSC is CopyBytesToAllNewNodes but only operates on sidecar nodes
func CopyBytesToAllNewNodesSC(ad *testnet.Adjunct, dataDst ...string) error {
	return copyBytesToAllNodes(ad.Main, settings{useNew: true, sidecar: ad.Index, reportError: true}, dataDst...)
}

// SingleCp copies over data to the given dest on node localNodeID.
func SingleCp(client ssh.Client, buildState *state.BuildState, node ssh.Node, data []byte, dest string) error {
	tmpFilename, err := util.GetUUIDString()
	if err != nil {
		return util.LogError(err)
	}

	err = buildState.Write(tmpFilename, string(data))
	if err != nil {
		return util.LogError(err)
	}

	intermediateDst := "/tmp/" + tmpFilename
	buildState.Defer(func() { client.Run("rm " + intermediateDst) })
	err = client.Scp(tmpFilename, intermediateDst)
	if err != nil {
		return util.LogError(err)
	}

	return client.DockerCp(node, intermediateDst, dest)
}

/*
	fn func(node ssh.Node) ([]byte, error)
*/
func createConfigs(tn *testnet.TestNet, dest string, s settings, fn func(ssh.Node) ([]byte, error)) error {
	nodes := tn.GetSSHNodes(s.useNew, s.sidecar != -1, s.sidecar)
	wg := sync.WaitGroup{}
	for _, node := range nodes {
		wg.Add(1)
		go func(client ssh.Client, node ssh.Node) {
			defer wg.Done()
			data, err := fn(node)
			if err != nil {
				tn.BuildState.ReportError(err)
				return
			}
			if data == nil {
				return //skip if nil
			}
			err = SingleCp(client, tn.BuildState, node, data, dest)
			if err != nil {
				tn.BuildState.ReportError(err)
				return
			}

		}(tn.Clients[node.GetServerID()], node)
	}

	wg.Wait()
	return tn.BuildState.GetError()
}

// CreateConfigs allows for individual generation of configuration files with error propagation.
// For each node, fn will be called, with (Server ID, local node number, absolute node number), and it will expect
// to have the configuration file returned or error.
func CreateConfigs(tn *testnet.TestNet, dest string, fn func(ssh.Node) ([]byte, error)) error {
	return createConfigs(tn, dest, settings{useNew: false, sidecar: -1, reportError: true}, fn)
}

// CreateConfigsNewNodes is CreateConfigs but it only operates on new nodes
func CreateConfigsNewNodes(tn *testnet.TestNet, dest string, fn func(ssh.Node) ([]byte, error)) error {
	return createConfigs(tn, dest, settings{useNew: true, sidecar: -1, reportError: true}, fn)
}

// CreateConfigsSC is CreateConfigs but it only operates on side cars
func CreateConfigsSC(ad *testnet.Adjunct, dest string, fn func(ssh.Node) ([]byte, error)) error {
	return createConfigs(ad.Main, dest, settings{useNew: false, sidecar: ad.Index, reportError: true}, fn)
}

// CreateConfigsNewNodesSC is CreateConfigsNewNodes but it only operates on side cars
func CreateConfigsNewNodesSC(ad *testnet.Adjunct, dest string, fn func(ssh.Node) ([]byte, error)) error {
	return createConfigs(ad.Main, dest, settings{useNew: true, sidecar: ad.Index, reportError: true}, fn)
}
