package helpers

import (
	db "../../db"
	state "../../state"
	util "../../util"
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"log"
	"sync"
)

func CopyToServers(clients []*util.SshClient, buildState *state.BuildState, src string, dst string) error {
	return CopyAllToServers(clients, buildState, src, dst)
}

func CopyAllToServers(clients []*util.SshClient, buildState *state.BuildState, srcDst ...string) error {
	if len(srcDst)%2 != 0 {
		return fmt.Errorf("Invalid number of variadic arguments, must be given an even number of them")
	}
	sem := semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()

	for i := range clients {
		for j := 0; j < len(srcDst)/2; j++ {
			sem.Acquire(ctx, 1)
			go func(i int, j int) {
				defer sem.Release(1)
				buildState.Defer(func() { clients[i].Run(fmt.Sprintf("rm -rf %s", srcDst[2*j+1])) })
				err := clients[i].Scp(srcDst[2*j], srcDst[2*j+1])
				if err != nil {
					log.Println(err)
					buildState.ReportError(err)
					return
				}
			}(i, j)
		}
	}

	sem.Acquire(ctx, conf.ThreadLimit)
	sem.Release(conf.ThreadLimit)
	return buildState.GetError()
}

func CopyToAllNodes(servers []db.Server, clients []*util.SshClient, buildState *state.BuildState, srcDst ...string) error {
	if len(srcDst)%2 != 0 {
		return fmt.Errorf("Invalid number of variadic arguments, must be given an even number of them")
	}
	sem := semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()
	wg := sync.WaitGroup{}
	for i, server := range servers {
		for j := 0; j < len(srcDst)/2; j++ {
			sem.Acquire(ctx, 1)
			rdy := make(chan bool, 1)
			wg.Add(1)
			intermediateDst := "/home/appo/" + srcDst[2*j]

			go func(i int, j int, server *db.Server, rdy chan bool) {
				defer sem.Release(1)
				defer wg.Done()
				ScpAndDeferRemoval(clients[i], buildState, srcDst[2*j], intermediateDst)
				rdy <- true
			}(i, j, &server, rdy)

			wg.Add(1)
			go func(i int, j int, server *db.Server, intermediateDst string, rdy chan bool) {
				defer wg.Done()
				<-rdy
				for k := range server.Ips {
					sem.Acquire(ctx, 1)
					wg.Add(1)
					go func(i int, j int, k int, intermediateDst string) {
						defer wg.Done()
						defer sem.Release(1)
						err := clients[i].DockerCp(k, intermediateDst, srcDst[2*j+1])
						if err != nil {
							log.Println(err)
							buildState.ReportError(err)
							return
						}
					}(i, j, k, intermediateDst)
				}
			}(i, j, &server, intermediateDst, rdy)
		}
	}

	wg.Wait()
	sem.Acquire(ctx, conf.ThreadLimit)
	sem.Release(conf.ThreadLimit)
	return buildState.GetError()
}

func CopyBytesToAllNodes(servers []db.Server, clients []*util.SshClient, buildState *state.BuildState, dataDst ...string) error {
	fmted := []string{}
	for i := 0; i < len(dataDst)/2; i++ {
		tmpFilename, err := util.GetUUIDString()
		if err != nil {
			log.Println(err)
			return err
		}
		err = buildState.Write(tmpFilename, dataDst[i*2])
		fmted = append(fmted, tmpFilename)
		fmted = append(fmted, dataDst[i*2+1])
	}
	return CopyToAllNodes(servers, clients, buildState, fmted...)
}

func SingleCp(client *util.SshClient, buildState *state.BuildState, localNodeId int, data []byte, dest string) error {
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

func CopyBytesToNodeFiles(client *util.SshClient, buildState *state.BuildState, transfers ...FileDest) error {
	sem := semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()

	for _, transfer := range transfers {
		sem.Acquire(ctx, 1)
		go func(transfer FileDest) {
			defer sem.Release(1)
			err := SingleCp(client, buildState, transfer.LocalNodeId, transfer.Data, transfer.Dest)
			if err != nil {
				log.Println(err)
				buildState.ReportError(err)
				return
			}
		}(transfer)
	}

	sem.Acquire(ctx, conf.ThreadLimit)
	sem.Release(conf.ThreadLimit)
	return buildState.GetError()
}

func CreateConfigs(servers []db.Server, clients []*util.SshClient, buildState *state.BuildState, dest string,
	fn func(serverNum int, localNodeNum int, absoluteNodeNum int) ([]byte, error)) error {

	sem := semaphore.NewWeighted(conf.ThreadLimit)
	ctx := context.TODO()
	node := 0
	for i, server := range servers {
		for j := range server.Ips {
			sem.Acquire(ctx, 1)
			go func(i int, j int, node int) {
				defer sem.Release(1)
				data, err := fn(i, j, node)
				if err != nil {
					log.Println(err)
					buildState.ReportError(err)
					return
				}
				err = SingleCp(clients[i], buildState, j, data, dest)
				if err != nil {
					log.Println(err)
					buildState.ReportError(err)
					return
				}

			}(i, j, node)
			node++
		}
	}

	sem.Acquire(ctx, conf.ThreadLimit)
	sem.Release(conf.ThreadLimit)
	return buildState.GetError()
}
