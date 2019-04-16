package helpers

import (
	state "../../state"
	util "../../util"
	"fmt"
	"log"
)

func ScpAndDeferRemoval(client *util.SshClient, buildState *state.BuildState, src string, dst string) {
	buildState.Defer(func() { client.Run(fmt.Sprintf("rm -rf %s", dst)) })
	err := client.Scp(src, dst)
	if err != nil {
		log.Println(err)
		buildState.ReportError(err)
		return
	}
}
