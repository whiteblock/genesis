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


func ScpAndDeferRemoval(client *util.SshClient, buildState *state.BuildState, src string, dst string){
	buildState.Defer(func() { client.Run(fmt.Sprintf("rm -rf %s", dst)) })
	err := clients[i].Scp(src, dst)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return
	}
}