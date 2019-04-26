package helpers

import (
	"../../db"
	"../../ssh"
	"../../state"
	"../../testnet"
	"../../util"
	"fmt"
	"log"
	"sync"
)

func CopyToServers(tn *testnet.TestNet, src string, dst string) error {
	return CopyAllToServers(tn, src, dst)
}

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
					log.Println(err)
					tn.BuildState.ReportError(err)
					return
				}
			}(client, j)
		}
	}
	wg.Wait()
	return tn.BuildState.GetError()
}

func copyToAllNodes(tn *testnet.TestNet, useNew bool, srcDst ...string) error {
	if len(srcDst)%2 != 0 {
		return fmt.Errorf("invalid number of variadic arguments, must be given an even number of them")
	}
	wg := sync.WaitGroup{}

	preOrderedNodes := tn.PreOrderNodes()
	if useNew {
		preOrderedNodes = tn.PreOrderNewNodes()
	}
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
			go func(nodes []db.Node, j int, intermediateDst string, rdy chan bool) {
				defer wg.Done()
				<-rdy
				for i := range nodes {
					wg.Add(1)
					go func(node *db.Node, j int, intermediateDst string) {
						defer wg.Done()
						err := tn.Clients[node.Server].DockerCp(node.LocalID, intermediateDst, srcDst[2*j+1])
						if err != nil {
							log.Println(err)
							tn.BuildState.ReportError(err)
							return
						}
					}(&nodes[i], j, intermediateDst)
				}
			}(nodes, j, intermediateDst, rdy)
		}
	}

	wg.Wait()
	return tn.BuildState.GetError()
}

func CopyToAllNodes(tn *testnet.TestNet, srcDst ...string) error {
	return copyToAllNodes(tn, false, srcDst...)
}

func CopyToAllNewNodes(tn *testnet.TestNet, srcDst ...string) error {
	return copyToAllNodes(tn, true, srcDst...)
}

func copyBytesToAllNodes(tn *testnet.TestNet, useNew bool, dataDst ...string) error {
	fmted := []string{}
	for i := 0; i < len(dataDst)/2; i++ {
		tmpFilename, err := util.GetUUIDString()
		if err != nil {
			log.Println(err)
			return err
		}
		err = tn.BuildState.Write(tmpFilename, dataDst[i*2])
		fmted = append(fmted, tmpFilename)
		fmted = append(fmted, dataDst[i*2+1])
	}
	return copyToAllNodes(tn, useNew, fmted...)
}

func CopyBytesToAllNodes(tn *testnet.TestNet, dataDst ...string) error {
	return copyBytesToAllNodes(tn, false, dataDst...)
}

func CopyBytesToAllNewNodes(tn *testnet.TestNet, dataDst ...string) error {
	return copyBytesToAllNodes(tn, true, dataDst...)
}

func SingleCp(client *ssh.Client, buildState *state.BuildState, localNodeId int, data []byte, dest string) error {
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
		log.Println(err)
		return err
	}

	return client.DockerCp(localNodeId, intermediateDst, dest)
}

type FileDest struct {
	Data        []byte
	Dest        string
	LocalNodeId int
}

func CopyBytesToNodeFiles(client *ssh.Client, buildState *state.BuildState, transfers ...FileDest) error {
	wg := sync.WaitGroup{}

	for _, transfer := range transfers {
		wg.Add(1)
		go func(transfer FileDest) {
			defer wg.Done()
			err := SingleCp(client, buildState, transfer.LocalNodeId, transfer.Data, transfer.Dest)
			if err != nil {
				log.Println(err)
				buildState.ReportError(err)
				return
			}
		}(transfer)
	}
	wg.Wait()
	return buildState.GetError()
}

/*
	fn func(serverid int, localNodeNum int, absoluteNodeNum int) ([]byte, error)
*/
func createConfigs(tn *testnet.TestNet, dest string, useNew bool, fn func(int, int, int) ([]byte, error)) error {
	nodes := tn.Nodes
	if useNew {
		nodes = tn.NewlyBuiltNodes
	}
	wg := sync.WaitGroup{}
	for _, node := range nodes {
		wg.Add(1)
		go func(client *ssh.Client, serverID int, localID int, absNum int) {
			defer wg.Done()
			data, err := fn(serverID, localID, absNum)
			if err != nil {
				log.Println(err)
				tn.BuildState.ReportError(err)
				return
			}
			if data == nil {
				return //skip if nil
			}
			err = SingleCp(client, tn.BuildState, localID, data, dest)
			if err != nil {
				log.Println(err)
				tn.BuildState.ReportError(err)
				return
			}

		}(tn.Clients[node.Server], node.Server, node.LocalID, node.AbsoluteNum)
	}

	wg.Wait()
	return tn.BuildState.GetError()
}

func CreateConfigsNewNodes(tn *testnet.TestNet, dest string, fn func(int, int, int) ([]byte, error)) error {
	return createConfigs(tn, dest, true, fn)
}

func CreateConfigs(tn *testnet.TestNet, dest string, fn func(int, int, int) ([]byte, error)) error {
	return createConfigs(tn, dest, false, fn)
}
