package helpers

import (
	"../../ssh"
	"../../state"
	"../../testnet"
	"../../util"
	"fmt"
	"log"
	"sync"
)

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
			go func(client *ssh.Client, j int) {
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

func copyToAllNodes(tn *testnet.TestNet, useNew bool, sidecar bool, srcDst ...string) error {
	if len(srcDst)%2 != 0 {
		return fmt.Errorf("invalid number of variadic arguments, must be given an even number of them")
	}
	wg := sync.WaitGroup{}
	preOrderedNodes := tn.PreOrderNodes(useNew, sidecar)

	for sid, nodes := range preOrderedNodes {
		for j := 0; j < len(srcDst)/2; j++ {
			rdy := make(chan bool, 1)
			wg.Add(1)
			intermediateDst := "/home/appo/" + srcDst[2*j]

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
							tn.BuildState.ReportError(err)
							return
						}
					}(nodes[i], j, intermediateDst)
				}
			}(nodes, j, intermediateDst, rdy)
		}
	}

	wg.Wait()
	return tn.BuildState.GetError()
}

// CopyToAllNodes copies files writen with BuildState's write function over to all of the nodes.
// Can handle multiple files, in pairs of src and dst
func CopyToAllNodes(tn *testnet.TestNet, srcDst ...string) error {
	return copyToAllNodes(tn, false, false, srcDst...)
}

// CopyToAllNewNodes copies files writen with BuildState's write function over to all of the newly built nodes.
// Can handle multiple files, in pairs of src and dst
func CopyToAllNewNodes(tn *testnet.TestNet, srcDst ...string) error {
	return copyToAllNodes(tn, true, false, srcDst...)
}

// CopyToAllNodesSC is CopyToAllNodes for side cars
func CopyToAllNodesSC(tn *testnet.TestNet, srcDst ...string) error {
	return copyToAllNodes(tn, false, true, srcDst...)
}

// CopyToAllNewNodesSC is CopyToAllNewNodes for side cars
func CopyToAllNewNodesSC(tn *testnet.TestNet, srcDst ...string) error {
	return copyToAllNodes(tn, true, true, srcDst...)
}

func copyBytesToAllNodes(tn *testnet.TestNet, useNew bool, sidecar bool, dataDst ...string) error {
	fmted := []string{}
	for i := 0; i < len(dataDst)/2; i++ {
		tmpFilename, err := util.GetUUIDString()
		if err != nil {
			return util.LogError(err)
		}
		err = tn.BuildState.Write(tmpFilename, dataDst[i*2])
		fmted = append(fmted, tmpFilename)
		fmted = append(fmted, dataDst[i*2+1])
	}
	return copyToAllNodes(tn, useNew, sidecar, fmted...)
}

// CopyBytesToAllNodes functions similiarly to CopyToAllNodes, except it operates on data and dst pairs instead of
// src and dest pairs, so you can just pass data directly to all of the nodes without having to call buildState.Write first.
func CopyBytesToAllNodes(tn *testnet.TestNet, dataDst ...string) error {
	return copyBytesToAllNodes(tn, false, false, dataDst...)
}

// CopyBytesToAllNewNodes is CopyBytesToAllNodes but only operates on newly built nodes
func CopyBytesToAllNewNodes(tn *testnet.TestNet, dataDst ...string) error {
	return copyBytesToAllNodes(tn, true, false, dataDst...)
}

// CopyBytesToAllNodesSC is CopyBytesToAllNodes but only operates on sidecar nodes
func CopyBytesToAllNodesSC(tn *testnet.TestNet, dataDst ...string) error {
	return copyBytesToAllNodes(tn, false, true, dataDst...)
}

// CopyBytesToAllNewNodesSC is CopyBytesToAllNewNodes but only operates on sidecar nodes
func CopyBytesToAllNewNodesSC(tn *testnet.TestNet, dataDst ...string) error {
	return copyBytesToAllNodes(tn, true, true, dataDst...)
}

// SingleCp copies over data to the given dest on node localNodeID.
func SingleCp(client *ssh.Client, buildState *state.BuildState, node ssh.Node, data []byte, dest string) error {
	tmpFilename, err := util.GetUUIDString()
	if err != nil {
		log.Println(err)
		return err
	}

	err = buildState.Write(tmpFilename, string(data))
	if err != nil {
		log.Println(err)
		return err
	}
	intermediateDst := "/home/appo/" + tmpFilename
	buildState.Defer(func() { client.Run("rm " + intermediateDst) })
	err = client.Scp(tmpFilename, intermediateDst)
	if err != nil {
		return util.LogError(err)
	}

	return client.DockerCp(node, intermediateDst, dest)
}

/*// FileDest represents a transfer of data
type FileDest struct {
	// Data is the data to be transfered
	Data []byte
	// Dest is the destination for the data
	Dest string
	// LocalNodeID is the local node number of the node to which the data will be transfered
	LocalNodeID int
}

//CopyBytesToNodeFiles executes the file transfers represented by the given file dests.
func CopyBytesToNodeFiles(client *ssh.Client, buildState *state.BuildState, transfers ...FileDest) error {
	wg := sync.WaitGroup{}

	for _, transfer := range transfers {
		wg.Add(1)
		go func(transfer FileDest) {
			defer wg.Done()
			err := SingleCp(client, buildState, transfer.LocalNodeID, transfer.Data, transfer.Dest)
			if err != nil {
				log.Println(err)
				buildState.ReportError(err)
				return
			}
		}(transfer)
	}
	wg.Wait()
	return buildState.GetError()
}*/

/*
	fn func(serverid int, localNodeNum int, absoluteNodeNum int) ([]byte, error)
*/
func createConfigs(tn *testnet.TestNet, dest string, useNew bool, sidecar bool, fn func(ssh.Node) ([]byte, error)) error {
	nodes := tn.GetSSHNodes(useNew, sidecar)
	wg := sync.WaitGroup{}
	for _, node := range nodes {
		wg.Add(1)
		go func(client *ssh.Client, node ssh.Node) {
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

// CreateConfigs allows for individual generation of configuration files with error propogation.
// For each node, fn will be called, with (Server ID, local node number, absolute node number), and it will expect
// to have the configuration file returned or error.
func CreateConfigs(tn *testnet.TestNet, dest string, fn func(ssh.Node) ([]byte, error)) error {
	return createConfigs(tn, dest, false, false, fn)
}

// CreateConfigsNewNodes is CreateConfigs but it only operates on new nodes
func CreateConfigsNewNodes(tn *testnet.TestNet, dest string, fn func(ssh.Node) ([]byte, error)) error {
	return createConfigs(tn, dest, true, false, fn)
}

// CreateConfigsSC is CreateConfigs but it only operates on side cars
func CreateConfigsSC(tn *testnet.TestNet, dest string, fn func(ssh.Node) ([]byte, error)) error {
	return createConfigs(tn, dest, false, true, fn)
}

// CreateConfigsNewNodesSC is CreateConfigsNewNodes but it only operates on side cars
func CreateConfigsNewNodesSC(tn *testnet.TestNet, dest string, fn func(ssh.Node) ([]byte, error)) error {
	return createConfigs(tn, dest, true, true, fn)
}
